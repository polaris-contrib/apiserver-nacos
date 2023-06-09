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

	"github.com/pole-group/polaris-apiserver-nacos/model"
	nacosmodel "github.com/pole-group/polaris-apiserver-nacos/model"
	nacospb "github.com/pole-group/polaris-apiserver-nacos/v2/pb"
)

func (h *NacosV2Server) handleServiceListRequest(ctx context.Context, req nacospb.BaseRequest,
	meta nacospb.RequestMeta) (nacospb.BaseResponse, error) {
	svcListReq, ok := req.(*nacospb.ServiceListRequest)
	if !ok {
		return nil, ErrorInvalidRequestBodyType
	}
	resp := &nacospb.ServiceListResponse{
		Response: &nacospb.Response{
			ResultCode: int(model.ErrorCode_Success.Code),
			Success:    true,
			Message:    "success",
		},
		Count:        0,
		ServiceNames: []string{},
	}

	namespace := nacosmodel.ToPolarisNamespace(svcListReq.Namespace)
	_, services := h.discoverSvr.Cache().Service().ListServices(namespace)
	offset := (svcListReq.PageNo - 1) * svcListReq.PageSize
	limit := svcListReq.PageSize
	if offset < 0 {
		offset = 0
	}
	if offset > len(services) {
		return resp, nil
	}
	groupPrefix := svcListReq.GroupName + model.ReplaceNacosGroupConnectStr
	if svcListReq.GroupName == model.DefaultServiceGroup {
		groupPrefix = ""
	}
	hasGroupPrefix := len(groupPrefix) != 0
	temp := make([]string, 0, len(services))
	for i := range services {
		svc := services[i]
		if !hasGroupPrefix {
			temp = append(temp, svc.Name)
			continue
		}
		if strings.HasPrefix(svc.Name, groupPrefix) {
			temp = append(temp, model.GetServiceName(svc.Name))
		}
	}
	var viewList []string
	if offset+limit > len(services) {
		viewList = temp[offset:]
	} else {
		viewList = temp[offset : offset+limit]
	}
	resp.ServiceNames = viewList
	resp.Count = len(temp)
	return resp, nil
}
