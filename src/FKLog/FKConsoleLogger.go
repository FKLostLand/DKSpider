package FKLog

import (
	"FKBase"
	"errors"
	"io"
	"log"
	"os"
	//"runtime"
	"runtime"
)

type Brush func(string) string

func NewLinuxConsoleBrush(color string) Brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

// TODO: windows 10 is supporting colorful console now.
// See: https://stackoverflow.com/questions/2048509/how-to-echo-with-different-colors-in-the-windows-command-line
func NewWindowsConsoleBrush(color string) Brush {
	return func(text string) string {
		return text
	}
}

var linuxColors = []Brush{
	NewLinuxConsoleBrush("1;37"), // LevelApp
	NewLinuxConsoleBrush("1;37"), // LevelEmergency
	NewLinuxConsoleBrush("1;36"), // LevelAlert
	NewLinuxConsoleBrush("1;35"), // LevelCritical
	NewLinuxConsoleBrush("1;31"), // LevelError
	NewLinuxConsoleBrush("1;33"), // LevelWarning
	NewLinuxConsoleBrush("1;32"), // LevelNotice
	NewLinuxConsoleBrush("1;34"), // LevelInformational
	NewLinuxConsoleBrush("1;34"), // LevelDebug
}

var windowsColors = []Brush{
	NewWindowsConsoleBrush("1;37"), // LevelApp
	NewWindowsConsoleBrush("1;37"), // LevelEmergency
	NewWindowsConsoleBrush("1;36"), // LevelAlert
	NewWindowsConsoleBrush("1;35"), // LevelCritical
	NewWindowsConsoleBrush("1;31"), // LevelError
	NewWindowsConsoleBrush("1;33"), // LevelWarning
	NewWindowsConsoleBrush("1;32"), // LevelNotice
	NewWindowsConsoleBrush("1;34"), // LevelInformational
	NewWindowsConsoleBrush("1;34"), // LevelDebug
}

type ConsoleWriter struct {
	lg    *log.Logger
	Level int `json:"level"` // 输出的日志层级
}

func createConsoleLogger() ILogger {
	cw := &ConsoleWriter{
		Level: FKBase.LevelDebug,
		lg:    log.New(os.Stdout, "", log.LstdFlags),
	}
	return cw
}

// Init console logger with json config.
// config like:
//	{
//	"writer":"os.Stdout",
//	"level":LevelTrace
//	}
func (c *ConsoleWriter) Init(config map[string]interface{}) error {
	if config == nil {
		return nil
	}
	if l, ok := config["level"]; ok {
		if l2, ok2 := l.(int); ok2 {
			c.Level = l2
		} else {
			return errors.New("console config-level's type is incorrect!")
		}
	}
	if w, ok := config["writer"]; ok {
		if w2, ok2 := w.(io.Writer); ok2 {
			c.lg = log.New(w2, "", log.LstdFlags)
		}
	}
	return nil
}

func (c *ConsoleWriter) WriteMsg(msg string, level int) error {
	if level > c.Level {
		return nil
	}
	if goos := runtime.GOOS; goos == "windows" {
		c.lg.Println(windowsColors[level](msg))
	} else {
		c.lg.Println(linuxColors[level](msg))
	}

	return nil
}

func (c *ConsoleWriter) Destroy() {

}

func (c *ConsoleWriter) Flush() {

}
