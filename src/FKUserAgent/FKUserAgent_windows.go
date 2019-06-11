package FKUserAgent

import (
	"fmt"
	"runtime"
	"syscall"
)

func osName() string {
	return runtime.GOOS
}

func osVersion() string {
	v, err := syscall.GetVersion()
	if err != nil {
		return "0.0"
	}
	major := uint8(v)
	minor := uint8(v >> 8)
	return fmt.Sprintf("%d.%d", major, minor)
}
