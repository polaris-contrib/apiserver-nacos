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
	"encoding/json"
	"io/ioutil"
	"strings"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/polarismesh/polaris/common/utils"
)

// Handler HTTP请求/回复处理器
type Handler struct {
	Request  *restful.Request
	Response *restful.Response
}

func (h *Handler) postParseMessage(requestID string) (context.Context, error) {
	platformID := h.Request.HeaderParameter("Platform-Id")
	platformToken := h.Request.HeaderParameter("Platform-Token")
	token := h.Request.HeaderParameter("Polaris-Token")
	authToken := h.Request.HeaderParameter(utils.HeaderAuthTokenKey)
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.StringContext("request-id"), requestID)
	ctx = context.WithValue(ctx, utils.StringContext("platform-id"), platformID)
	ctx = context.WithValue(ctx, utils.StringContext("platform-token"), platformToken)
	if token != "" {
		ctx = context.WithValue(ctx, utils.StringContext("polaris-token"), token)
	}
	if authToken != "" {
		ctx = context.WithValue(ctx, utils.ContextAuthTokenKey, authToken)
	}

	var operator string
	addrSlice := strings.Split(h.Request.Request.RemoteAddr, ":")
	if len(addrSlice) == 2 {
		operator = "HTTP:" + addrSlice[0]
		if platformID != "" {
			operator += "(" + platformID + ")"
		}
	}
	if staffName := h.Request.HeaderParameter("Staffname"); staffName != "" {
		operator = staffName
	}
	ctx = context.WithValue(ctx, utils.StringContext("operator"), operator)

	return ctx, nil
}

// ParseHeaderContext 将http请求header中携带的用户信息提取出来
func (h *Handler) ParseHeaderContext() context.Context {
	requestID := h.Request.HeaderParameter("Request-Id")
	platformID := h.Request.HeaderParameter("Platform-Id")
	platformToken := h.Request.HeaderParameter("Platform-Token")
	token := h.Request.HeaderParameter("Polaris-Token")
	authToken := h.Request.HeaderParameter(utils.HeaderAuthTokenKey)

	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.StringContext("request-id"), requestID)
	ctx = context.WithValue(ctx, utils.StringContext("platform-id"), platformID)
	ctx = context.WithValue(ctx, utils.StringContext("platform-token"), platformToken)
	ctx = context.WithValue(ctx, utils.ContextClientAddress, h.Request.Request.RemoteAddr)
	if token != "" {
		ctx = context.WithValue(ctx, utils.StringContext("polaris-token"), token)
	}
	if authToken != "" {
		ctx = context.WithValue(ctx, utils.ContextAuthTokenKey, authToken)
	}

	var operator string
	addrSlice := strings.Split(h.Request.Request.RemoteAddr, ":")
	if len(addrSlice) == 2 {
		operator = "HTTP:" + addrSlice[0]
		if platformID != "" {
			operator += "(" + platformID + ")"
		}
	}
	if staffName := h.Request.HeaderParameter("Staffname"); staffName != "" {
		operator = staffName
	}
	ctx = context.WithValue(ctx, utils.StringContext("operator"), operator)

	return ctx
}

// ParseQueryParams 解析并获取HTTP的query params
func ParseQueryParams(req *restful.Request) map[string]string {
	queryParams := make(map[string]string)
	for key, value := range req.Request.URL.Query() {
		if len(value) > 0 {
			queryParams[key] = value[0] // 暂时默认只支持一个查询
		}
	}

	return queryParams
}

// ParseJsonBody parse http body as json object
func ParseJsonBody(req *restful.Request, value interface{}) error {
	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, value); err != nil {
		return err
	}
	return nil
}
