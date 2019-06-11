package FKTimer

import (
	"FKLog"
	"sync"
	"time"
)

func CreateTimer() *Timer {
	return &Timer{
		setting: make(map[string]*Clock),
	}
}

type Timer struct {
	setting map[string]*Clock
	closed  bool
	sync.RWMutex
}

// 休眠等待，并返回定时器是否可以继续使用
func (self *Timer) Sleep(id string) bool {
	self.RLock()
	if self.closed {
		self.RUnlock()
		return false
	}

	c, ok := self.setting[id]
	self.RUnlock()
	if !ok {
		return false
	}

	c.sleep()

	self.RLock()
	defer self.RUnlock()
	if self.closed {
		return false
	}
	_, ok = self.setting[id]

	return ok
}

// @bell==nil时为倒计时器，此时@tol为睡眠时长
// @bell!=nil时为闹铃，此时@tol用于指定醒来时刻（从now起遇到的第tol个bell）
func (self *Timer) Set(id string, tol time.Duration, bell *Bell) bool {
	self.Lock()
	defer self.Unlock()

	if self.closed {
		FKLog.G_Log.Critical("************************ ……设置定时器 [%s] 失败，定时系统已关闭 ……************************", id)
		return false
	}
	c, ok := createClock(id, tol, bell)
	if !ok {
		FKLog.G_Log.Critical("************************ ……设置定时器 [%s] 失败，参数不正确 ……************************", id)
		return ok
	}
	self.setting[id] = c
	FKLog.G_Log.Critical("************************ ……设置定时器 [%s] 成功 ……************************", id)
	return ok
}

func (self *Timer) Drop() {
	self.Lock()
	defer self.Unlock()

	self.closed = true
	for _, c := range self.setting {
		c.wake()
	}
	self.setting = make(map[string]*Clock)
}

func createTimer() *time.Timer {
	t := time.NewTimer(0)
	<-t.C
	return t
}
