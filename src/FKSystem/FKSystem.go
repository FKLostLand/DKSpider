package FKSystem

import "runtime"

const (
	EnumSystemUnknown = iota - 1
	EnumSystemWindows
	EnumSystemLinux
	EnumSystemMacOS
)

func GetSystemOSType() int {
	switch runtime.GOOS {
	case "darwin":
		return EnumSystemMacOS
	case "windows":
		return EnumSystemWindows
	case "linux":
		return EnumSystemLinux
	default:
		return EnumSystemUnknown
	}
}
