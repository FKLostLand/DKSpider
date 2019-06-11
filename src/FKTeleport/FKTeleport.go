package FKTeleport

import (
	"encoding/json"
	"log"
	"time"
)

type TeleportNode struct {
	// 本节点唯一标识符
	uid string
	// 运行模式 1 SERVER  2 CLIENT (用于判断自身模式)
	mode int
	// 服务器端口号，格式如":9988"
	port string
	// 服务器地址（不含端口号），格式如"127.0.0.1"
	serverAddr string
	// 长连接池，刚一连接时key为host:port形式，随后通过身份验证替换为UID
	connPool map[string]*Connect
	// 连接时长，心跳时长的依据，默认为常量DEFAULT_TIMEOUT
	timeout time.Duration
	// 粘包处理
	*Protocol
	// 全局接收缓存通道
	apiReadChan chan *NetData
	// 每个连接对象的发送缓存通道长度
	connWChanCap int
	// 每个连接对象的接收缓冲区大小
	connBufferLen int
	// 应用程序API
	api API
	// 服务器模式专有成员
	*serverCore
	// 客户端模式专有成员
	*clientCore
}

// 设置唯一标识符，mine为本节点UID(默认ip:port)
// server为服务器UID(默认为常量DEFAULT_SERVER_UID，此参数仅客户端模式下有用)
// 可不调用该方法，此时UID均为默认
func (self *TeleportNode) SetUID(mine string, server ...string) Teleport {
	if len(server) > 0 {
		self.clientCore.serverUID = server[0]
	}
	self.uid = mine
	return self
}

// 指定应用程序API
func (self *TeleportNode) SetAPI(api API) Teleport {
	self.api = api
	return self
}

// 主动推送信息，直到有连接出现开始发送，不写nodeuid默认随机发送给一个节点
func (self *TeleportNode) Request(body interface{}, operation string, flag string, nodeuid ...string) {
	var conn *Connect
	var uid string
	if len(nodeuid) == 0 {
		for {
			if self.CountNodes() > 0 {
				break
			}
			time.Sleep(LOOP_TIMEOUT)
		}
		// 一个随机节点的信息
		for uid, conn = range self.connPool {
			if conn.Usable {
				nodeuid = append(nodeuid, uid)
				break
			}
		}
	}
	// 等待并取得连接实例
	conn = self.getConn(nodeuid[0])
	for conn == nil || !conn.Usable {
		conn = self.getConn(nodeuid[0])
		time.Sleep(LOOP_TIMEOUT)
	}
	conn.WriteChan <- CreateNetData(self.uid, nodeuid[0], operation, flag, body)
}

// 断开连接，参数为空则断开所有连接，服务器模式下停止监听
func (self *TeleportNode) Close(nodeuid ...string) {
	if self.mode == CLIENT {
		self.clientCore.mustClose = true

	} else if self.mode == SERVER && self.serverCore.listener != nil {
		self.serverCore.listener.Close()
		log.Printf(" *     —— 服务器已终止监听 %v！ ——", self.port)
	}

	if len(nodeuid) == 0 {
		// 断开全部连接
		for uid, conn := range self.connPool {
			delete(self.connPool, uid)
			conn.Close()
			self.closeMsg(uid, conn.Addr(), conn.Short)
		}
		return
	}

	for _, uid := range nodeuid {
		conn := self.connPool[uid]
		delete(self.connPool, uid)
		conn.Close()
		self.closeMsg(uid, conn.Addr(), conn.Short)
	}
}

// 设置包头字符串
func (self *TeleportNode) SetPackHeader(header string) Teleport {
	self.Protocol.ReSet(header)
	return self
}

// 设置全局接收缓存通道长度
func (self *TeleportNode) SetApiRChan(length int) Teleport {
	self.apiReadChan = make(chan *NetData, length)
	return self
}

// 设置每个连接对象的发送缓存通道长度
func (self *TeleportNode) SetConnWChan(length int) Teleport {
	self.connWChanCap = length
	return self
}

// 每个连接对象的接收缓冲区大小
func (self *TeleportNode) SetConnBuffer(length int) Teleport {
	self.connBufferLen = length
	return self
}

// 设置连接超长(心跳频率)
func (self *TeleportNode) SetTimeout(long time.Duration) Teleport {
	self.timeout = long
	return self
}

// 返回运行模式
func (self *TeleportNode) GetMode() int {
	return self.mode
}

// 返回当前有效连接节点数
func (self *TeleportNode) CountNodes() int {
	count := 0
	for _, conn := range self.connPool {
		if conn != nil && conn.Usable {
			count++
		}
	}
	return count
}

func (self *TeleportNode) read(conn *Connect) bool {
	read_len, err := conn.Read(conn.Buffer)
	if err != nil {
		return false
	}
	if read_len == 0 {
		return false // connection already closed by client
	}
	conn.TmpBuffer = append(conn.TmpBuffer, conn.Buffer[:read_len]...)
	self.save(conn)
	return true
}

// 根据uid获取连接对象
func (self *TeleportNode) getConn(nodeuid string) *Connect {
	return self.connPool[nodeuid]
}

// 根据uid获取连接对象地址
func (self *TeleportNode) getConnAddr(nodeuid string) string {
	conn := self.getConn(nodeuid)
	if conn == nil {
		// log.Printf("已与节点 %v 失去连接，无法完成发送请求！", nodeuid)
		return ""
	}
	return conn.Addr()
}

