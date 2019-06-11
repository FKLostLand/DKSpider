package FKUserAgent

import (
	"runtime"
	"syscall"
)

func osName() string {
	buf := &syscall.Utsname{}
	err := syscall.Uname(buf)
	if err != nil {
		return runtime.GOOS
	}
	return charsToString(buf.Sysname)
}

func osVersion() string {
	buf := &syscall.Utsname{}
	err := syscall.Uname(buf)
	if err != nil {
		return "0.0"
	}
	return charsToString(buf.Release)
}

func charsToString(ca [65]int8) string {
	s := make([]byte, len(ca))
	var lens int
	for ; lens < len(ca); lens++ {
		if ca[lens] == 0 {
			break
		}
		s[lens] = uint8(ca[lens])
	}
	return string(s[0:lens])
}
