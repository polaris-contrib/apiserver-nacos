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
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/polarismesh/polaris/common/eventhub"
	"github.com/polarismesh/polaris/common/model"
	commontime "github.com/polarismesh/polaris/common/time"
	"github.com/polarismesh/polaris/common/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	nacospb "github.com/pole-group/nacosserver/v2/pb"
)

type (
	EventType int32

	ConnIDKey         struct{}
	ClientIPKey       struct{}
	ConnectionInfoKey struct{}

	// Client
	Client struct {
		ID             string             `json:"id"`
		Addr           *net.TCPAddr       `json:"addr"`
		ConnMeta       ConnectionMeta     `json:"conn_meta"`
		refreshTimeRef atomic.Value       `json:"refresh_time"`
		streamRef      atomic.Value       `json"json:"-"`
		closed         int32              `json:"-"`
		ctx            context.Context    `json:"-"`
		cancel         context.CancelFunc `json:"-"`
	}

	// ConnectionEvent
	ConnectionEvent struct {
		EventType EventType
		ConnID    string
		Client    *Client
	}

	// ConnectionMeta
	ConnectionMeta struct {
		ConnectType      string
		ClientIp         string
		RemoteIp         string
		RemotePort       int
		LocalPort        int
		Version          string
		ConnectionId     string
		CreateTime       time.Time
		AppName          string
		Tenant           string
		Labels           map[string]string
		ClientAttributes nacospb.ClientAbilities
	}

	// SyncServerStream
	SyncServerStream struct {
		lock   sync.Mutex
		stream grpc.ServerStream
	}
)

func (c *Client) loadStream() (*SyncServerStream, bool) {
	val := c.streamRef.Load()
	if val == nil {
		return nil, false
	}
	stream, ok := val.(*SyncServerStream)
	return stream, ok
}

func (c *Client) loadRefreshTime() time.Time {
	return c.refreshTimeRef.Load().(time.Time)
}

func (c *Client) Close() {
	c.cancel()
}

// Context returns the context for this stream.
func (s *SyncServerStream) Context() context.Context {
	return s.stream.Context()
}

func (s *SyncServerStream) SendMsg(m interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.stream.SendMsg(m)
}

const (
	ClientConnectionEvent = "ClientConnectionEvent"

	_ EventType = iota
	EventClientConnected
	EventClientDisConnected
)

type ConnectionManager struct {
	inFlights   *InFlights
	lock        sync.RWMutex
	tcpConns    map[string]net.Conn
	clients     map[string]*Client // conn_id => Client
	connections map[string]*Client // TCPAddr => Client
	cancel      context.CancelFunc
}

func newConnectionManager() *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())
	mgr := &ConnectionManager{
		connections: map[string]*Client{},
		clients:     map[string]*Client{},
		tcpConns:    make(map[string]net.Conn),
		inFlights:   &InFlights{inFlights: map[string]*ClientInFlights{}},
		cancel:      cancel,
	}
	go mgr.connectionActiveCheck(ctx)
	return mgr
}

// OnAccept call when net.Conn accept
func (h *ConnectionManager) OnAccept(conn net.Conn) {
	addr := conn.RemoteAddr().(*net.TCPAddr)

	h.lock.Lock()
	defer h.lock.Unlock()
	h.tcpConns[addr.String()] = conn
}

// OnRelease call when net.Conn release
func (h *ConnectionManager) OnRelease(conn net.Conn) {
	addr := conn.RemoteAddr().(*net.TCPAddr)

	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.tcpConns, addr.String())
}

// OnClose call when net.Listener close
func (h *ConnectionManager) OnClose() {
	// do nothing
}

func (h *ConnectionManager) RegisterConnection(ctx context.Context, payload *nacospb.Payload,
	req *nacospb.ConnectionSetupRequest) error {

	connID := ValueConnID(ctx)

	connMeta := ConnectionMeta{
		ClientIp:         payload.GetMetadata().GetClientIp(),
		Version:          "",
		ConnectionId:     connID,
		CreateTime:       time.Now(),
		AppName:          "-",
		Tenant:           req.Tenant,
		Labels:           req.Labels,
		ClientAttributes: req.ClientAbilities,
	}
	if val, ok := req.Labels["AppName"]; ok {
		connMeta.AppName = val
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	client, ok := h.clients[connID]
	if !ok {
		return errors.New("Connection register fail, Not Found target client")
	}

	client.ConnMeta = connMeta
	return nil
}

func (h *ConnectionManager) UnRegisterConnection(connID string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	eventhub.Publish(ClientConnectionEvent, ConnectionEvent{
		EventType: EventClientDisConnected,
		ConnID:    connID,
		Client:    h.clients[connID],
	})
	client, ok := h.clients[connID]
	if ok {
		delete(h.clients, connID)
		delete(h.connections, client.Addr.String())

		tcpConn, ok := h.tcpConns[client.Addr.String()]
		if ok {
			_ = tcpConn.Close()
		}
	}
}

func (h *ConnectionManager) GetClient(id string) (*Client, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	client, ok := h.clients[id]
	return client, ok
}

func (h *ConnectionManager) GetClientByAddr(addr string) (*Client, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	client, ok := h.connections[addr]
	return client, ok
}

// TagRPC can attach some information to the given context.
// The context used for the rest lifetime of the RPC will be derived from
// the returned context.
func (h *ConnectionManager) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	// do nothing
	return ctx
}

