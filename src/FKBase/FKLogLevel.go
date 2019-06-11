package FKBase

import "strings"

// 遵循 RFC5424 协议的日志层级
const (
	LevelUnknown = -10
	LevelNothing = iota - 1
	LevelApp     // 特殊自定义级别
	LevelEmergency
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

func LogLevelStringToInt(l string) int {
	switch strings.ToLower(l) {
	case "app":
		return LevelApp
	case "emergency":
		return LevelEmergency
	case "alert":
		return LevelAlert
	case "critical":
		return LevelCritical
	case "error":
		return LevelError
	case "warning":
		return LevelWarning
	case "notice":
		return LevelNotice
	case "informational":
		return LevelInformational
	case "info":
		return LevelInformational
	case "debug":
		return LevelDebug
	}
	return LevelUnknown
}
