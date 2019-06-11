package FKLog

import (
	"FKBase"
	"FKConfig"
	"errors"
	"fmt"
	"io"
	"path"
	"runtime"
	"sync"
)

type (
	defaultLogMgr struct {
		*DefaultLogMgr
	}
)

// 注册添加一个新的日志输出方式
func registerNewLogger(name string, ploggerType loggerType) {
	if ploggerType == nil {
		fmt.Errorf("Try to registe a empty logger.")
	}
	globalAdapters[name] = ploggerType
}

// 单个消息结构
type LogMsg struct {
	level int
	msg   string
}

type loggerType func() ILogger

var globalAdapters = make(map[string]loggerType)

var syncOnceRegisterLog sync.Once

// 默认Logger的状态
const (
	UnInit = iota - 1
	WORKING
	PAUSING
	CLOSED
)

type DefaultLogMgr struct {
	lock                sync.RWMutex
	level               int
	enableFuncCallDepth bool
	loggerFuncCallDepth int
	asynchronous        bool // 是否开启异步Logger
	msg                 chan *LogMsg
	peek                chan *LogMsg
	peekLevel           int
	peekLevelPreset     int
	outputs             map[string]ILogger
	status              int
}

// channelLength 是标示Chan的长度，如果这个Logger的Chan缓冲满了，那么Logger将使用其他方式进行日志处理。
func CreateDefaultLogger(channelLength int64, peekLevel ...int) *DefaultLogMgr {
	logger := new(DefaultLogMgr)
	logger.level = FKBase.LevelDebug
	logger.loggerFuncCallDepth = 2
	logger.msg = make(chan *LogMsg, channelLength)
	logger.outputs = make(map[string]ILogger)
	logger.status = WORKING
	logger.peek = make(chan *LogMsg, channelLength)
	if len(peekLevel) > 0 {
		logger.peekLevelPreset = peekLevel[0]
	} else {
		logger.peekLevelPreset = FKBase.LevelNothing
	}
	return logger
}

// 设置实时log信息显示终端
func (self *defaultLogMgr) SetOutput(show io.Writer) ILogMgr {
	self.DefaultLogMgr.SetLogger("console", map[string]interface{}{
		"writer": show,
		"level":  FKConfig.CONFIG_LOG_CONSOLE_LEVEL,
	})
	return self
}

// 暂停输出日志
func (self *DefaultLogMgr) Pause() {
	if i, _ := self.Status(); i != WORKING {
		return
	}
	self.SetStatus(PAUSING)
}

// 恢复暂停状态，继续输出日志
func (self *DefaultLogMgr) Continue() {
	if i, _ := self.Status(); i != PAUSING {
		return
	}
	self.SetStatus(WORKING)
}

// 是否开启日志捕获模式
func (self *DefaultLogMgr) EnableLogPeek(enable bool) {
	if enable {
		self.peekLevel = self.peekLevelPreset
	} else {
		self.peekLevel = FKBase.LevelNothing
	}
}

// 按先后顺序实时捕获日志，每次返回1条，ok标记日志是否被关闭
func (self *DefaultLogMgr) Peek() (level int, msg string, ok bool) {
	lm := <-self.peek
	if lm == nil {
		return 0, "", false
	}
	return lm.level, lm.msg, true
}

func (self *DefaultLogMgr) peekMsg(lm *LogMsg) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	if self.status == CLOSED {
		return
	}
	self.peek <- lm
}

// 正常关闭日志输出
func (self *DefaultLogMgr) Close() {
	self.lock.Lock()
	self.status = CLOSED
	close(self.peek)
	self.lock.Unlock()

	self.lock.RLock()
	defer self.lock.RUnlock()
	for {
		if len(self.msg) > 0 {
			bm := <-self.msg
			for _, l := range self.outputs {
				err := l.WriteMsg(bm.msg, bm.level)
				if err != nil {
					fmt.Println("ERROR, unable to WriteMsg (while closing logger):", err)
				}
			}
			continue
		}
		break
	}

	for _, l := range self.outputs {
		l.Flush()
		l.Destroy()
	}
}

// 返回运行状态，如1,"WORKING"
func (self *DefaultLogMgr) Status() (int, string) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	switch self.status {
	case WORKING:
		return WORKING, "WORKING"
	case PAUSING:
		return PAUSING, "PAUSING"
	case CLOSED:
		return CLOSED, "CLOSED"
	}
	return 0, "UNINIT"
}

func (self *DefaultLogMgr) SetStatus(status int) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.status = status
}

// 增加一种日志处理方式
func (self *DefaultLogMgr) SetLogger(adapterName string, config map[string]interface{}) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if log, ok := globalAdapters[adapterName]; ok {
		lg := log()
		err := lg.Init(config)
		self.outputs[adapterName] = lg
		if err != nil {
			fmt.Println("DefaultLogMgr.SetLogger: " + err.Error())
			return err
		}
	} else {
		return fmt.Errorf("G_Adapters: Unknown adaptername %q (forgotten registe?)", adapterName)
	}
	return nil
}

