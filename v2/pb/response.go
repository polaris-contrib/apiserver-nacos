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

	"github.com/polaris-contrib/apiserver-nacos/model"
)

// BaseResponse
type BaseResponse interface {
	GetResponseType() string
	SetRequestId(requestId string)
	GetRequestId() string
	GetBody() string
	GetErrorCode() int
	IsSuccess() bool
	GetResultCode() int
	GetMessage() string
}

// Response
type Response struct {
	ResultCode int    `json:"resultCode"`
	ErrorCode  int    `json:"errorCode"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	RequestId  string `json:"requestId"`
}

func (r *Response) GetRequestId() string {
	return r.RequestId
}

func (r *Response) SetRequestId(requestId string) {
	r.RequestId = requestId
}

func (r *Response) GetBody() string {
	//nolint:errchkjson
	data, _ := json.Marshal(r)
	return string(data)
}

func (r *Response) IsSuccess() bool {
	return r.Success
}

func (r *Response) GetErrorCode() int {
	return r.ErrorCode
}

func (r *Response) GetResultCode() int {
	return r.ResultCode
}

func (r *Response) GetMessage() string {
	return r.Message
}

// ConnectResetResponse
type ConnectResetResponse struct {
	*Response
}

func (c *ConnectResetResponse) GetResponseType() string {
	return "ConnectResetResponse"
}

// ClientDetectionResponse
type ClientDetectionResponse struct {
	*Response
}

func (c *ClientDetectionResponse) GetResponseType() string {
	return "ClientDetectionResponse"
}

// NewServerCheckResponse
func NewServerCheckResponse() *ServerCheckResponse {
	return &ServerCheckResponse{
		Response: &Response{
			ResultCode: int(model.Response_Success.Code),
			ErrorCode:  int(model.ErrorCode_Success.Code),
			Success:    true,
			Message:    "success",
			RequestId:  "",
		},
		ConnectionId: "",
	}
}

// ServerCheckResponse
type ServerCheckResponse struct {
	*Response
	ConnectionId string `json:"connectionId"`
}

func (c *ServerCheckResponse) GetResponseType() string {
	return "ServerCheckResponse"
}

// InstanceResponse
type InstanceResponse struct {
	*Response
}

func (c *InstanceResponse) GetResponseType() string {
	return "InstanceResponse"
}

// BatchInstanceResponse
type BatchInstanceResponse struct {
	*Response
}

func (c *BatchInstanceResponse) GetResponseType() string {
	return "BatchInstanceResponse"
}

// QueryServiceResponse
type QueryServiceResponse struct {
	*Response
	ServiceInfo model.Service `json:"serviceInfo"`
}

func (c *QueryServiceResponse) GetResponseType() string {
	return "QueryServiceResponse"
}

// SubscribeServiceResponse
type SubscribeServiceResponse struct {
	*Response
	ServiceInfo model.ServiceInfo `json:"serviceInfo"`
}

func (c *SubscribeServiceResponse) GetResponseType() string {
	return "SubscribeServiceResponse"
}

// ServiceListResponse
type ServiceListResponse struct {
	*Response
	Count        int      `json:"count"`
	ServiceNames []string `json:"serviceNames"`
}

func (c *ServiceListResponse) GetResponseType() string {
	return "ServiceListResponse"
}

// NotifySubscriberResponse
type NotifySubscriberResponse struct {
	*Response
}

func (c *NotifySubscriberResponse) GetResponseType() string {
	return "NotifySubscriberResponse"
}

// NewHealthCheckResponse
func NewHealthCheckResponse() *HealthCheckResponse {
	return &HealthCheckResponse{
		Response: &Response{
			ResultCode: int(model.Response_Success.Code),
			ErrorCode:  int(model.ErrorCode_Success.Code),
			Success:    true,
			Message:    "success",
			RequestId:  "",
		},
	}
}

// HealthCheckResponse
type HealthCheckResponse struct {
	*Response
}

func (c *HealthCheckResponse) GetResponseType() string {
	return "HealthCheckResponse"
}

// ErrorResponse
type ErrorResponse struct {
	*Response
}

func (c *ErrorResponse) GetResponseType() string {
	return "ErrorResponse"
}
