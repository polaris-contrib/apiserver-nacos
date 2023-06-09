/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v2

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/polarismesh/polaris/cache"
	"github.com/polarismesh/polaris/common/eventhub"
	"github.com/polarismesh/polaris/common/model"
	"github.com/polarismesh/polaris/common/utils"
	"github.com/polarismesh/polaris/service"
	"github.com/polarismesh/polaris/store"
	"github.com/polarismesh/specification/source/go/api/v1/service_manage"
	"go.uber.org/zap"

	nacosmodel "github.com/pole-group/polaris-apiserver-nacos/model"
)

type Checker struct {
	discoverSvr service.DiscoverServer

	cacheMgr      *cache.CacheManager
	connectionMgr *ConnectionManager
	clientMgr     *ConnectionClientManager

	lock          sync.RWMutex
	selfInstances map[string]*service_manage.Instance
	instances     map[string]*service_manage.Instance

	leader int32

	cancel context.CancelFunc
}

const (
	eventhubSubscriberName = "nacos-v2-checker"
)

// newChecker 创建 nacos 长连接和实例信息绑定关系的健康检查，如果长连接不存在，则该连接上绑定的实例信息将失效
func newChecker(discoverSvr service.DiscoverServer, connectionMgr *ConnectionManager,
	clientMgr *ConnectionClientManager) *Checker {

	ctx, cancel := context.WithCancel(context.Background())
	checker := &Checker{
		discoverSvr:   discoverSvr,
		cacheMgr:      discoverSvr.Cache(),
		connectionMgr: connectionMgr,
		clientMgr:     clientMgr,
		selfInstances: make(map[string]*service_manage.Instance),
		instances:     make(map[string]*service_manage.Instance),
		cancel:        cancel,
	}
	go checker.runCheck(ctx)
	checker.cacheMgr.AddListener(cache.CacheNameInstance, []cache.Listener{checker})
	// 注册 leader 变化事件
	subscriber := &CheckerLeaderSubscriber{checker: checker}
	_ = eventhub.Subscribe(eventhub.LeaderChangeEventTopic, eventhubSubscriberName, subscriber)
	return checker
}

func (c *Checker) isLeader() bool {
	return atomic.LoadInt32(&c.leader) == 1
}

// OnCreated callback when cache value created
func (c *Checker) OnCreated(value interface{}) {
	ins, _ := value.(*model.Instance)

	c.lock.Lock()
	defer c.lock.Unlock()

	if isSelfServiceInstance(ins.Proto) {
		c.selfInstances[ins.Proto.GetId().GetValue()] = ins.Proto
		return
	}

	if _, ok := ins.Proto.GetMetadata()[nacosmodel.InternalNacosClientConnectionID]; ok {
		c.instances[ins.Proto.GetId().GetValue()] = ins.Proto
	}
}

// OnUpdated callback when cache value updated
func (c *Checker) OnUpdated(value interface{}) {
	ins, _ := value.(*model.Instance)
	_, ok := ins.Proto.GetMetadata()[nacosmodel.InternalNacosClientConnectionID]
	c.lock.Lock()
	defer c.lock.Unlock()

	if ok {
		c.instances[ins.Proto.GetId().GetValue()] = ins.Proto
	}
}

// OnDeleted callback when cache value deleted
func (c *Checker) OnDeleted(value interface{}) {
	ins, _ := value.(*model.Instance)
	_, ok := ins.Proto.GetMetadata()[nacosmodel.InternalNacosClientConnectionID]
	c.lock.Lock()
	defer c.lock.Unlock()

	if ok {
		delete(c.instances, ins.Proto.GetId().GetValue())
	}
}

// OnBatchCreated callback when cache value created
func (c *Checker) OnBatchCreated(value interface{}) {}

// OnBatchUpdated callback when cache value updated
func (c *Checker) OnBatchUpdated(value interface{}) {}

// OnBatchDeleted callback when cache value deleted
func (c *Checker) OnBatchDeleted(value interface{}) {}

