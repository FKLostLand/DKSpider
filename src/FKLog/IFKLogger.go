package FKLog

type ILogger interface {
	Init(config map[string]interface{}) error
	WriteMsg(msg string, level int) error
	Destroy()
	Flush()
}
