package Example

import (
	"FKTeleport"
	"log"
)

func main() {
	// 开启Teleport错误日志调试
	FKTeleport.globalIsDebugTeleport = true
	tp := FKTeleport.CreateTeleport().SetUID("C3", "abc").SetAPI(FKTeleport.API{
		"报到": new(报到),
	})
	tp.Client("127.0.0.1", ":20125")
	select {}
}

type 报到 struct{}

func (*报到) Process(receive *FKTeleport.NetData) *FKTeleport.NetData {
	if receive.Status == FKTeleport.SUCCESS {
		log.Printf("%v", receive.Body)
	}
	return nil
}
