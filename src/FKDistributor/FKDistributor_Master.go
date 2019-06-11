package FKDistributor

import (
	"FKLog"
	"FKTeleport"
	"encoding/json"
)

// 创建主节点API
func CreateMaster(n Distributor) FKTeleport.API {
	return FKTeleport.API{
		// 分配任务给客户端
		"task": &masterTaskHandle{n},
		// 打印接收到的日志
		"log": &masterLogHandle{},
	}
}

// 主节点自动分配任务的操作
type masterTaskHandle struct {
	Distributor
}

func (h *masterTaskHandle) Process(receive *FKTeleport.NetData) *FKTeleport.NetData {
	b, _ := json.Marshal(h.Send(h.CountNodes()))
	return FKTeleport.ReturnData(string(b))
}

// 主节点自动接收从节点消息并打印的操作
type masterLogHandle struct{}

func (*masterLogHandle) Process(receive *FKTeleport.NetData) *FKTeleport.NetData {
	FKLog.G_Log.Informational(" * ")
	FKLog.G_Log.Informational(" *     [ %s ]    %s", receive.From, receive.Body)
	FKLog.G_Log.Informational(" * ")
	return nil
}
