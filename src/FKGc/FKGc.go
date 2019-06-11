package FKGc

import (
	"encoding/json"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
	"unsafe"
)

const (
	// GC清理触发大小
	GC_SIZE = 50 << 20
	// GC检查频率间隔
	GC_CHECK_INTERVAL = 2 * time.Minute
)

var (
	syncOnceStartGCThread sync.Once
)

type MemStats struct {
	m *runtime.MemStats
}

// 开启定时手动GC
func StartManualGCThread() {
	go syncOnceStartGCThread.Do(func() {
		tick := time.Tick(GC_CHECK_INTERVAL)
		for {
			<-tick
			var memStatus runtime.MemStats
			runtime.ReadMemStats(&memStatus)
			var m MemStats
			m.m = &memStatus
			//fmt.Println(m.toString())
			if memStatus.HeapReleased >= GC_SIZE {
				debug.FreeOSMemory()
				runtime.GC()
			}
		}
	})
}

func (m MemStats) toString() string {
	s, _ := json.Marshal(m.m)
	return *(*string)(unsafe.Pointer(&s))
}
