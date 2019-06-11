package FKLog

import (
	"testing"
)

func TestTcpLogger(t *testing.T) {
	log := CreateDefaultLogger(1000)
	log.SetLogger("tcp", map[string]interface{}{"net": "tcp", "addr": ":9090"})
	log.Informational("test")
}
