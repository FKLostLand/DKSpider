package FKLog

import (
	"testing"
)

func testConsoleCalls(self *DefaultLogMgr) {
	self.Emergency("emergency")
	self.Alert("alert")
	self.Critical("critical")
	self.Error("error")
	self.Warning("warning")
	self.Notice("notice")
	self.Informational("informational")
	self.Debug("debug")
}

func TestConsole(t *testing.T) {
	log1 := CreateDefaultLogger(10000)
	log1.EnableFuncCallDepth(true)
	log1.SetLogger("console", nil)
	testConsoleCalls(log1)

	log2 := CreateDefaultLogger(100)
	log2.SetLogger("console", map[string]interface{}{"level": 3})
	testConsoleCalls(log2)
}

func BenchmarkConsole(b *testing.B) {
	log := CreateDefaultLogger(10000)
	log.EnableFuncCallDepth(true)
	log.SetLogger("console", nil)
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
}
