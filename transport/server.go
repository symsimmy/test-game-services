package transport

import (
	"context"
	"github.com/symsimmy/due/cluster"
	"github.com/symsimmy/due/internal/endpoint"
	"github.com/symsimmy/due/packet"
	"github.com/symsimmy/due/session"
)

type Server interface {
	// Start 启动服务器
	Start() error
	// Stop 停止服务器
	Stop() error
	// Addr 监听地址
	Addr() string
	// Scheme 协议
	Scheme() string
	// Endpoint 服务端口
	Endpoint() *endpoint.Endpoint
	// RegisterService 注册服务
	RegisterService(desc, service interface{}) error
}

type GateProvider interface {
	// Bind 绑定用户与网关间的关系
	Bind(ctx context.Context, cid, uid int64) error
	// Unbind 解绑用户与网关间的关系
	Unbind(ctx context.Context, uid int64) error
	// GetIP 获取客户端IP地址
	GetIP(ctx context.Context, kind session.Kind, target int64) (ip string, err error)
	// IsOnline 检测是否在线
	IsOnline(ctx context.Context, kind session.Kind, target int64) (isOnline bool, err error)
	// GetID 获取conn的id
	GetID(ctx context.Context, kind session.Kind, target int64) (id int64, err error)
	// Push 发送消息（异步）
	Push(ctx context.Context, kind session.Kind, target int64, message *packet.Message) error
	// Multicast 推送组播消息（异步）
	Multicast(ctx context.Context, kind session.Kind, targets []int64, message *packet.Message) (total int64, err error)
	// Broadcast 推送广播消息（异步）
	Broadcast(ctx context.Context, kind session.Kind, message *packet.Message) (total int64, err error)
	// Disconnect 断开连接
	Disconnect(ctx context.Context, kind session.Kind, target int64, isForce bool) error
	// Stat 统计会话总数
	Stat(ctx context.Context, kind session.Kind) (total int64, err error)
	// Block
	Block(ctx context.Context, oNid string, nNid string, targets uint64)
	// Release
	Release(ctx context.Context, targets uint64)
}

type NodeProvider interface {
	// Trigger 触发事件
	Trigger(ctx context.Context, args *TriggerArgs) (miss bool, err error)
	// Deliver 投递消息
	Deliver(ctx context.Context, args *DeliverArgs) (miss bool, err error)
}

type BatchPushArgs struct {
	BatchPushArgs []*BatchPushArg
}

type BatchPushArg struct {
	kind    session.Kind
	target  int64
	message []*packet.Message
}

type DeliverArgs struct {
	GID     string
	NID     string
	CID     int64
	UID     int64
	Message *Message
}

type TriggerArgs struct {
	Event cluster.Event
	GID   string
	CID   int64
	UID   int64
}

type (
	ReceiveHandler    func(route uint16, msg []byte) (reply interface{}, err error)
	DisconnectHandler func(target string)
)

const (
	Bind        = 1
	Unbind      = 2
	Push        = 3
	Multicast   = 4
	Broadcast   = 5
	Trigger     = 6
	Deliver     = 7
	Disconnect  = 8
	GetID       = 9
	GetIP       = 10
	Stat        = 11
	IsOnline    = 12
	BlockConn   = 13
	ReleaseConn = 14
	Ping        = 15
)
