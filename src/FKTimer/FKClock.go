package FKTimer

import (
	"FKLog"
	"time"
)

const (
	// 闹钟
	A = iota
	// 倒计时
	T
)

type (
	Clock struct {
		clockID    string
		clockType  int           // 模式（闹铃or倒计时）
		tol        time.Duration // 倒计时的睡眠时长 或 指定闹铃醒来时刻为从now起遇到的第tol个bell
		wakeupBell *Bell         // 闹铃醒来时刻
		timer      *time.Timer
	}
	Bell struct {
		Hour int
		Min  int
		Sec  int
	}
)

// @bell==nil时为倒计时器，此时@tol为睡眠时长
// @bell!=nil时为闹铃，此时@tol用于指定醒来时刻（从now起遇到的第tol个bell）
func createClock(id string, tol time.Duration, bell *Bell) (*Clock, bool) {
	if tol <= 0 {
		return nil, false
	}
	if bell == nil {
		return &Clock{
			clockID:   id,
			clockType: T,
			tol:       tol,
			timer:     createTimer(),
		}, true
	}
	if !(bell.Hour >= 0 && bell.Hour < 24 && bell.Min >= 0 && bell.Min < 60 && bell.Sec >= 0 && bell.Sec < 60) {
		return nil, false
	}
	return &Clock{
		clockID:    id,
		clockType:  A,
		tol:        tol,
		wakeupBell: bell,
		timer:      createTimer(),
	}, true
}

func (self *Clock) sleep() {
	d := self.duration()
	self.timer.Reset(d)
	t0 := time.Now()
	FKLog.G_Log.Critical("************************ ……定时器 <%s> 睡眠 %v ，计划 %v 醒来 ……************************", self.clockID, d, t0.Add(d).Format("2006-01-02 15:04:05"))
	<-self.timer.C
	t1 := time.Now()
	FKLog.G_Log.Critical("************************ ……定时器 <%s> 在 %v 醒来，实际睡眠 %v ……************************", self.clockID, t1.Format("2006-01-02 15:04:05"), t1.Sub(t0))
}

func (self *Clock) wake() {
	self.timer.Reset(0)
}

func (self *Clock) duration() time.Duration {
	switch self.clockType {
	case A:
		t := time.Now()
		year, month, day := t.Date()
		wakeupBell := time.Date(year, month, day, self.wakeupBell.Hour, self.wakeupBell.Min, self.wakeupBell.Sec, 0, time.Local)
		if wakeupBell.Before(t) {
			wakeupBell = wakeupBell.Add(time.Hour * 24 * self.tol)
		} else {
			wakeupBell = wakeupBell.Add(time.Hour * 24 * (self.tol - 1))
		}
		return wakeupBell.Sub(t)
	case T:
		return self.tol
	}
	return 0
}
