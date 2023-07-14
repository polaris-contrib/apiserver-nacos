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

package model

import (
	"strings"

	"github.com/polarismesh/polaris/common/model"
	"github.com/polarismesh/polaris/common/utils"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Service struct {
	CacheMillis              uint64     `json:"cacheMillis"`
	Hosts                    []Instance `json:"hosts"`
	Checksum                 string     `json:"checksum"`
	LastRefTime              uint64     `json:"lastRefTime"`
	Clusters                 string     `json:"clusters"`
	Name                     string     `json:"name"`
	GroupName                string     `json:"groupName"`
	Valid                    bool       `json:"valid"`
	AllIPs                   bool       `json:"allIPs"`
	ReachProtectionThreshold bool       `json:"reachProtectionThreshold"`
}

type SimpleServiceInfo struct {
	Namespace string
	Name      string `json:"name"`
	GroupName string `json:"groupName"`
}

type ServiceInfo struct {
	Namespace                string      `json:"-"`
	Name                     string      `json:"name"`
	GroupName                string      `json:"groupName"`
	Clusters                 string      `json:"clusters"`
	Hosts                    []*Instance `json:"hosts"`
	Checksum                 string      `json:"checksum"`
	CacheMillis              int64       `json:"cacheMillis"`
	LastRefTime              int64       `json:"lastRefTime"`
	ReachProtectionThreshold bool        `json:"reachProtectionThreshold"`
}

// NewEmptyServiceInfo .
func NewEmptyServiceInfo(name, group string) *ServiceInfo {
	return &ServiceInfo{
		Name:      name,
		GroupName: group,
		Hosts:     []*Instance{},
	}
}

type Instance struct {
	Id          string            `json:"instanceId"`
	IP          string            `json:"ip"`
	Port        int32             `json:"port"`
	Weight      float64           `json:"weight"`
	Healthy     bool              `json:"healthy"`
	Enabled     bool              `json:"enabled"`
	Ephemeral   bool              `json:"ephemeral"`
	ClusterName string            `json:"clusterName"`
	ServiceName string            `json:"serviceName"`
	Metadata    map[string]string `json:"metadata"`
}

func (i *Instance) FromSpecInstance(specIns *model.Instance) {
	i.Id = specIns.ID()
	i.IP = specIns.Host()
	i.Port = int32(specIns.Port())
	i.Weight = float64(specIns.Weight())
	i.Ephemeral = true
	i.Healthy = specIns.Healthy()
	i.Enabled = !specIns.Isolate()

	copyMeta := make(map[string]string)
	for k, v := range specIns.Metadata() {
		copyMeta[k] = v
	}
	i.Metadata = copyMeta
	i.ClusterName = i.Metadata[InternalNacosCluster]
	i.ServiceName = i.Metadata[InternalNacosServiceName]
}

func (i *Instance) DeepClone() *Instance {
	copyMeta := make(map[string]string, len(i.Metadata))
	for k, v := range i.Metadata {
		copyMeta[k] = v
	}

	return &Instance{
		Id:          i.Id,
		IP:          i.IP,
		Port:        i.Port,
		Weight:      i.Weight,
		Healthy:     i.Healthy,
		Enabled:     i.Enabled,
		Ephemeral:   i.Ephemeral,
		ClusterName: i.ClusterName,
		ServiceName: i.ServiceName,
		Metadata:    copyMeta,
	}
}

func (i *Instance) ToSpecInstance() *apiservice.Instance {
	ret := &apiservice.Instance{
		Id:                wrapperspb.String(i.Id),
		Service:           wrapperspb.String(i.ServiceName),
		Host:              wrapperspb.String(i.IP),
		Port:              wrapperspb.UInt32(uint32(i.Port)),
		Weight:            wrapperspb.UInt32(uint32(i.Weight)),
		EnableHealthCheck: wrapperspb.Bool(true),
		HealthCheck: &apiservice.HealthCheck{
			Type: apiservice.HealthCheck_HEARTBEAT,
			Heartbeat: &apiservice.HeartbeatHealthCheck{
				Ttl: &wrapperspb.UInt32Value{
					Value: 5,
				},
			},
		},
		Healthy:  wrapperspb.Bool(i.Healthy),
		Isolate:  wrapperspb.Bool(!i.Enabled),
		Metadata: i.Metadata,
	}
	if len(ret.GetId().GetValue()) == 0 {
		ret.Id = nil
	}
	if len(ret.Metadata) == 0 {
		ret.Metadata = make(map[string]string)
	}
	return ret
}

type ClientBeat struct {
	Namespace   string            `json:"namespace"`
	ServiceName string            `json:"serviceName"`
	Cluster     string            `json:"cluster"`
	Ip          string            `json:"ip"`
	Port        int               `json:"port"`
	Weight      float64           `json:"weight"`
	Ephemeral   bool              `json:"ephemeral"`
	Metadata    map[string]string `json:"metadata"`
}

func (c *ClientBeat) ToSpecInstance() (*apiservice.Instance, error) {
	return nil, nil
}

func PrepareSpecInstance(namespace, service string, ins *Instance) *apiservice.Instance {
	pSvc := ReplaceNacosService(service)

	specIns := ins.ToSpecInstance()
	specIns.Service = utils.NewStringValue(pSvc)
	specIns.Namespace = utils.NewStringValue(namespace)

	specIns.Metadata[InternalNacosCluster] = ins.ClusterName
	specIns.Metadata[InternalNacosServiceName] = ins.ServiceName

	return specIns
}

func ReplaceNacosService(service string) string {
	// nacos 的服务名和分组名默认是通过 @@ 进行连接的，这里可能需要按照北极星服务名支持的方式，replace 替换下 @@ 连接符号为 __
	return strings.ReplaceAll(service, DefaultNacosGroupConnectStr, ReplaceNacosGroupConnectStr)
}

func BuildServiceName(svcName, groupName string) string {
	if groupName == DefaultServiceGroup {
		return svcName
	}
	return groupName + ReplaceNacosGroupConnectStr + svcName
}
