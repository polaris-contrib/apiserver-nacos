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

package nacosserver

import (
	"context"
	"sync"

	"github.com/polarismesh/polaris/apiserver"
	"github.com/polarismesh/polaris/auth"
	connlimit "github.com/polarismesh/polaris/common/conn/limit"
	"github.com/polarismesh/polaris/common/secure"
	"github.com/polarismesh/polaris/namespace"
	"github.com/polarismesh/polaris/service"
	"github.com/polarismesh/polaris/service/healthcheck"

	"github.com/pole-group/polaris-apiserver-nacos/core"
	"github.com/pole-group/polaris-apiserver-nacos/core/push"
	nacosv1 "github.com/pole-group/polaris-apiserver-nacos/v1"
	nacosv2 "github.com/pole-group/polaris-apiserver-nacos/v2"
)

const (
	ProtooclName = "service-nacos"
)

type NacosServer struct {
	connLimitConfig *connlimit.Config
	tlsInfo         *secure.TLSInfo
	option          map[string]interface{}
	apiConf         map[string]apiserver.APIConfig

	httpPort uint32
	grpcPort uint32

	pushCenter core.PushCenter
	store      *core.NacosDataStorage

	userSvr           auth.UserServer
	strategySvr       auth.StrategyServer
	namespaceSvr      namespace.NamespaceOperateServer
	discoverSvr       service.DiscoverServer
	originDiscoverSvr service.DiscoverServer
	healthSvr         *healthcheck.Server

	v1Svr *nacosv1.NacosV1Server
	v2Svr *nacosv2.NacosV2Server
}

// GetProtocol API协议名
func (n *NacosServer) GetProtocol() string {
	return ProtooclName
}

// GetPort API的监听端口
func (n *NacosServer) GetPort() uint32 {
	return n.httpPort
}

// Initialize API初始化逻辑
func (n *NacosServer) Initialize(ctx context.Context, option map[string]interface{},
	apiConf map[string]apiserver.APIConfig) error {
	n.option = option
	n.apiConf = apiConf
	listenPort, _ := option["listenPort"].(int64)
	if listenPort == 0 {
		listenPort = 8848
	}
	n.httpPort = uint32(listenPort)
	n.grpcPort = uint32(listenPort + 1000)

	// 连接数限制的配置
	if raw, _ := option["connLimit"].(map[interface{}]interface{}); raw != nil {
		connLimitConfig, err := connlimit.ParseConnLimitConfig(raw)
		if err != nil {
			return err
		}
		n.connLimitConfig = connLimitConfig
	}

	// tls 配置信息
	if raw, _ := option["tls"].(map[interface{}]interface{}); raw != nil {
		tlsConfig, err := secure.ParseTLSConfig(raw)
		if err != nil {
			return err
		}
		n.tlsInfo = &secure.TLSInfo{
			CertFile:      tlsConfig.CertFile,
			KeyFile:       tlsConfig.KeyFile,
			TrustedCAFile: tlsConfig.TrustedCAFile,
		}
	}

	return nil
}

// Run API服务的主逻辑循环
func (n *NacosServer) Run(errCh chan error) {
	if err := n.prepareRun(); err != nil {
		errCh <- err
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		option := copyOption(n.option)
		option["listenPort"] = n.httpPort
		if err := n.v1Svr.Initialize(context.Background(), option, n.httpPort, n.apiConf); err != nil {
			errCh <- err
			return
		}
		n.v1Svr.Run(errCh)
	}()

	go func() {
		defer wg.Done()
		option := copyOption(n.option)
		option["listenPort"] = n.grpcPort
		if err := n.v2Svr.Initialize(context.Background(), n.option, n.grpcPort, n.apiConf); err != nil {
			errCh <- err
			return
		}
		n.v2Svr.Run(errCh)
	}()

	wg.Wait()
}

func copyOption(m map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{}
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

func (n *NacosServer) prepareRun() error {
	var err error
	n.namespaceSvr, err = namespace.GetServer()
	if err != nil {
		return err
	}

	n.discoverSvr, err = service.GetServer()
	if err != nil {
		return err
	}
	n.originDiscoverSvr, err = service.GetOriginServer()
	if err != nil {
		return err
	}

	n.userSvr, err = auth.GetUserServer()
	if err != nil {
		return err
	}
	n.strategySvr, err = auth.GetStrategyServer()
	if err != nil {
		return err
	}

	n.healthSvr, err = healthcheck.GetServer()
	if err != nil {
		return err
	}
	n.store = core.NewNacosDataStorage(n.discoverSvr.Cache())
	udpPush, err := push.NewUDPPushCenter(n.store)
	if err != nil {
		return err
	}
	n.v1Svr = nacosv1.NewNacosV1Server(udpPush, n.store,
		nacosv1.WithConnLimitConfig(n.connLimitConfig),
		nacosv1.WithTLS(n.tlsInfo),
		nacosv1.WithNamespaceSvr(n.namespaceSvr),
		nacosv1.WithDiscoverSvr(n.discoverSvr),
		nacosv1.WithHealthSvr(n.healthSvr),
		nacosv1.WithAuthSvr(n.userSvr, n.strategySvr),
	)

	n.v2Svr = nacosv2.NewNacosV2Server(n.v1Svr, n.store,
		nacosv2.WithConnLimitConfig(n.connLimitConfig),
		nacosv2.WithTLS(n.tlsInfo),
		nacosv2.WithNamespaceSvr(n.namespaceSvr),
		nacosv2.WithDiscoverSvr(n.discoverSvr),
		nacosv2.WithOriginDiscoverSvr(n.originDiscoverSvr),
		nacosv2.WithHealthSvr(n.healthSvr),
		nacosv2.WithAuthSvr(n.userSvr, n.strategySvr),
	)

	n.v2Svr.RegistryDebugRoute()
	return nil
}

// Stop 停止API端口监听
func (n *NacosServer) Stop() {
	if n.v1Svr != nil {
		n.v1Svr.Stop()
	}
}

// Restart 重启API
func (n *NacosServer) Restart(option map[string]interface{}, api map[string]apiserver.APIConfig,
	errCh chan error) error {
	return nil
}