// HandleRPC processes the RPC stats.
func (h *ConnectionManager) HandleRPC(ctx context.Context, s stats.RPCStats) {
	// do nothing
}

// TagConn can attach some information to the given context.
// The returned context will be used for stats handling.
// For conn stats handling, the context used in HandleConn for this
// connection will be derived from the context returned.
// For RPC stats handling,
//   - On server side, the context used in HandleRPC for all RPCs on this
//
// connection will be derived from the context returned.
//   - On client side, the context is not derived from the context returned.
func (h *ConnectionManager) TagConn(ctx context.Context, connInfo *stats.ConnTagInfo) context.Context {
	h.lock.Lock()
	defer h.lock.Unlock()

	clientAddr := connInfo.RemoteAddr.(*net.TCPAddr)
	client, ok := h.connections[clientAddr.String()]
	if !ok {
		connId := fmt.Sprintf("%d_%s_%d_%s", commontime.CurrentMillisecond(), clientAddr.IP, clientAddr.Port,
			utils.LocalHost)
		client := &Client{
			ID:             connId,
			Addr:           clientAddr,
			refreshTimeRef: atomic.Value{},
			streamRef:      atomic.Value{},
		}
		client.refreshTimeRef.Store(time.Now())
		h.clients[connId] = client
		h.connections[clientAddr.String()] = client
	}

	client = h.connections[clientAddr.String()]
	return context.WithValue(ctx, ConnIDKey{}, client.ID)
}

// HandleConn processes the Conn stats.
func (h *ConnectionManager) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnBegin:
		h.lock.RLock()
		defer h.lock.RUnlock()
		connID, _ := ctx.Value(ConnIDKey{}).(string)
		nacoslog.Info("[NACOS-V2][ConnMgr] grpc conn begin", zap.String("conn-id", connID))
		eventhub.Publish(ClientConnectionEvent, ConnectionEvent{
			EventType: EventClientConnected,
			ConnID:    connID,
			Client:    h.clients[connID],
		})
	case *stats.ConnEnd:
		connID, _ := ctx.Value(ConnIDKey{}).(string)
		nacoslog.Info("[NACOS-V2][ConnMgr] grpc conn end", zap.String("conn-id", connID))
		h.UnRegisterConnection(connID)
	}
}

func (h *ConnectionManager) RefreshClient(ctx context.Context) {
	connID := ValueConnID(ctx)
	h.lock.RLock()
	defer h.lock.RUnlock()

	client, ok := h.clients[connID]
	if ok {
		client.refreshTimeRef.Store(time.Now())
	}
}

func (h *ConnectionManager) GetStream(connID string) *SyncServerStream {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if _, ok := h.clients[connID]; !ok {
		return nil
	}

	client := h.clients[connID]
	stream, _ := client.loadStream()
	return stream
}

func (h *ConnectionManager) listConnections() map[string]*Client {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := map[string]*Client{}
	for connID := range h.clients {
		ret[connID] = h.clients[connID]
	}
	return ret
}

