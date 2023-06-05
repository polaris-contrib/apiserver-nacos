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

package core

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"time"

	"github.com/polarismesh/polaris/common/model"

	nacosmodel "github.com/pole-group/nacosserver/model"
)

type PushType string

const (
	NoopPush     PushType = "noop"
	UDPCPush     PushType = "udp"
	GRPCPush     PushType = "grpc"
	AssemblyPush PushType = "assembly"
)

type PushCenter interface {
	AddSubscriber(s Subscriber)
	RemoveSubscriber(s Subscriber)
	EnablePush(s Subscriber) bool
	Type() PushType
}

type PushData struct {
	Service          *model.Service
	ServiceInfo      *nacosmodel.ServiceInfo
	UDPData          interface{}
	CompressUDPData  []byte
	GRPCData         interface{}
	CompressGRPCData []byte
}

func WarpGRPCPushData(p *PushData) {
	data := map[string]interface{}{
		"type": "dom",
		"data": map[string]interface{}{
			"dom":             p.Service.Name,
			"cacheMillis":     p.ServiceInfo.CacheMillis,
			"lastRefTime":     p.ServiceInfo.LastRefTime,
			"checksum":        p.ServiceInfo.Checksum,
			"useSpecifiedURL": false,
			"hosts":           p.ServiceInfo.Hosts,
			"metadata":        p.Service.Meta,
		},
		"lastRefTime": time.Now().Nanosecond(),
	}
	p.GRPCData = data
	//nolint:errchkjson
	body, _ := json.Marshal(data)
	p.CompressGRPCData = CompressIfNecessary(body)
}

func WarpUDPPushData(p *PushData) {
	data := map[string]interface{}{
		"type": "dom",
		"data": map[string]interface{}{
			"dom":             p.Service.Name,
			"cacheMillis":     p.ServiceInfo.CacheMillis,
			"lastRefTime":     p.ServiceInfo.LastRefTime,
			"checksum":        p.ServiceInfo.Checksum,
			"useSpecifiedURL": false,
			"hosts":           p.ServiceInfo.Hosts,
			"metadata":        p.Service.Meta,
		},
		"lastRefTime": time.Now().Nanosecond(),
	}
	p.UDPData = data
	//nolint:errchkjson
	body, _ := json.Marshal(data)
	p.CompressUDPData = CompressIfNecessary(body)
}

const (
	maxDataSizeUncompress = 1024
)

func CompressIfNecessary(data []byte) []byte {
	if len(data) <= maxDataSizeUncompress {
		return data
	}

	var ret bytes.Buffer
	writer := gzip.NewWriter(&ret)
	_, err := writer.Write(data)
	if err != nil {
		return data
	}
	return ret.Bytes()
}

type Subscriber struct {
	Key         string
	ConnID      string
	AddrStr     string
	Agent       string
	App         string
	Ip          string
	Port        int
	NamespaceId string
	Group       string
	Service     string
	Cluster     string
	Type        PushType
}
