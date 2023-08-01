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
	"net/http"

	"github.com/emicklei/go-restful/v3"

	"github.com/polaris-contrib/apiserver-nacos/core"
	"github.com/polaris-contrib/apiserver-nacos/model"
)

func (n *NacosV1Server) GetClientServer() (*restful.WebService, error) {
	ws := new(restful.WebService)
	ws.Path("/nacos/v1/ns").Consumes(restful.MIME_JSON, model.MIME).Produces(restful.MIME_JSON)
	n.addInstanceAccess(ws)
	n.addSystemAccess(ws)
	return ws, nil
}

func (n *NacosV1Server) GetAuthServer() (*restful.WebService, error) {
	ws := new(restful.WebService)
	ws.Route(ws.POST("/v1/auth/login").To(n.Login))
	ws.Route(ws.POST("/v1/auth/users/login").To(n.Login))
	return ws, nil
}

func (n *NacosV1Server) AddServiceAccess(ws *restful.WebService) {
	ws.Route(ws.GET("/service/list").To(n.ListServices))
}

func (n *NacosV1Server) addInstanceAccess(ws *restful.WebService) {
	ws.Route(ws.POST("/instance").To(n.RegisterInstance))
	ws.Route(ws.PUT("/instance").To(n.UpdateInstance))
	ws.Route(ws.DELETE("/instance").To(n.DeRegisterInstance))
	ws.Route(ws.PUT("/instance/beat").To(n.Heartbeat))
	ws.Route(ws.GET("/instance/list").To(n.ListInstances))
}

func (n *NacosV1Server) addSystemAccess(ws *restful.WebService) {
	ws.Route(ws.GET("/operator/metrics").To(n.ServerHealthStatus))
}

func (n *NacosV1Server) Login(req *restful.Request, rsp *restful.Response) {
	handler := Handler{
		Request:  req,
		Response: rsp,
	}

	ctx := handler.ParseHeaderContext()
	data, err := n.handleLogin(ctx, ParseQueryParams(req))
	if err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}
	core.WrirteNacosResponse(data, rsp)
}

func (n *NacosV1Server) ListServices(req *restful.Request, rsp *restful.Response) {
	pageNo, err := requiredInt(req, model.ParamPageNo)
	if err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}
	pageSize, err := requiredInt(req, model.ParamPageSize)
	if err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}
	namespace := optional(req, model.ParamNamespaceID, model.DefaultNacosNamespace)
	namespace = model.ToPolarisNamespace(namespace)
	groupName := optional(req, model.ParamGroupName, model.DefaultServiceGroup)
	// selector := optional(req, model.ParamSelector, "")
	serviceList, count := model.HandleServiceListRequest(n.discoverSvr, namespace, groupName, pageNo, pageSize)
	resp := map[string]interface{}{
		"count": count,
		"doms":  serviceList,
	}
	core.WrirteNacosResponse(resp, rsp)
}

func (n *NacosV1Server) RegisterInstance(req *restful.Request, rsp *restful.Response) {
	handler := Handler{
		Request:  req,
		Response: rsp,
	}

	namespace := optional(req, model.ParamNamespaceID, model.DefaultNacosNamespace)
	namespace = model.ToPolarisNamespace(namespace)
	ins, err := BuildInstance(namespace, req)
	if err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}

	ctx := handler.ParseHeaderContext()
	if err := n.handleRegister(ctx, namespace, ins.ServiceName, ins); err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}
	core.WrirteSimpleResponse("ok", http.StatusOK, rsp)
}

func (n *NacosV1Server) UpdateInstance(req *restful.Request, rsp *restful.Response) {
	handler := Handler{
		Request:  req,
		Response: rsp,
	}

	namespace := optional(req, model.ParamNamespaceID, model.DefaultNacosNamespace)
	namespace = model.ToPolarisNamespace(namespace)
	ins, err := BuildInstance(namespace, req)
	if err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}

	ctx := handler.ParseHeaderContext()
	if err := n.handleUpdate(ctx, namespace, ins.ServiceName, ins); err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}
	core.WrirteSimpleResponse("ok", http.StatusOK, rsp)
}

func (n *NacosV1Server) DeRegisterInstance(req *restful.Request, rsp *restful.Response) {
	handler := Handler{
		Request:  req,
		Response: rsp,
	}

	namespace := optional(req, model.ParamNamespaceID, model.DefaultNacosNamespace)
	namespace = model.ToPolarisNamespace(namespace)
	ins, err := BuildInstance(namespace, req)
	if err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}

	ctx := handler.ParseHeaderContext()
	if err := n.handleDeregister(ctx, namespace, ins.ServiceName, ins); err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}
	core.WrirteSimpleResponse("ok", http.StatusOK, rsp)
}

func (n *NacosV1Server) Heartbeat(req *restful.Request, rsp *restful.Response) {
	handler := Handler{
		Request:  req,
		Response: rsp,
	}

	beat, err := BuildClientBeat(req)
	if err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}

	ctx := handler.ParseHeaderContext()
	data, err := n.handleBeat(ctx, beat.Namespace, beat.ServiceName, beat)
	if err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}
	core.WrirteNacosResponse(data, rsp)
}

func (n *NacosV1Server) ListInstances(req *restful.Request, rsp *restful.Response) {
	handler := Handler{
		Request:  req,
		Response: rsp,
	}

	ctx := handler.ParseHeaderContext()
	params := ParseQueryParams(req)

	params[model.ParamNamespaceID] = model.ToPolarisNamespace(params[model.ParamNamespaceID])
	data, err := n.handleQueryInstances(ctx, params)
	if err != nil {
		core.WrirteNacosErrorResponse(err, rsp)
		return
	}
	core.WrirteNacosResponse(data, rsp)
}

func (n *NacosV1Server) ServerHealthStatus(req *restful.Request, rsp *restful.Response) {
	core.WrirteNacosResponse(map[string]interface{}{
		"status": "UP",
	}, rsp)
}
