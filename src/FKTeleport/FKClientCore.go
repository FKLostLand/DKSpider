package FKTeleport

import (
	"log"
	"net"
	"time"
)

// 客户端专有成员
type clientCore struct {
	// 客户端模式下，控制是否为短链接
	short bool
	// 强制终止客户端
	mustClose bool
	// 服务器UID
	serverUID string
}

// 启动客户端模式
func (self *TeleportNode) Client(serverAddr string, port string, isShort ...bool) {
	if len(isShort) > 0 && isShort[0] {
		self.clientCore.short = true
	} else if self.timeout == 0 {
		// 默认心跳间隔时长
		self.timeout = DEFAULT_TIMEOUT_C
	}
	// 服务器UID默认为常量DEFAULT_SERVER_UID
	if self.clientCore.serverUID == "" {
		self.clientCore.serverUID = DEFAULT_SERVER_UID
	}
	self.reserveAPI()
	self.mode = CLIENT

	// 设置端口
	if port != "" {
		self.port = port
	} else {
		self.port = DEFAULT_PORT
	}

	self.serverAddr = serverAddr

	self.clientCore.mustClose = false

	go self.apiHandle()
	go self.client()
}

// 以客户端模式启动
func (self *TeleportNode) client() {
	if !self.short {
		log.Println(" *     —— 正在连接服务器……" + self.serverAddr + self.port)
	}

RetryLabel:
	conn, err := net.Dial("tcp", self.serverAddr + self.port)
	if err != nil {
		if self.clientCore.mustClose {
			self.clientCore.mustClose = false
			return
		}
		time.Sleep(LOOP_TIMEOUT)
		goto RetryLabel
	}
	Printf("Debug: 成功连接服务器: %v", conn.RemoteAddr().String())

	// 开启该连接处理协程(读写两条协程)
	self.cGoConn(conn)

	// 与服务器意外断开后自动重拨
	if !self.short {
		for self.CountNodes() > 0 {
			time.Sleep(LOOP_TIMEOUT)
		}
		// 判断是否为意外断开
		if _, ok := self.connPool[self.clientCore.serverUID]; ok {
			goto RetryLabel
		}
	}
}

// 为连接开启读写两个协程
func (self *TeleportNode) cGoConn(conn net.Conn) {
	remoteAddr, connect := CreateConnect(conn, self.connBufferLen, self.connWChanCap)

	// 添加连接到节点池
	self.connPool[self.clientCore.serverUID] = connect

	if self.uid == "" {
		// 设置默认UID
		self.uid = conn.LocalAddr().String()
	}

	if !self.short {
		self.send(CreateNetData(self.uid, self.clientCore.serverUID, IDENTITY, "", nil))
		log.Printf(" *     —— 成功连接到服务器：%v ——", remoteAddr)
	} else {
		connect.Short = true
	}

	// 标记连接已经正式生效可用
	self.connPool[self.clientCore.serverUID].Usable = true

	// 开启读写双工协程
	go self.cReader(self.clientCore.serverUID)
	go self.cWriter(self.clientCore.serverUID)
}

// 客户端读数据
func (self *TeleportNode) cReader(nodeuid string) {
	// 退出时关闭连接，删除连接池中的连接
	defer func() {
		self.closeConn(nodeuid, true)
	}()

	var conn = self.getConn(nodeuid)

	for {
		if !self.read(conn) {
			break
		}
	}
}

// 客户端发送数据
func (self *TeleportNode) cWriter(nodeuid string) {
	// 退出时关闭连接，删除连接池中的连接
	defer func() {
		self.closeConn(nodeuid, true)
	}()

	var conn = self.getConn(nodeuid)

	for conn != nil {
		if self.short {
			self.send(<-conn.WriteChan)
			continue
		}

		timing := time.After(self.timeout)
		data := new(NetData)
		select {
		case data = <-conn.WriteChan:
		case <-timing:
			// 保持心跳
			data = CreateNetData(self.uid, nodeuid, HEARTBEAT, "", nil)
		}

		self.send(data)
	}
}
