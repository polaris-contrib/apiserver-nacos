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

package nacos_grpc_service

import (
	fmt "fmt"
	"strconv"
	"time"

	"github.com/pole-group/polaris-apiserver-nacos/model"
)

// NamingRequest
type NamingRequest struct {
	*Request
	Namespace   string `json:"namespace"`
	ServiceName string `json:"serviceName"`
	GroupName   string `json:"groupName"`
	Module      string `json:"module"`
}

// NewNamingRequest
func NewNamingRequest(namespace, serviceName, groupName string) *NamingRequest {
	request := Request{
		Headers: make(map[string]string, 8),
	}
	return &NamingRequest{
		Request:     &request,
		Namespace:   namespace,
		ServiceName: serviceName,
		GroupName:   groupName,
		Module:      "naming",
	}
}

func (r *NamingRequest) GetStringToSign() string {
	data := strconv.FormatInt(time.Now().Unix()*1000, 10)
	if r.ServiceName != "" || r.GroupName != "" {
		data = fmt.Sprintf("%s@@%s@@%s", data, r.GroupName, r.ServiceName)
	}
	return data
}

// InstanceRequest
type InstanceRequest struct {
	*NamingRequest
	Type     string         `json:"type"`
	Instance model.Instance `json:"instance"`
}

// NewInstanceRequest
func NewInstanceRequest(namespace, serviceName, groupName, reqType string,
	instance model.Instance) *InstanceRequest {
	return &InstanceRequest{
		NamingRequest: NewNamingRequest(namespace, serviceName, groupName),
		Type:          reqType,
		Instance:      instance,
	}
}

func (r *InstanceRequest) GetRequestType() string {
	return "InstanceRequest"
}

// BatchInstanceRequest .
type BatchInstanceRequest struct {
	*NamingRequest
	Type      string            `json:"type"`
	Instances []*model.Instance `json:"instances"`
}

func (r *BatchInstanceRequest) GetRequestType() string {
	return "BatchInstanceRequest"
}

func (r *BatchInstanceRequest) Normalize() {
	for i := range r.Instances {
		ins := r.Instances[i]
		if len(ins.ServiceName) == 0 {
			ins.ServiceName = r.ServiceName
		}
	}
}

// NotifySubscriberRequest
type NotifySubscriberRequest struct {
	*NamingRequest
	ServiceInfo *model.ServiceInfo `json:"serviceInfo"`
}

func (r *NotifySubscriberRequest) GetRequestType() string {
	return "NotifySubscriberRequest"
}

// SubscribeServiceRequest
type SubscribeServiceRequest struct {
	*NamingRequest
	Subscribe bool   `json:"subscribe"`
	Clusters  string `json:"clusters"`
}

// NewSubscribeServiceRequest .
func NewSubscribeServiceRequest(namespace, serviceName, groupName, clusters string,
	subscribe bool) *SubscribeServiceRequest {
	return &SubscribeServiceRequest{
		NamingRequest: NewNamingRequest(namespace, serviceName, groupName),
		Subscribe:     subscribe,
		Clusters:      clusters,
	}
}

func (r *SubscribeServiceRequest) GetRequestType() string {
	return "SubscribeServiceRequest"
}

// ServiceListRequest
type ServiceListRequest struct {
	*NamingRequest
	PageNo   int    `json:"pageNo"`
	PageSize int    `json:"pageSize"`
	Selector string `json:"selector"`
}

// NewServiceListRequest .
func NewServiceListRequest(namespace, serviceName, groupName string, pageNo,
	pageSize int, selector string) *ServiceListRequest {
	return &ServiceListRequest{
		NamingRequest: NewNamingRequest(namespace, serviceName, groupName),
		PageNo:        pageNo,
		PageSize:      pageSize,
		Selector:      selector,
	}
}

func (r *ServiceListRequest) GetRequestType() string {
	return "ServiceListRequest"
}

// ServiceQueryRequest
type ServiceQueryRequest struct {
	*NamingRequest
	Cluster     string `json:"cluster"`
	HealthyOnly bool   `json:"healthyOnly"`
	UdpPort     int    `json:"udpPort"`
}

// NewServiceQueryRequest .
func NewServiceQueryRequest(namespace, serviceName, groupName, cluster string, healthyOnly bool,
	udpPort int) *ServiceQueryRequest {
	return &ServiceQueryRequest{
		NamingRequest: NewNamingRequest(namespace, serviceName, groupName),
		Cluster:       cluster,
		HealthyOnly:   healthyOnly,
		UdpPort:       udpPort,
	}
}

func (r *ServiceQueryRequest) GetRequestType() string {
	return "ServiceQueryRequest"
}
