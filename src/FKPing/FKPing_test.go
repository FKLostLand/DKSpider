package FKPing

import (
	"testing"
)

func TestPing(t *testing.T) {
	alive, err, timeDelay := Ping("127.0.0.1", 1)
	if !alive {
		t.Log("Ping address failed")
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(timeDelay.String())

	alive, err, timeDelay = Ping("confluence-wrd.pai9.net", 1)
	if !alive {
		t.Log("Ping address failed")
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(timeDelay.String())

	alive, err, timeDelay = Ping("www.baidu2.com", 1)
	if !alive {
		t.Log("Ping address failed")
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(timeDelay.String())
}
