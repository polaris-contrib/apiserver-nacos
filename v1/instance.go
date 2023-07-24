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

package v1

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/polarismesh/polaris/common/utils"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/polaris-contrib/apiserver-nacos/core"
	"github.com/polaris-contrib/apiserver-nacos/model"
)

func (n *NacosV1Server) handleRegister(ctx context.Context, namespace, serviceName string, ins *model.Instance) error {
	specIns := model.PrepareSpecInstance(namespace, serviceName, ins)
	resp := n.discoverSvr.RegisterInstance(ctx, specIns)
	if apimodel.Code(resp.GetCode().GetValue()) != apimodel.Code_ExecuteSuccess {
		return &model.NacosError{
			ErrCode: int32(model.ExceptionCode_ServerError),
			ErrMsg:  resp.GetInfo().GetValue(),
		}
	}
	return nil
}

func (n *NacosV1Server) handleUpdate(ctx context.Context, namespace, serviceName string, ins *model.Instance) error {
	specIns := model.PrepareSpecInstance(namespace, serviceName, ins)
	resp := n.discoverSvr.UpdateInstance(ctx, specIns)
	if apimodel.Code(resp.GetCode().GetValue()) != apimodel.Code_ExecuteSuccess {
		return &model.NacosError{
			ErrCode: int32(model.ExceptionCode_ServerError),
			ErrMsg:  resp.GetInfo().GetValue(),
		}
	}
	return nil
}

func (n *NacosV1Server) handleDeregister(ctx context.Context, namespace, service string, ins *model.Instance) error {
	specIns := model.PrepareSpecInstance(namespace, service, ins)
	resp := n.discoverSvr.DeregisterInstance(ctx, specIns)
	if apimodel.Code(resp.GetCode().GetValue()) != apimodel.Code_ExecuteSuccess {
		return &model.NacosError{
			ErrCode: int32(model.ExceptionCode_ServerError),
			ErrMsg:  resp.GetInfo().GetValue(),
		}
	}
	return nil
}

// handleBeat com.alibaba.nacos.naming.core.InstanceOperatorClientImpl#handleBeat
func (n *NacosV1Server) handleBeat(ctx context.Context, namespace, service string,
	clientBeat *model.ClientBeat) (map[string]interface{}, error) {
	svc := n.discoverSvr.Cache().Service().GetServiceByName(service, namespace)
	if svc == nil {
		return nil, &model.NacosError{
			ErrCode: int32(model.ExceptionCode_ServerError),
			ErrMsg:  "service not found: " + service + "@" + namespace,
		}
	}

	resp := n.healthSvr.Report(ctx, &apiservice.Instance{
		Service:   utils.NewStringValue(model.ReplaceNacosService(service)),
		Namespace: utils.NewStringValue(namespace),
		Host:      utils.NewStringValue(clientBeat.Ip),
		Port:      utils.NewUInt32Value(uint32(clientBeat.Port)),
	})
	rspCode := apimodel.Code(resp.GetCode().GetValue())

	if rspCode == apimodel.Code_ExecuteSuccess {
		return map[string]interface{}{
			"code":               10200,
			"clientBeatInterval": model.ClientBeatIntervalMill,
			"lightBeatEnabled":   true,
		}, nil
	}

	if rspCode == apimodel.Code_NotFoundResource {
		return map[string]interface{}{
			"code":               20404,
			"clientBeatInterval": model.ClientBeatIntervalMill,
			"lightBeatEnabled":   true,
		}, nil
	}

	return nil, &model.NacosError{
		ErrCode: int32(model.ExceptionCode_ServerError),
		ErrMsg:  resp.GetInfo().GetValue(),
	}

}

// handleQueryInstances com.alibaba.nacos.naming.controllers.InstanceController#list
func (n *NacosV1Server) handleQueryInstances(ctx context.Context, params map[string]string) (interface{}, error) {
	namespace := params[model.ParamNamespaceID]
	group := model.GetGroupName(params[model.ParamServiceName])
	service := model.GetServiceName(params[model.ParamServiceName])
	clusters := params["clusters"]
	clientIP := params["clientIP"]
	udpPort, _ := strconv.ParseInt(params["udpPort"], 10, 32)
	healthyOnly, _ := strconv.ParseBool(params["healthyOnly"])

	if n.pushCenter != nil && udpPort > 0 {
		n.pushCenter.AddSubscriber(core.Subscriber{
			Key:         fmt.Sprintf("%s:%d", clientIP, udpPort),
			App:         model.DefaultString(params["app"], "unknown"),
			AddrStr:     clientIP,
			Ip:          clientIP,
			Port:        int(udpPort),
			NamespaceId: namespace,
			Group:       group,
			Service:     service,
			Cluster:     clusters,
			Type:        core.UDPCPush,
		})
	}

	filterCtx := &core.FilterContext{
		Service:     core.ToNacosService(n.discoverSvr.Cache(), namespace, service, group),
		Clusters:    strings.Split(clusters, ","),
		EnableOnly:  true,
		HealthyOnly: healthyOnly,
	}
	// 默认只下发 enable 的实例
	result := n.store.ListInstances(filterCtx, core.SelectInstancesWithHealthyProtection)
	// adapt for nacos v1.x SDK
	result.Name = fmt.Sprintf("%s%s%s", result.GroupName, model.DefaultNacosGroupConnectStr, result.Name)
	result.Namespace = model.ToNacosNamespace(namespace)
	return result, nil
}
