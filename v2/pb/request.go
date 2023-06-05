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
	"encoding/json"
)

var (
	customerPayloadRegistry = map[string]func() BaseRequest{}
)

const (
	// TypeConnectionSetupRequest
	TypeConnectionSetupRequest = "ConnectionSetupRequest"
	// TypeServerCheckRequest
	TypeServerCheckRequest = "ServerCheckRequest"
	// TypeInstanceRequest
	TypeInstanceRequest = "InstanceRequest"
	// TypeBatchInstanceRequest
	TypeBatchInstanceRequest = "BatchInstanceRequest"
)

func init() {
	// system
	registryCustomerPayload(func() BaseRequest {
		return &ConnectionSetupRequest{}
	})
	registryCustomerPayload(func() BaseRequest {
		return &ConnectResetRequest{}
	})
	registryCustomerPayload(func() BaseRequest {
		return &ServerCheckRequest{}
	})
	registryCustomerPayload(func() BaseRequest {
		return &ClientDetectionRequest{}
	})
	registryCustomerPayload(func() BaseRequest {
		return &HealthCheckRequest{}
	})

	// discovery
	registryCustomerPayload(func() BaseRequest {
		return &InstanceRequest{}
	})
	registryCustomerPayload(func() BaseRequest {
		return &BatchInstanceRequest{}
	})
	registryCustomerPayload(func() BaseRequest {
		return &NotifySubscriberRequest{}
	})
	registryCustomerPayload(func() BaseRequest {
		return &SubscribeServiceRequest{}
	})
	registryCustomerPayload(func() BaseRequest {
		return &ServiceListRequest{}
	})
	registryCustomerPayload(func() BaseRequest {
		return &ServiceQueryRequest{}
	})
}

func registryCustomerPayload(builder func() BaseRequest) {
	example := builder()
	customerPayloadRegistry[example.GetRequestType()] = builder
}

// CustomerPayload
type CustomerPayload interface{}

// RequestMeta
type RequestMeta struct {
	ConnectionID  string
	ClientIP      string
	ClientVersion string
	Labels        map[string]string
}

// Request
type Request struct {
	Headers   map[string]string `json:"-"`
	RequestId string            `json:"requestId"`
}

// BaseRequest
type BaseRequest interface {
	GetHeaders() map[string]string
	GetRequestType() string
	GetBody(request BaseRequest) string
	PutAllHeaders(headers map[string]string)
	GetRequestId() string
	GetStringToSign() string
}

func (r *Request) PutAllHeaders(headers map[string]string) {
	for k, v := range headers {
		r.Headers[k] = v
	}
}

func (r *Request) ClearHeaders() {
	r.Headers = make(map[string]string)
}

func (r *Request) GetHeaders() map[string]string {
	if len(r.Headers) == 0 {
		return map[string]string{}
	}
	return r.Headers
}

func (r *Request) GetBody(request BaseRequest) string {
	//nolint:errchkjson
	js, _ := json.Marshal(request)
	return string(js)
}

func (r *Request) GetRequestId() string {
	return r.RequestId
}

func (r *Request) GetStringToSign() string {
	return ""
}
