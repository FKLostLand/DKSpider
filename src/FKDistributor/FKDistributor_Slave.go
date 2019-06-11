package FKDistributor

import (
	"FKLog"
	"FKTeleport"
	"encoding/json"
)

// 创建从节点API
func CreateSlave(n Distributor) FKTeleport.API {
	return FKTeleport.API{
		// 接收来自服务器的任务并加入任务库
		"task": &slaveTaskHandle{n},
	}
}

// 从节点自动接收主节点任务的操作
type slaveTaskHandle struct {
	Distributor
}

func (h *slaveTaskHandle) Process(receive *FKTeleport.NetData) *FKTeleport.NetData {
	t := &DistributeTask{}
	err := json.Unmarshal([]byte(receive.Body.(string)), t)
	if err != nil {
		FKLog.G_Log.Error("json解码失败 %v", receive.Body)
		return nil
	}
	h.Receive(t)
	return nil
}
