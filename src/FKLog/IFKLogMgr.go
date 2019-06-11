package FKLog

import (
	"FKConfig"
	"fmt"
	"io"
	"os"
	"path"
)

type (
	ILogMgr interface {
		// 设置实时log信息显示终端
		SetOutput(show io.Writer) ILogMgr
		// 暂停输出日志
		Pause()
		// 恢复暂停状态，继续输出日志
		Continue()
		// 是否开启日志捕获模式
		EnableLogPeek(bool)
		// 按先后顺序实时捕获日志，每次返回1条，ok标记日志是否被关闭
		Peek() (level int, msg string, ok bool)
		// 正常关闭日志输出
		Close()
		// 返回运行状态，如1,"WORKING"
		Status() (int, string)
		// 设置当前运行状态
		SetStatus(status int)
		// 删除一种日志处理方式
		DelLogger(adapterName string) error
		// 增加一种日志处理方式
		SetLogger(adapterName string, config map[string]interface{}) error

		// 以下打印方法除正常log输出外，若为客户端或服务端模式还将进行socket信息发送
		Debug(format string, v ...interface{})
		Informational(format string, v ...interface{})
		App(format string, v ...interface{})
		Notice(format string, v ...interface{})
		Warning(format string, v ...interface{})
		Error(format string, v ...interface{})
		Critical(format string, v ...interface{})
		Alert(format string, v ...interface{})
		Emergency(format string, v ...interface{})
	}
)

var G_Log = func() ILogMgr {
	syncOnceRegisterLog.Do(func() {
		registerNewLogger("console", createConsoleLogger)
		registerNewLogger("file", createFileLogger)
		registerNewLogger("mail", createMailLogger)
		registerNewLogger("tcp", createTCPLogger)
	})
	p, _ := path.Split(FKConfig.LOG_DIR_PATH + FKConfig.DEFAULT_LOG_FILE_PATH)
	// 创建日志目录
	d, err := os.Stat(p)
	if err != nil || !d.IsDir() {
		if err := os.MkdirAll(p, 0777); err != nil {
			fmt.Errorf("Error: " + err.Error())
		}
	}

	// 创建默认日志处理机制
	lm := &defaultLogMgr{
		DefaultLogMgr: CreateDefaultLogger(FKConfig.CONFIG_LOG_CAP, FKConfig.CONFIG_LOG_TO_BACKEND_LEVEL),
	}
	// 是否打印行信息
	lm.DefaultLogMgr.EnableFuncCallDepth(FKConfig.CONFIG_IS_LOG_LINE_INFO)
	// 全局日志打印级别（亦是日志文件输出级别）
	lm.DefaultLogMgr.SetLevel(FKConfig.CONFIG_LOG_PRINT_LEVEL)
	// 是否异步输出日志
	lm.DefaultLogMgr.Async(FKConfig.IS_ASYNC_LOG)
	// 设置日志显示级别
	lm.DefaultLogMgr.SetLogger("console", map[string]interface{}{
		"level": FKConfig.CONFIG_LOG_CONSOLE_LEVEL,
	})
	// 是否保存所有日志到本地文件
	if FKConfig.CONFIG_IS_LOG_SAVE_TO_LOCAL {
		err = lm.DefaultLogMgr.SetLogger("file", map[string]interface{}{
			"filename": FKConfig.LOG_DIR_PATH + FKConfig.DEFAULT_LOG_FILE_PATH,
		})
		if err != nil {
			fmt.Printf("create log file failed：%v", err)
		}
	}

	return lm
}()

//检查并打印错误
func CheckErr(err error) {
	if err != nil {
		G_Log.Error("%v", err)
	}
}