// 关闭连接，退出协程
func (self *TeleportNode) closeConn(nodeuid string, reconnect bool) {
	conn, ok := self.connPool[nodeuid]
	if !ok {
		return
	}

	// 是否允许自动重连
	if reconnect {
		self.connPool[nodeuid] = nil
	} else {
		delete(self.connPool, nodeuid)
	}

	if conn == nil {
		return
	}
	conn.Close()
	self.closeMsg(nodeuid, conn.Addr(), conn.Short)
}

// 关闭连接时log信息
func (self *TeleportNode) closeMsg(uid, addr string, short bool) {
	if short {
		return
	}
	switch self.mode {
	case SERVER:
		log.Printf(" *     —— 与客户端 %v (%v) 断开连接！——", uid, addr)
	case CLIENT:
		log.Printf(" *     —— 与服务器 %v 断开连接！——", addr)
	}
}

// 通信数据编码与发送
func (self *TeleportNode) send(data *NetData) {
	if data.From == "" {
		data.From = self.uid
	}

	d, err := json.Marshal(*data)
	if err != nil {
		Println("Debug: 发送数据-编码出错", err)
		return
	}
	conn := self.getConn(data.To)
	if conn == nil {
		Printf("Debug: 发送数据-连接已断开: %+v", data)
		return
	}
	// 封包
	end := self.Packet(d)
	// 发送
	conn.Write(end)
	Printf("Debug: 发送数据-成功: %+v", data)
}

// 解码收到的数据并存入缓存
func (self *TeleportNode) save(conn *Connect) {
	Printf("Debug: 收到数据-Byte: %v", conn.TmpBuffer)
	// 解包
	dataSlice := make([][]byte, 10)
	dataSlice, conn.TmpBuffer = self.Unpack(conn.TmpBuffer)

	for _, data := range dataSlice {
		Printf("Debug: 收到数据-解码前: %v", string(data))

		d := new(NetData)
		err := json.Unmarshal(data, d)
		if err == nil {
			// 修复缺失请求方地址的请求
			if d.From == "" {
				d.From = conn.Addr()
			}
			// 添加到读取缓存
			self.apiReadChan <- d
			Printf("Debug: 收到数据-NetData: %+v", d)
		} else {
			Printf("Debug: 收到数据-解码错误: %v", err)
		}
	}
}

// 使用API并发处理请求
func (self *TeleportNode) apiHandle() {
	for {
		req := <-self.apiReadChan
		go func(req *NetData) {
			var conn *Connect

			operation, from, to, flag := req.Operation, req.To, req.From, req.Flag
			handle, ok := self.api[operation]

			// 非法请求返回错误
			if !ok {
				// log.Printf("%+v", req)
				if self.mode == SERVER {
					self.autoErrorHandle(req, LLLEGAL, "服务器 ("+self.getConn(to).LocalAddr().String()+") 不存在API接口: "+req.Operation+"！", to)
					log.Printf("客户端 %v (%v) 正在请求不存在的API接口: %v！", to, self.getConnAddr(to), req.Operation)
				} else {
					self.autoErrorHandle(req, LLLEGAL, "客户端 "+from+" ("+self.getConn(to).LocalAddr().String()+") 不存在API接口: "+req.Operation+"！", to)
					log.Printf("服务器 (%v) 正在请求不存在的API接口: %v！", self.getConnAddr(to), req.Operation)
				}
				return
			}

			resp := handle.Process(req)
			if resp == nil {
				if conn = self.getConn(to); conn != nil && self.getConn(to).Short {
					self.closeConn(to, false)
				}
				return //continue
			}

			if resp.To == "" {
				resp.To = to
			}

			// 若指定节点连接不存在，则向原请求端返回错误
			if conn = self.getConn(resp.To); conn == nil {
				self.autoErrorHandle(req, FAILURE, "", to)
				return
			}

			// 默认指定与req相同的操作符
			if resp.Operation == "" {
				resp.Operation = operation
			}

			if resp.From == "" {
				resp.From = from
			}

			if resp.Flag == "" {
				resp.Flag = flag
			}

			conn.WriteChan <- resp

		}(req)
	}
}

func (self *TeleportNode) autoErrorHandle(data *NetData, status int, msg string, reqFrom string) bool {
	oldConn := self.getConn(reqFrom)
	if oldConn == nil {
		return false
	}
	respErr := ReturnError(data, status, msg)
	respErr.From = self.uid
	respErr.To = reqFrom
	oldConn.WriteChan <- respErr
	return true
}

// 连接权限验证
func (self *TeleportNode) checkRights(data *NetData, addr string) bool {
	if data.To != self.uid {
		log.Printf("陌生连接(%v)提供的服务器标识符错误，请求被拒绝！", addr)
		return false
	}
	return true
}

// 强制设定系统保留的API
func (self *TeleportNode) reserveAPI() {
	// 添加保留规则——身份识别
	self.api[IDENTITY] = identi
	// 添加保留规则——忽略心跳请求
	self.api[HEARTBEAT] = beat
}

var identi, beat = new(identity), new(heartbeat)

type identity struct{}

func (*identity) Process(receive *NetData) *NetData {
	return nil
}

type heartbeat struct{}

func (*heartbeat) Process(receive *NetData) *NetData {
	return nil
}