func (c *Checker) runCheck(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.realCheck()
		}
	}
}

// 根据元数据的 Metadata ConnID 进行判断当前的长连接是否存在，如果对应长连接不存在，则反注册改实例信息数据。
// BUT: 一个实例 T1 时刻对应长连接为 Conn-1，T2 时刻对应的长连接为 Conn-2，但是在 T1 ～ T2 之间的某个时刻检测发现长连接不存在
// 此时发起一个反注册请求，该请求在 T3 时刻发起，是否会影响 T2 时刻注册上来的实例？
func (c *Checker) realCheck() {
	copyMap := make(map[string]*service_manage.Instance, len(c.instances))

	// 减少锁的耗时
	c.lock.RLock()
	for k, v := range c.instances {
		copyMap[k] = v
	}
	c.lock.RUnlock()

	turnUnhealth := map[string]struct{}{}
	turnHealth := map[string]struct{}{}

	for instanceID, instance := range copyMap {
		connID := instance.Metadata[nacosmodel.InternalNacosClientConnectionID]
		if !strings.HasSuffix(connID, utils.LocalHost) {
			if c.isLeader() {
				// 看下所属的 server 是否健康
				found := false
				for i := range c.selfInstances {
					selfIns := c.selfInstances[i]
					if strings.HasSuffix(connID, selfIns.GetHost().GetValue()) && selfIns.GetHealthy().GetValue() {
						found = true
						break
					}
				}
				if !found {
					// connID 的负责 server 不存在，直接变为不健康
					turnUnhealth[instanceID] = struct{}{}
				}
				continue
			}
		}
		_, exist := c.connectionMgr.GetClient(connID)
		if !exist {
			// 如果实例对应的连接ID不存在，设置为不健康
			turnUnhealth[instanceID] = struct{}{}
			continue
		}
		if !instance.GetHealthy().GetValue() && exist {
			turnHealth[instanceID] = struct{}{}
		}
	}

	ids := make([]interface{}, 0, len(turnUnhealth))
	for id := range turnUnhealth {
		ids = append(ids, id)
	}
	if err := c.discoverSvr.Cache().GetStore().
		BatchSetInstanceHealthStatus(ids, model.StatusBoolToInt(false), utils.NewUUID()); err != nil {
		nacoslog.Error("[NACOS-V2][Checker] batch set instance health_status to unhealthy",
			zap.Any("instance-ids", ids), zap.Error(err))
	}

	ids = make([]interface{}, 0, len(turnUnhealth))
	for id := range turnHealth {
		ids = append(ids, id)
	}
	if err := c.discoverSvr.Cache().GetStore().
		BatchSetInstanceHealthStatus(ids, model.StatusBoolToInt(false), utils.NewUUID()); err != nil {
		nacoslog.Error("[NACOS-V2][Checker] batch set instance health_status to healty",
			zap.Any("instance-ids", ids), zap.Error(err))
	}
}

// CheckerLeaderSubscriber
type CheckerLeaderSubscriber struct {
	checker *Checker
}

// PreProcess do preprocess logic for event
func (c *CheckerLeaderSubscriber) PreProcess(ctx context.Context, value any) any {
	return value
}

// OnEvent event trigger
func (c *CheckerLeaderSubscriber) OnEvent(ctx context.Context, i interface{}) error {
	electionEvent, ok := i.(store.LeaderChangeEvent)
	if !ok {
		return nil
	}
	if electionEvent.Key != store.ElectionKeySelfServiceChecker {
		return nil
	}
	if electionEvent.Leader {
		atomic.StoreInt32(&c.checker.leader, 1)
	} else {
		atomic.StoreInt32(&c.checker.leader, 0)
	}
	return nil
}

func isSelfServiceInstance(instance *service_manage.Instance) bool {
	metadata := instance.GetMetadata()
	if svcName, ok := metadata[model.MetaKeyPolarisService]; ok {
		return svcName == "service-nacos"
	}
	return false
}
