package Example

import (
	"FKTeleport"
	"log"
)

// 有标识符UID的demo，保证了客户端链接唯一性
var tp = FKTeleport.CreateTeleport()

func main() {
	// 开启Teleport错误日志调试
	FKTeleport.globalIsDebugTeleport = true
	tp.SetUID("fad", "abc").SetAPI(FKTeleport.API{
		"报到":     new(报到),
		"非法请求测试": new(非法请求测试),
	})
	tp.Client("127.0.0.1", ":20125")
	tp.Request("我是客户端，我来报个到", "报到", "f")
	tp.Request("我是客户端，我来报个到", "非法请求测试", "")
	select {}
}

type 报到 struct{}

func (*报到) Process(receive *FKTeleport.NetData) *FKTeleport.NetData {
	if receive.Status == FKTeleport.SUCCESS {
		log.Printf("%v", receive.Body)
	}
	if receive.Status == FKTeleport.FAILURE {
		log.Printf("%v", "请求处理失败！")
	}
	return nil
}

type 非法请求测试 struct{}

func (*非法请求测试) Process(receive *FKTeleport.NetData) *FKTeleport.NetData {
	log.Printf("%v", receive.Body)
	tp.Close()
	return nil
}
