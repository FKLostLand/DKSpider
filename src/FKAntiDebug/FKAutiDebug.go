package FKAntiDebug

import (
	"time"
	"math/rand"
)

// 开机反检测休眠
func AntiCheckSleep(){
	// 反初始挂接
	takeALongRest()
}

// 反病毒检查
func AntiVirus(){
	// 反内存监视
	allocateMagicMemory()
	// 反汇编
	jump()
}

// 随机整形
func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

// 休眠 X 秒
func goToSleep(sleeptime int) {
	time.Sleep(time.Duration(sleeptime) * time.Second)
}

// 眨眨眼吧，为避免占用太高的CPU
func takeASnap() {
	time.Sleep(time.Duration(randInt(1, 5)) * time.Millisecond)
}

// 咪一下，为避免占用太高的CPU
func takeAShortRest() {
	time.Sleep(time.Duration(randInt(75, 200)) * time.Millisecond)
}

// 稍微休眠会儿，为避免被检测
func takeALongRest() {
	time.Sleep(time.Duration(randInt(250, 500)) * time.Millisecond)
}

// 慢慢睡，睡到爽……一般是没啥重要的事，但是偶尔要检查下
func takeALonglongRest() {
	time.Sleep(time.Duration(randInt(2, 5)) * time.Second)
}

// 分配魔法内存
func allocateMagicMemory() {
	var fkSonNum = 1124
	for i := 0; i < fkSonNum; i++ {
		var fkSize = 51504891
		fkBuffer := make([]byte, fkSize)
		fkBuffer[0] = 7
		var alisaBuffer [41201991]byte
		alisaBuffer[0] = 3
	}
}

var fkMagicNumber int64 = 0

func jump() {
	fkMagicNumber++
	hop1()
}

func hop1() {
	fkMagicNumber++
	takeAShortRest()
	hop2()
}
func hop2() {
	fkMagicNumber++
	takeAShortRest()
	hop3()
}
func hop3() {
	fkMagicNumber++
	takeAShortRest()
	hop4()
}
func hop4() {
	fkMagicNumber++
	takeAShortRest()
	hop5()
}
func hop5() {
	fkMagicNumber++
	takeAShortRest()
	hop6()
}
func hop6() {
	fkMagicNumber++
	takeAShortRest()
	hop7()
}
func hop7() {
	fkMagicNumber++
	takeAShortRest()
	hop8()
}
func hop8() {
	fkMagicNumber++
	takeAShortRest()
	hop9()
}
func hop9() {
	fkMagicNumber++
	takeAShortRest()
	hop10()
}
func hop10() {
	fkMagicNumber++
}
