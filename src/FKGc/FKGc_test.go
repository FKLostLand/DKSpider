package FKGc

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func makeBuffer() []byte {
	return make([]byte, rand.Intn(5000000)+5000000)
}

func Test_StartManualGCThread(t *testing.T) {
	StartManualGCThread()
	t.Log("fk manual gc thread opened...")
	bufferPool := make([][]byte, 500)
	var i = 0
	for {
		var memStatus runtime.MemStats
		runtime.ReadMemStats(&memStatus)
		var m MemStats
		m.m = &memStatus
		println(m.toString())

		b := makeBuffer()
		bufferPool[i] = b
		i++

		time.Sleep(4 * time.Second)
	}
}
