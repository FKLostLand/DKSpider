package FKPing

import "time"

// 对外接口
func Ping(address string, timeoutSecond int) (alive bool, err error, timedelay time.Duration) {
	t := time.Now()
	err = pinger(address, timeoutSecond)
	return err == nil, err, time.Now().Sub(t)
}
