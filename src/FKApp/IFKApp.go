package FKApp

import (
	"FKDistributor"
	"FKSpider"
	"io"
)

type (
	// App抽象接口
	IApp interface {
		Init(mode int, port int, ip string, w ...io.Writer) IApp   // 初始化
		ReInit(mode int, port int, ip string, w ...io.Writer) IApp // 切换运行模式
		SetLog(io.Writer) IApp                                     // 设置全局Log显示终端
		PauseLog() IApp                                            // 暂停Log
		ContinueLog() IApp                                         // 继续Log
		GetAppConfig(k ...string) interface{}                      // 获取全局参数
		SetAppConfig(k string, v interface{}) IApp                 // 设置全局参数
		SpiderPrepare(spiders []*FKSpider.Spider) IApp             // 准备蜘蛛（在Run()之前调用）
		Run()                                                      // 阻塞模式运行直至任务完成
		PauseRecover()                                             // Offline 模式下暂停\恢复任务
		Stop()                                                     // 终止任务
		IsRunning() bool                                           // 检查任务是否正在执行
		IsPausing() bool                                           // 检查任务是否在暂停状态
		IsStopped() bool                                           // 检查任务是否终止
		Status() int                                               // 获取当前状态
		GetSpiderByName(string) *FKSpider.Spider                   // 根据蜘蛛名获取蜘蛛对象\
		GetSpiderQueue() FKSpider.SpiderQueue                      // 获取蜘蛛队列接口实例
		GetTaskJar() *FKDistributor.DistributeTaskList             // 返回任务库
		GetOutputTypeList() []string                               // 获取全部输出方式
		GetSpiderTypeList() []*FKSpider.Spider                     // 获取蜘蛛类型
		FKDistributor.Distributor                                  // 实现分布式接口
	}
)

var G_App = CreateAppInstance()

func CreateAppInstance() IApp {
	return newApp()
}
