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

package model

import (
	"strings"
)

const (
	MIME = "application/x-www-form-urlencoded"
)

const (
	ParamCode              = "code"
	ParamServiceName       = "serviceName"
	ParamClusterList       = "clusters"
	ParamCluster           = "cluster"
	ParamClusterName       = "clusterName"
	ParamNamespaceID       = "namespaceId"
	ParamGroupName         = "groupName"
	ParamInstanceIP        = "ip"
	ParamInstancePort      = "port"
	ParamInstanceHealthy   = "healthy"
	ParamInstanceEphemeral = "ephemeral"
	ParamInstanceWeight    = "weight"
	ParamInstanceEnabled   = "enabled"
	ParamInstanceEnable    = "enable"
	ParamInstanceMetadata  = "metadata"
	ParamInstanceBeat      = "beat"
)

const (
	DefaultNacosNamespace       = "public"
	DefaultServiceGroup         = "DEFAULT_GROUP"
	DefaultServiceClusterName   = "DEFAULT"
	DefaultNacosGroupConnectStr = "@@"
	ReplaceNacosGroupConnectStr = "__"
)

const (
	InstanceMaxWeight     = float64(1000)
	InstanceMinWeight     = float64(0)
	InstanceDefaultWeight = float64(1)

	ClientBeatIntervalMill = 5000
)

const (
	InternalNacosCluster                 = "internal-nacos-cluster"
	InternalNacosServiceName             = "internal-nacos-service"
	InternalNacosServiceProtectThreshold = "internal-nacos-protectThreshold"
	InternalNacosServiceType             = "internal-nacos-service"
	InternalNacosClientConnectionID      = "internal-nacos-clientconnId"
)

func GetServiceName(s string) string {
	s = ReplaceNacosService(s)
	if !strings.Contains(s, ReplaceNacosGroupConnectStr) {
		return s
	}
	ss := strings.Split(s, ReplaceNacosGroupConnectStr)
	return ss[1]
}

func GetGroupName(s string) string {
	s = ReplaceNacosService(s)
	if !strings.Contains(s, ReplaceNacosGroupConnectStr) {
		return DefaultServiceGroup
	}
	ss := strings.Split(s, ReplaceNacosGroupConnectStr)
	return ss[0]
}

func DefaultString(s, d string) string {
	if len(s) == 0 {
		return d
	}
	return s
}