func (h *ConnectionManager) connectionActiveCheck(ctx context.Context) {
	delay := time.NewTimer(1000 * time.Millisecond)
	defer delay.Stop()

	keepAliveTime := 4 * 5 * time.Second

	handler := func() {
		defer func() {
			delay.Reset(3000 * time.Millisecond)
		}()
		now := time.Now()
		connections := h.listConnections()
		outDatedConnections := map[string]*Client{}
		connIds := make([]string, 0, 4)
		for connID, conn := range connections {
			if now.Sub(conn.loadRefreshTime()) >= keepAliveTime {
				outDatedConnections[connID] = conn
				connIds = append(connIds, connID)
			}
		}

		if len(outDatedConnections) != 0 {
			nacoslog.Info("[NACOS-V2][ConnectionManager] out dated connection",
				zap.Int("size", len(outDatedConnections)), zap.Strings("conn-ids", connIds))
		}

		successConnections := new(sync.Map)
		wait := &sync.WaitGroup{}
		wait.Add(len(outDatedConnections))
		for connID := range outDatedConnections {
			req := nacospb.NewClientDetectionRequest()
			req.RequestId = utils.NewUUID()

			outDateConnectionId := connID
			outDateConnection := outDatedConnections[outDateConnectionId]
			// add inflight first
			_ = h.inFlights.AddInFlight(&InFlight{
				ConnID:    ValueConnID(ctx),
				RequestID: req.RequestId,
				Callback: func(resp nacospb.BaseResponse, err error) {
					defer wait.Done()

					select {
					case <-ctx.Done():
						// 已经结束不作处理
						return
					default:
						if resp != nil && resp.IsSuccess() {
							outDateConnection.refreshTimeRef.Store(time.Now())
							successConnections.Store(outDateConnectionId, struct{}{})
						}
					}
				},
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			defer cancel()
			wait.Wait()
		}()

		<-ctx.Done()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			// TODO log
		}
		for connID := range outDatedConnections {
			if _, ok := successConnections.Load(connID); !ok {
				h.UnRegisterConnection(connID)
			}
		}
	}

	for {
		select {
		case <-delay.C:
			handler()
		case <-ctx.Done():
			return
		}
	}
}

type (
	// ConnectionClientManager
	ConnectionClientManager struct {
		lock        sync.RWMutex
		clients     map[string]*ConnectionClient // ConnID => ConnectionClient
		inteceptors []ClientConnectionInterceptor
	}

	// ClientConnectionInterceptor
	ClientConnectionInterceptor interface {
		// HandleClientConnect .
		HandleClientConnect(ctx context.Context, client *ConnectionClient)
		// HandleClientDisConnect .
		HandleClientDisConnect(ctx context.Context, client *ConnectionClient)
	}

	// ConnectionClient .
	ConnectionClient struct {
		// ConnID 物理连接唯一ID标识
		ConnID string
		lock   sync.RWMutex
		// PublishInstances 这个连接上发布的实例信息
		PublishInstances map[model.ServiceKey]map[string]struct{}
		destroy          int32
	}
)

func newConnectionClientManager(inteceptors []ClientConnectionInterceptor) *ConnectionClientManager {
	mgr := &ConnectionClientManager{
		clients: map[string]*ConnectionClient{},
	}
	_ = eventhub.Subscribe(ClientConnectionEvent, utils.NewUUID(), mgr)
	return mgr
}

// PreProcess do preprocess logic for event
func (cm *ConnectionClientManager) PreProcess(_ context.Context, a any) any {
	return a
}

// OnEvent event process logic
func (c *ConnectionClientManager) OnEvent(ctx context.Context, a any) error {
	event, ok := a.(ConnectionEvent)
	if !ok {
		return nil
	}
	switch event.EventType {
	case EventClientConnected:
		c.addConnectionClientIfAbsent(event.ConnID)
		c.lock.RLock()
		defer c.lock.RUnlock()
		client := c.clients[event.ConnID]
		for i := range c.inteceptors {
			c.inteceptors[i].HandleClientConnect(ctx, client)
		}
	case EventClientDisConnected:
		c.lock.Lock()
		defer c.lock.Unlock()

		client, ok := c.clients[event.ConnID]
		if ok {
			for i := range c.inteceptors {
				c.inteceptors[i].HandleClientDisConnect(ctx, client)
			}
			client.Destroy()
			delete(c.clients, event.ConnID)
		}
	}

	return nil
}

func (c *ConnectionClientManager) addServiceInstance(connID string, svc model.ServiceKey, instanceIDS ...string) {
	c.addConnectionClientIfAbsent(connID)
	c.lock.RLock()
	defer c.lock.RUnlock()
	client := c.clients[connID]
	client.addServiceInstance(svc, instanceIDS...)
}

func (c *ConnectionClientManager) delServiceInstance(connID string, svc model.ServiceKey, instanceIDS ...string) {
	c.addConnectionClientIfAbsent(connID)
	c.lock.RLock()
	defer c.lock.RUnlock()
	client := c.clients[connID]
	client.delServiceInstance(svc, instanceIDS...)
}

func (c *ConnectionClientManager) addConnectionClientIfAbsent(connID string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.clients[connID]; !ok {
		client := &ConnectionClient{
			ConnID:           connID,
			PublishInstances: make(map[model.ServiceKey]map[string]struct{}),
		}
		c.clients[connID] = client
	}
}

func (c *ConnectionClient) RangePublishInstance(f func(svc model.ServiceKey, ids []string)) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for svc, ids := range c.PublishInstances {
		ret := make([]string, 0, 16)
		for i := range ids {
			ret = append(ret, i)
		}
		f(svc, ret)
	}
}

func (c *ConnectionClient) addServiceInstance(svc model.ServiceKey, instanceIDS ...string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.PublishInstances[svc]; !ok {
		c.PublishInstances[svc] = map[string]struct{}{}
	}
	publishInfos := c.PublishInstances[svc]

	for i := range instanceIDS {
		publishInfos[instanceIDS[i]] = struct{}{}
	}
	c.PublishInstances[svc] = publishInfos
}

func (c *ConnectionClient) delServiceInstance(svc model.ServiceKey, instanceIDS ...string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.PublishInstances[svc]; !ok {
		c.PublishInstances[svc] = map[string]struct{}{}
	}
	publishInfos := c.PublishInstances[svc]

	for i := range instanceIDS {
		delete(publishInfos, instanceIDS[i])
	}
	c.PublishInstances[svc] = publishInfos
}

func (c *ConnectionClient) Destroy() {
	atomic.StoreInt32(&c.destroy, 1)
}

func (c *ConnectionClient) isDestroy() bool {
	return atomic.LoadInt32(&c.destroy) == 1
}
