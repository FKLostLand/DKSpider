// Teleport是一款适用于分布式系统的高并发API框架.
// 它采用socket全双工通信，实现S/C对等工作，支持长、短两种连接模式，支持断开后自动连接与手动断开连接，内部数据传输格式为JSON。
package FKTeleport

import "time"

const (
	SERVER = iota + 1
	CLIENT
)

// API中定义操作时必须保留的字段
const (
	// 身份登记
	IDENTITY = "+identity+"
	// 心跳操作符
	HEARTBEAT = "+heartbeat+"
	// 默认包头
	DEFAULT_PACK_HEADER = "fkteleport+"
	// SERVER默认UID
	DEFAULT_SERVER_UID = "server"
	// 默认端口
	DEFAULT_PORT = ":8080"
	// 服务器默认心跳间隔时长(20S)
	DEFAULT_TIMEOUT_S = 20e9
	// 客户端默认心跳间隔时长(15S)
	DEFAULT_TIMEOUT_C = 15e9
	// 等待连接的轮询时长(1s)
	LOOP_TIMEOUT = 1e9
)

type Teleport interface {
	// *以服务器模式运行，端口默认为常量DEFAULT_PORT
	Server(port ...string)
	// *以客户端模式运行，port为空时默认等于常量DEFAULT_PORT
	Client(serverAddr string, port string, isShort ...bool)
	// *主动推送信息，不写nodeuid默认随机发送给一个节点
	Request(body interface{}, operation string, flag string, nodeuid ...string)
	// 指定自定义的应用程序API
	SetAPI(api API) Teleport
	// 断开连接，参数为空则断开所有连接，服务器模式下还将停止监听
	Close(nodeuid ...string)

	// 设置唯一标识符，mine为本节点UID(默认ip:port)
	// server为服务器UID(默认为常量DEFAULT_SERVER_UID，此参数仅客户端模式下有用)
	// 可不调用该方法，此时UID均为默认
	SetUID(mine string, server ...string) Teleport
	// 设置包头字符串
	SetPackHeader(string) Teleport
	// 设置指定API处理的数据的接收缓存通道长度
	SetApiRChan(int) Teleport
	// 设置每个连接对象的发送缓存通道长度
	SetConnWChan(int) Teleport
	// 设置每个连接对象的接收缓冲区大小
	SetConnBuffer(int) Teleport
	// 设置连接超时(心跳频率)
	SetTimeout(time.Duration) Teleport

	// 返回运行模式
	GetMode() int
	// 返回当前有效连接节点数
	CountNodes() int
}

// 创建接口实例
func CreateTeleport() Teleport {
	return &TeleportNode{
		connPool:      make(map[string]*Connect),
		api:           API{},
		Protocol:      CreateProtocol(DEFAULT_PACK_HEADER),
		apiReadChan:   make(chan *NetData, 4096),
		connWChanCap:  2048,
		connBufferLen: 1024,
		serverCore:    new(serverCore),
		clientCore:    new(clientCore),
	}
}
