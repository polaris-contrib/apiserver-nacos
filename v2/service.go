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

	"github.com/polaris-contrib/apiserver-nacos/model"
	nacospb "github.com/polaris-contrib/apiserver-nacos/v2/pb"
)

func (h *NacosV2Server) handleServiceQueryRequest(ctx context.Context, req nacospb.BaseRequest,
	meta nacospb.RequestMeta) (nacospb.BaseResponse, error) {
	svcQueryReq, ok := req.(*nacospb.ServiceQueryRequest)
	if !ok {
		return nil, ErrorInvalidRequestBodyType
	}
	resp := &nacospb.QueryServiceResponse{
		Response: &nacospb.Response{
			ResultCode: int(model.Response_Success.Code),
			Success:    true,
			Message:    "success",
		},
		ServiceInfo: model.Service{
			Name:      svcQueryReq.ServiceName,
			GroupName: svcQueryReq.GroupName,
			Hosts:     make([]model.Instance, 0),
		},
	}
	namespace := model.ToPolarisNamespace(svcQueryReq.Namespace)
	svc := h.discoverSvr.Cache().Service().GetServiceByName(svcQueryReq.ServiceName, namespace)
	if svc == nil {
		return resp, nil
	}
	insts := h.discoverSvr.Cache().Instance().GetInstancesByServiceID(svc.ID)
	if insts == nil {
		return resp, nil
	}
	for _, inst := range insts {
		ins := model.Instance{}
		ins.FromSpecInstance(inst)
		resp.ServiceInfo.Hosts = append(resp.ServiceInfo.Hosts, ins)
	}
	return resp, nil
}
