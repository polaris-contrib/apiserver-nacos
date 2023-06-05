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
	"github.com/polarismesh/polaris/auth"
	connlimit "github.com/polarismesh/polaris/common/conn/limit"
	"github.com/polarismesh/polaris/common/secure"
	"github.com/polarismesh/polaris/namespace"
	"github.com/polarismesh/polaris/service"
	"github.com/polarismesh/polaris/service/healthcheck"
)

type option func(svr *NacosV1Server)

func WithTLS(tlsInfo *secure.TLSInfo) option {
	return func(svr *NacosV1Server) {
		svr.tlsInfo = tlsInfo
	}
}

func WithConnLimitConfig(connLimitConfig *connlimit.Config) option {
	return func(svr *NacosV1Server) {
		svr.connLimitConfig = connLimitConfig
	}
}

func WithNamespaceSvr(namespaceSvr namespace.NamespaceOperateServer) option {
	return func(svr *NacosV1Server) {
		svr.namespaceSvr = namespaceSvr
	}
}

func WithDiscoverSvr(discoverSvr service.DiscoverServer) option {
	return func(svr *NacosV1Server) {
		svr.discoverSvr = discoverSvr
	}
}

func WithHealthSvr(healthSvr *healthcheck.Server) option {
	return func(svr *NacosV1Server) {
		svr.healthSvr = healthSvr
	}
}

func WithAuthSvr(userSvr auth.UserServer, checkerSvr auth.StrategyServer) option {
	return func(svr *NacosV1Server) {
		svr.userSvr = userSvr
		svr.checkerSvr = checkerSvr
	}
}
