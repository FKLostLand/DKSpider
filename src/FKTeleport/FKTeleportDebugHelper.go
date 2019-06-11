package FKTeleport

import "log"

// 错误调试
var globalIsDebugTeleport bool

func Printf(format string, v ...interface{}) {
	if !globalIsDebugTeleport {
		return
	}
	log.Printf(format, v...)
}

func Println(v ...interface{}) {
	if !globalIsDebugTeleport {
		return
	}
	log.Println(v...)
}

func Fatal(v ...interface{}) {
	if !globalIsDebugTeleport {
		return
	}
	log.Fatal(v...)
}