// 删除一种日志处理方式
func (self *DefaultLogMgr) DelLogger(adapterName string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if lg, ok := self.outputs[adapterName]; ok {
		lg.Destroy()
		delete(self.outputs, adapterName)
		return nil
	} else {
		return fmt.Errorf("G_Adapters: unknown adaptername %q (forgotten registe?)", adapterName)
	}
}

func (self *DefaultLogMgr) App(format string, v ...interface{}) {
	if FKBase.LevelApp > self.level {
		return
	}
	msg := fmt.Sprintf("[CUSTM] "+format, v...)
	self.writerMsg(FKBase.LevelApp, msg)
}

func (self *DefaultLogMgr) Emergency(format string, v ...interface{}) {
	if FKBase.LevelEmergency > self.level {
		return
	}
	msg := fmt.Sprintf("[EMERG] "+format, v...)
	self.writerMsg(FKBase.LevelEmergency, msg)
}

func (self *DefaultLogMgr) Alert(format string, v ...interface{}) {
	if FKBase.LevelAlert > self.level {
		return
	}
	msg := fmt.Sprintf("[ALERT] "+format, v...)
	self.writerMsg(FKBase.LevelAlert, msg)
}

func (self *DefaultLogMgr) Critical(format string, v ...interface{}) {
	if FKBase.LevelCritical > self.level {
		return
	}
	msg := fmt.Sprintf("[CRITI] "+format, v...)
	self.writerMsg(FKBase.LevelCritical, msg)
}

func (self *DefaultLogMgr) Error(format string, v ...interface{}) {
	if FKBase.LevelError > self.level {
		return
	}
	msg := fmt.Sprintf("[ERROR] "+format, v...)
	self.writerMsg(FKBase.LevelError, msg)
}

func (self *DefaultLogMgr) Warning(format string, v ...interface{}) {
	if FKBase.LevelWarning > self.level {
		return
	}
	msg := fmt.Sprintf("[WARNG] "+format, v...)
	self.writerMsg(FKBase.LevelWarning, msg)
}

func (self *DefaultLogMgr) Notice(format string, v ...interface{}) {
	if FKBase.LevelNotice > self.level {
		return
	}
	msg := fmt.Sprintf("[NOTIC] "+format, v...)
	self.writerMsg(FKBase.LevelNotice, msg)
}

func (self *DefaultLogMgr) Informational(format string, v ...interface{}) {
	if FKBase.LevelInformational > self.level {
		return
	}
	msg := fmt.Sprintf("[INFOR] "+format, v...)
	self.writerMsg(FKBase.LevelInformational, msg)
}

func (self *DefaultLogMgr) Debug(format string, v ...interface{}) {
	if FKBase.LevelDebug > self.level {
		return
	}
	msg := fmt.Sprintf("[DEBUG] "+format, v...)
	self.writerMsg(FKBase.LevelDebug, msg)
}

func (self *DefaultLogMgr) SetLevel(l int) {
	self.level = l
}

func (self *DefaultLogMgr) SetPeekLevel(l int) {
	self.peekLevel = l
}

func (self *DefaultLogMgr) SetLogFuncCallDepth(d int) {
	self.loggerFuncCallDepth = d
}

func (self *DefaultLogMgr) GetLogFuncCallDepth() int {
	return self.loggerFuncCallDepth
}

func (self *DefaultLogMgr) EnableFuncCallDepth(b bool) {
	self.enableFuncCallDepth = b
}

func (self *DefaultLogMgr) Flush() {
	for _, l := range self.outputs {
		l.Flush()
	}
}

func (self *DefaultLogMgr) Async(enable bool) *DefaultLogMgr {
	self.asynchronous = enable
	if enable {
		go self.startLogger()
	}
	return self
}

func (self *DefaultLogMgr) writerMsg(loglevel int, msg string) error {
	if i, s := self.Status(); i != WORKING {
		return errors.New("The current status is " + s)
	}

	lm := new(LogMsg)
	lm.level = loglevel
	if self.enableFuncCallDepth {
		_, file, line, ok := runtime.Caller(self.loggerFuncCallDepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)
		lm.msg = fmt.Sprintf("[%s:%d] %s", filename, line, msg)
	} else {
		lm.msg = msg
	}

	if lm.level <= self.peekLevel {
		self.peekMsg(lm)
	}

	if self.asynchronous {
		self.msg <- lm
	} else {
		self.lock.RLock()
		defer self.lock.RUnlock()
		for name, l := range self.outputs {
			err := l.WriteMsg(lm.msg, lm.level)
			if err != nil {
				fmt.Println("unable to WriteMsg to adapter:", name, err)
				return err
			}
		}
	}
	return nil
}

func (self *DefaultLogMgr) startLogger() {
	for self.asynchronous || len(self.msg) > 0 {
		select {
		case bm := <-self.msg:
			self.lock.RLock()
			for _, l := range self.outputs {
				err := l.WriteMsg(bm.msg, bm.level)
				if err != nil {
					fmt.Println("ERROR, unable to WriteMsg:", err)
				}
			}
			self.lock.RUnlock()
		}
	}
}
