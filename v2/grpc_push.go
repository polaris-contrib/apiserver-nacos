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
	"sync"

	"github.com/polarismesh/polaris/common/log"
	commontime "github.com/polarismesh/polaris/common/time"
	"go.uber.org/zap"

	"github.com/polaris-contrib/apiserver-nacos/core"
)

type Sender func(sub core.Subscriber, data *core.PushData) error

type GrpcPushCenter struct {
	*core.BasePushCenter
	sender Sender
}

func NewGrpcPushCenter(store *core.NacosDataStorage, sender Sender) (core.PushCenter, error) {
	return &GrpcPushCenter{
		BasePushCenter: core.NewBasePushCenter(store),
		sender:         sender,
	}, nil
}

func (p *GrpcPushCenter) AddSubscriber(s core.Subscriber) {
	notifier := &GRPCNotifier{
		subscriber: s,
		sender:     p.sender,
	}
	if ok := p.BasePushCenter.AddSubscriber(s, notifier); !ok {
		_ = notifier.Close()
		return
	}
	log.Info("[NACOS-V2][PushCenter] add subscriber", zap.String("type", string(p.Type())),
		zap.String("conn-id", s.ConnID))
	client := p.BasePushCenter.GetSubscriber(s)
	if client != nil {
		client.RefreshLastTime()
	}
}

func (p *GrpcPushCenter) RemoveSubscriber(s core.Subscriber) {
	log.Info("[NACOS-V2][PushCenter] remove subscriber", zap.String("type", string(p.Type())),
		zap.String("conn-id", s.ConnID))
	p.BasePushCenter.RemoveSubscriber(s)
}

func (p *GrpcPushCenter) EnablePush(s core.Subscriber) bool {
	return p.Type() == s.Type
}

func (p *GrpcPushCenter) Type() core.PushType {
	return core.GRPCPush
}

type GRPCNotifier struct {
	lock        sync.Mutex
	subscriber  core.Subscriber
	sender      Sender
	lastRefTime int64
}

func (c *GRPCNotifier) Notify(d *core.PushData) error {
	return c.sender(c.subscriber, d)
}

func (c *GRPCNotifier) IsZombie() bool {
	return commontime.CurrentMillisecond()-c.lastRefTime > 10*1000
}

func (c *GRPCNotifier) Close() error {
	return nil
}
