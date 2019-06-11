// +build windows
package FKAntiDebug

import (
	"os"
	"syscall"
	"regexp"
	"fmt"
	"bytes"
	"github.com/StackExchange/wmi"
	"FKSystem"
)

func AntiDebugCheck(){
	OSType := FKSystem.GetSystemOSType()
	if OSType == FKSystem.EnumSystemWindows {
		if isDetect() { // 若開啟了反調試程序的檢查，且被調試程序掛上了
			doSthAfterDetect() // 做些标准处理
		}
	}
}

var debugBlacklist 		      = [...]string{ // DEBUG工具，如果检测到这些软件，客户端将关闭退出
	"NETSTAT", "FILEMON", "PROCMON", "REGMON", "CAIN", "NETMON", "Tcpview", "vpcmap",
	"vmsrvc", "vmusrvc", "wireshark", "VBoxTray", "VBoxService", "IDA", "WPE PRO",
	"The Wireshark Network Analyzer", "WinDbg", "OllyDbg", "Colasoft Capsa", "Microsoft Network Monitor",
	"Fiddler", "SmartSniff", "Immunity Debugger", "Process Explorer", "PE Tools", "AQtime",
	"DS-5 Debug", "Dbxtool", "Topaz", "FusionDebug", "NetBeans", "Rational Purify", ".NET Reflector",
	"Cheat Engine", "Sigma Engine",
}

// 是否中獎（是否被調試程序掛上了）
func isDetect() bool {
	if detectIsNameInHashMode() || detectIsInDebuggerPresent() || detectIsDebugProcessExist() {
		return true
	}
	return false
}

// 执行中奖后处理（被调试程序挂上了之后的处理）
func doSthAfterDetect() {
	fmt.Printf("Black hat always hide himself.")
	os.Exit(-1)
}

// Step1: 檢查exe執行時，是否是hash名执行。若是，表示被調試程序掛上了
func detectIsNameInHashMode() bool {
	match, _ := regexp.MatchString("[a-f0-9]{32}", os.Args[0])
	return match
}

// Step2: 調用系統函數，檢查是否是不是被調試程序掛上了
func detectIsInDebuggerPresent() bool {
	Flag, _, _ := syscall.NewLazyDLL("kernel32.dll").
		NewProc("IsDebuggerPresent").Call()
	if Flag != 0 {
		return true
	}
	return false
}

// Step3: 检查DEBUG工具黑名单是否在进程列表中
func detectIsDebugProcessExist() bool {
	for i := 0; i < len(debugBlacklist); i++ {
		b, _ := checkIsProcessInWin32ProcessList(debugBlacklist[i])
		if  b{
			return true
		}
	}
	return false
}

type win32Process struct {
	Name           string
	ExecutablePath *string
}

// 检查当前进程列表，是否包含指定名字的进程
func checkIsProcessInWin32ProcessList(proc string) (bool, string) {
	var dst []win32Process
	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		return false, ""
	}
	for _, v := range dst {
		if bytes.Contains([]byte(v.Name), []byte(proc)) {
			return true, proc
		}
	}
	return false, ""
}