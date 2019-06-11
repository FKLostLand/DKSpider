package Example

import (
	"FKTeleport"
)

// 有标识符UID的demo，保证了客户端链接唯一性
func main() {
	FKTeleport.globalIsDebugTeleport = true
	tp := FKTeleport.CreateTeleport().SetUID("C2", "abc")
	tp.Client("127.0.0.1", ":20125", true)
	tp.Request("我是短链接客户端，我来报个到", "短链接报到", "shortOne")
	select {}
}
