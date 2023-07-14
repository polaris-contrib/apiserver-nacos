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

	"github.com/polaris-contrib/apiserver-nacos/model"
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
func NewNamingRequest() *NamingRequest {
	request := &Request{
		Headers:   make(map[string]string, 8),
		RequestId: "",
	}
	return &NamingRequest{
		Request:     request,
		Namespace:   model.DefaultNacosNamespace,
		ServiceName: "",
		GroupName:   model.DefaultServiceGroup,
		Module:      "naming",
	}
}

func NewBasicNamingRequest(requestId, namespace, serviceName, groupName string) *NamingRequest {
	request := &Request{
		Headers:   make(map[string]string, 8),
		RequestId: requestId,
	}
	return &NamingRequest{
		Request:     request,
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
func NewInstanceRequest() *InstanceRequest {
	return &InstanceRequest{
		NamingRequest: NewNamingRequest(),
		Type:          TypeInstanceRequest,
		Instance:      model.Instance{},
	}
}

func (r *InstanceRequest) GetRequestType() string {
	return TypeInstanceRequest
}

// BatchInstanceRequest .
type BatchInstanceRequest struct {
	*NamingRequest
	Type      string            `json:"type"`
	Instances []*model.Instance `json:"instances"`
}

func NewBatchInstanceRequest() *BatchInstanceRequest {
	return &BatchInstanceRequest{
		NamingRequest: NewNamingRequest(),
		Type:          TypeBatchInstanceRequest,
		Instances:     make([]*model.Instance, 0),
	}
}

func (r *BatchInstanceRequest) GetRequestType() string {
	return TypeBatchInstanceRequest
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

func NewNotifySubscriberRequest() *NotifySubscriberRequest {
	return &NotifySubscriberRequest{
		NamingRequest: NewNamingRequest(),
		ServiceInfo:   &model.ServiceInfo{},
	}
}

func (r *NotifySubscriberRequest) GetRequestType() string {
	return TypeNotifySubscriberRequest
}

// SubscribeServiceRequest
type SubscribeServiceRequest struct {
	*NamingRequest
	Subscribe bool   `json:"subscribe"`
	Clusters  string `json:"clusters"`
}

// NewSubscribeServiceRequest .
func NewSubscribeServiceRequest() *SubscribeServiceRequest {
	return &SubscribeServiceRequest{
		NamingRequest: NewNamingRequest(),
		Subscribe:     true,
		Clusters:      "",
	}
}

func (r *SubscribeServiceRequest) GetRequestType() string {
	return TypeSubscribeServiceRequest
}

// ServiceListRequest
type ServiceListRequest struct {
	*NamingRequest
	PageNo   int    `json:"pageNo"`
	PageSize int    `json:"pageSize"`
	Selector string `json:"selector"`
}

// NewServiceListRequest .
func NewServiceListRequest() *ServiceListRequest {
	return &ServiceListRequest{
		NamingRequest: NewNamingRequest(),
		PageNo:        0,
		PageSize:      10,
		Selector:      "",
	}
}

func (r *ServiceListRequest) GetRequestType() string {
	return TypeServiceListRequest
}

// ServiceQueryRequest
type ServiceQueryRequest struct {
	*NamingRequest
	Cluster     string `json:"cluster"`
	HealthyOnly bool   `json:"healthyOnly"`
	UdpPort     int    `json:"udpPort"`
}

// NewServiceQueryRequest .
func NewServiceQueryRequest() *ServiceQueryRequest {
	return &ServiceQueryRequest{
		NamingRequest: NewNamingRequest(),
		Cluster:       "",
		HealthyOnly:   false,
		UdpPort:       0,
	}
}

func (r *ServiceQueryRequest) GetRequestType() string {
	return TypeServiceQueryRequest
}
