package FKApp

import (
	"FKCrawler"
	"FKDistributor"
	"FKLog"
	"FKPipeline"
	"FKScheduler"
	"FKSpider"
	"FKStatus"
	"FKTeleport"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	App struct {
		*FKStatus.AppRuntimeConfig                      // 全局配置
		*FKSpider.SpiderSpecies                         // 全部蜘蛛种类
		FKSpider.SpiderQueue                            // 当前任务的蜘蛛队列
		*FKDistributor.DistributeTaskList               // 服务器与客户端间传递任务的存储库
		FKCrawler.CrawlerPool                           // 爬行回收池
		FKTeleport.Teleport                             // socket长连接双工通信接口，json数据传输
		sum                               [2]uint64     // 执行计数
		takeTime                          time.Duration // 执行计时
		status                            int           // 运行状态
		finish                            chan bool
		syncOnceExitApp                   sync.Once
		canSocketLog                      bool
		sync.RWMutex
	}
)

func newApp() *App {
	return &App{
		AppRuntimeConfig:   FKStatus.GlobalRuntimeTaskConfig,
		SpiderSpecies:      FKSpider.GlobalSpiderSpecies,
		status:             FKStatus.UNINIT,
		Teleport:           FKTeleport.CreateTeleport(),
		DistributeTaskList: FKDistributor.CreateFKDistributeTaskList(),
		SpiderQueue:        FKSpider.CreateSpiderQueue(),
		CrawlerPool:        FKCrawler.CreateCrawlerPool(),
	}
}

// 使用App前必须先进行Init初始化（SetLog()除外）
func (a *App) Init(mode int, port int, master string, w ...io.Writer) IApp {
	a.canSocketLog = false
	if len(w) > 0 {
		a.SetLog(w[0])
	}
	a.ContinueLog()

	a.AppRuntimeConfig.Mode, a.AppRuntimeConfig.MasterPort, a.AppRuntimeConfig.MasterIP = mode, port, master
	a.Teleport = FKTeleport.CreateTeleport()
	a.DistributeTaskList = FKDistributor.CreateFKDistributeTaskList()
	a.SpiderQueue = FKSpider.CreateSpiderQueue()
	a.CrawlerPool = FKCrawler.CreateCrawlerPool()

	switch a.AppRuntimeConfig.Mode {
	case FKStatus.SERVER:
		FKLog.G_Log.EnableLogPeek(false)
		if a.checkPort() {
			FKLog.G_Log.Informational(" *     —— 当前运行模式为：[ 服务器 ] 模式")
			a.Teleport.SetAPI(FKDistributor.CreateMaster(a)).Server(":" +
				strconv.Itoa(a.AppRuntimeConfig.MasterPort))
		}

	case FKStatus.CLIENT:
		if a.checkAll() {
			FKLog.G_Log.Informational(" *     —— 当前运行模式为：[ 客户端 ] 模式")
			a.Teleport.SetAPI(FKDistributor.CreateSlave(a)).Client(
				a.AppRuntimeConfig.MasterIP, ":"+strconv.Itoa(a.AppRuntimeConfig.MasterPort))
			// 开启节点间log打印
			a.canSocketLog = true
			FKLog.G_Log.EnableLogPeek(true)
			go a.socketLog()
		}
	case FKStatus.OFFLINE:
		FKLog.G_Log.EnableLogPeek(false)
		FKLog.G_Log.Informational(" *     —— 当前运行模式为：[ 单机 ] 模式")
		return a
	default:
		FKLog.G_Log.Warning(" *     —— 请指定正确的运行模式！——")
		return a
	}
	return a
}

// 切换运行模式时使用
func (a *App) ReInit(mode int, port int, master string, w ...io.Writer) IApp {
	if !a.IsStopped() {
		a.Stop()
	}
	a.PauseLog()
	if a.Teleport != nil {
		a.Teleport.Close()
	}
	// 等待结束
	if mode == FKStatus.UNSET {
		a = newApp()
		a.AppRuntimeConfig.Mode = FKStatus.UNSET
		return a
	}
	// 重新开启
	a = newApp().Init(mode, port, master, w...).(*App)
	return a
}

// 设置全局log实时显示终端
func (a *App) SetLog(w io.Writer) IApp {
	FKLog.G_Log.SetOutput(w)
	return a
}

// 暂停log打印
func (a *App) PauseLog() IApp {
	FKLog.G_Log.Pause()
	return a
}

// 继续log打印
func (a *App) ContinueLog() IApp {
	FKLog.G_Log.Continue()
	return a
}

// 获取全局参数
func (a *App) GetAppConfig(k ...string) interface{} {
	defer func() {
		if err := recover(); err != nil {
			FKLog.G_Log.Error("%v", err)
		}
	}()
	if len(k) == 0 {
		return a.AppRuntimeConfig
	}
	key := strings.Title(k[0])
	acv := reflect.ValueOf(a.AppRuntimeConfig).Elem()
	return acv.FieldByName(key).Interface()
}

// 设置全局参数
func (a *App) SetAppConfig(k string, v interface{}) IApp {
	defer func() {
		if err := recover(); err != nil {
			FKLog.G_Log.Error("%v", err)
		}
	}()
	if k == "RequestLimit" && v.(int64) <= 0 {
		v = int64(math.MaxInt64)
	} else if k == "DockerCap" && v.(int) < 1 {
		v = int(1)
	}
	acv := reflect.ValueOf(a.AppRuntimeConfig).Elem()
	key := strings.Title(k)
	if acv.FieldByName(key).CanSet() {
		acv.FieldByName(key).Set(reflect.ValueOf(v))
	}

	return a
}

// SpiderPrepare()必须在设置全局运行参数之后，Run()的前一刻执行
// original为spider包中未有过赋值操作的原始蜘蛛种类
// 已被显式赋值过的spider将不再重新分配Keywords
// client模式下不调用该方法
func (a *App) SpiderPrepare(original []*FKSpider.Spider) IApp {
	a.SpiderQueue.Reset()
	// 遍历任务
	for _, sp := range original {
		spcopy := sp.Copy()
		spcopy.SetPausetime(a.AppRuntimeConfig.MedianPauseTime)
		if spcopy.GetLimit() == math.MaxInt64 {
			spcopy.SetLimit(a.AppRuntimeConfig.RequestLimit)
		} else {
			spcopy.SetLimit(-1 * a.AppRuntimeConfig.RequestLimit)
		}
		a.SpiderQueue.Add(spcopy)
	}
	// 遍历自定义配置
	a.SpiderQueue.AddKeywords(a.AppRuntimeConfig.Keywords)
	return a
}

// 运行任务
func (a *App) Run() {
	// 确保开启报告
	a.ContinueLog()
	if a.AppRuntimeConfig.Mode != FKStatus.CLIENT && a.SpiderQueue.Len() == 0 {
		FKLog.G_Log.Warning(" *     —— 任务列表不能为空")
		a.PauseLog()
		return
	}
	a.finish = make(chan bool)
	a.syncOnceExitApp = sync.Once{}
	// 重置计数
	a.sum[0], a.sum[1] = 0, 0
	// 重置计时
	a.takeTime = 0
	// 设置状态
	a.setStatus(FKStatus.RUN)
	defer a.setStatus(FKStatus.UNINIT)
	// 任务执行
	switch a.AppRuntimeConfig.Mode {
	case FKStatus.OFFLINE:
		a.offline()
	case FKStatus.SERVER:
		a.server()
	case FKStatus.CLIENT:
		a.client()
	default:
		return
	}
	<-a.finish
}

// Offline 模式下暂停\恢复任务
func (a *App) PauseRecover() {
	switch a.Status() {
	case FKStatus.PAUSE:
		a.setStatus(FKStatus.RUN)
	case FKStatus.RUN:
		a.setStatus(FKStatus.PAUSE)
	}

	FKScheduler.PauseRecover()
}

// Offline 模式下中途终止任务
func (a *App) Stop() {
	if a.status == FKStatus.UNINIT {
		return
	}
	if a.status != FKStatus.STOP {
		// Warning：不可颠倒停止的顺序
		a.setStatus(FKStatus.STOP)
		FKScheduler.Stop()
		a.CrawlerPool.StopAllCrawlers()
	}
	for !a.IsStopped() {
		time.Sleep(time.Second)
	}
}

// 检查任务是否正在运行
func (a *App) IsRunning() bool {
	return a.status == FKStatus.RUN
}

// 检查任务是否处于暂停状态
func (a *App) IsPausing() bool {
	return a.status == FKStatus.PAUSE
}

// 检查任务是否已经终止
func (a *App) IsStopped() bool {
	return a.status == FKStatus.UNINIT
}

// 返回当前运行状态
func (a *App) Status() int {
	a.RWMutex.RLock()
	defer a.RWMutex.RUnlock()
	return a.status
}

// 通过名字获取某蜘蛛
func (a *App) GetSpiderByName(name string) *FKSpider.Spider {
	return a.SpiderSpecies.GetByName(name)
}

// 获取蜘蛛队列接口实例
func (a *App) GetSpiderQueue() FKSpider.SpiderQueue {
	return a.SpiderQueue
}

// 返回任务库
func (a *App) GetTaskJar() *FKDistributor.DistributeTaskList {
	return a.DistributeTaskList
}

// 获取全部输出方式
func (a *App) GetOutputTypeList() []string {
	return FKPipeline.G_DataOutputLib
}

// 获取全部蜘蛛种类
func (a *App) GetSpiderTypeList() []*FKSpider.Spider {
	return a.SpiderSpecies.Get()
}

// 返回当前运行模式
func (a *App) GetMode() int {
	return a.AppRuntimeConfig.Mode
}

// 服务器客户端模式下返回节点数
func (a *App) CountNodes() int {
	return a.Teleport.CountNodes()
}

// 返回当前运行状态
func (a *App) setStatus(status int) {
	a.RWMutex.Lock()
	defer a.RWMutex.Unlock()
	a.status = status
}

// 离线模式运行
func (a *App) offline() {
	a.exec()
}

// 服务器模式运行，必须在SpiderPrepare()执行之后调用才可以成功添加任务
// 生成的任务与自身当前全局配置相同
func (a *App) server() {
	// 标记结束
	defer func() {
		a.syncOnceExitApp.Do(func() { close(a.finish) })
	}()

	// 便利添加任务到库
	tasksNum, spidersNum := a.addNewTask()

	if tasksNum == 0 {
		return
	}

	// 打印报告
	FKLog.G_Log.Informational(" * ")
	FKLog.G_Log.Informational(` *********************************************************************************************************************************** `)
	FKLog.G_Log.Informational(" * ")
	FKLog.G_Log.Informational(" *                               —— 本次成功添加 %v 条任务，共包含 %v 条采集规则 ——", tasksNum, spidersNum)
	FKLog.G_Log.Informational(" * ")
	FKLog.G_Log.Informational(` *********************************************************************************************************************************** `)
}

// 服务器模式下，生成task并添加至库
func (a *App) addNewTask() (tasksNum, spidersNum int) {
	length := a.SpiderQueue.Len()
	t := FKDistributor.DistributeTask{}
	// 从配置读取字段
	a.setTask(&t)

	for i, sp := range a.SpiderQueue.GetAll() {

		t.Spiders = append(t.Spiders, map[string]string{"name": sp.GetName(), "keywords": sp.GetKeywords()})
		spidersNum++

		// 每十个蜘蛛存为一个任务
		if i > 0 && i%10 == 0 && length > 10 {
			// 存入
			one := t
			a.DistributeTaskList.Push(&one)
			//FKLog.G_Log.App(" *     [新增任务]   详情： %#v", *t)

			tasksNum++

			// 清空spider
			t.Spiders = []map[string]string{}
		}
	}

	if len(t.Spiders) != 0 {
		// 存入
		one := t
		a.DistributeTaskList.Push(&one)
		tasksNum++
	}
	return
}

// 客户端模式运行
func (a *App) client() {
	// 标记结束
	defer func() {
		a.syncOnceExitApp.Do(func() { close(a.finish) })
	}()

	for {
		// 从任务库获取一个任务
		t := a.downTask()

		if a.Status() == FKStatus.STOP || a.Status() == FKStatus.UNINIT {
			return
		}

		// 准备运行
		a.taskToRun(t)

		// 重置计数
		a.sum[0], a.sum[1] = 0, 0
		// 重置计时
		a.takeTime = 0

		// 执行任务
		a.exec()
	}
}

// 客户端模式下获取任务
func (a *App) downTask() *FKDistributor.DistributeTask {
ReStartLabel:
	if a.Status() == FKStatus.STOP || a.Status() == FKStatus.UNINIT {
		return nil
	}
	if a.CountNodes() == 0 && a.DistributeTaskList.Len() == 0 {
		time.Sleep(time.Second)
		goto ReStartLabel
	}

	if a.DistributeTaskList.Len() == 0 {
		a.Request(nil, "task", "")
		for a.DistributeTaskList.Len() == 0 {
			if a.CountNodes() == 0 {
				goto ReStartLabel
			}
			time.Sleep(time.Second)
		}
	}
	return a.DistributeTaskList.Pull()
}

// client模式下从task准备运行条件
func (a *App) taskToRun(t *FKDistributor.DistributeTask) {
	// 清空历史任务
	a.SpiderQueue.Reset()

	// 更改全局配置
	a.setAppConf(t)

	// 初始化蜘蛛队列
	for _, n := range t.Spiders {
		sp := a.GetSpiderByName(n["name"])
		if sp == nil {
			continue
		}
		spcopy := sp.Copy()
		spcopy.SetPausetime(t.Pausetime)
		if spcopy.GetLimit() > 0 {
			spcopy.SetLimit(t.Limit)
		} else {
			spcopy.SetLimit(-1 * t.Limit)
		}
		if v, ok := n["keywords"]; ok {
			spcopy.SetKeywords(v)
		}
		a.SpiderQueue.Add(spcopy)
	}
}

// 开始执行任务
func (a *App) exec() {
	count := a.SpiderQueue.Len()
	FKStatus.ResetRequestPageNum()
	// 刷新输出方式的状态
	FKPipeline.RefreshOutput()
	// 初始化资源队列
	FKScheduler.Init()

	// 设置爬虫队列
	crawlerCap := a.CrawlerPool.ResetPoolSize(count)

	FKLog.G_Log.Informational(" *     执行任务总数(任务数[*自定义配置数])为 %v 个", count)
	FKLog.G_Log.Informational(" *     采集引擎池容量为 %v", crawlerCap)
	FKLog.G_Log.Informational(" *     并发协程最多 %v 个", a.AppRuntimeConfig.MaxThreadNum)
	FKLog.G_Log.Informational(" *     默认随机停顿 %v~%v 毫秒",
		a.AppRuntimeConfig.MedianPauseTime/2, a.AppRuntimeConfig.MedianPauseTime*2)
	FKLog.G_Log.App(" *  —— 开始抓取，请耐心等候 ——")
	FKLog.G_Log.Informational(` *********************************************************************************************************************************** `)

	// 开始计时
	FKStatus.GlobalAppStartTime = time.Now()

	// 根据模式选择合理的并发
	if a.AppRuntimeConfig.Mode == FKStatus.OFFLINE {
		// 可控制执行状态
		go a.goRun(count)
	} else {
		// 保证接收服务端任务的同步
		a.goRun(count)
	}
}

// 任务执行
func (a *App) goRun(count int) {
	// 执行任务
	var i int
	for i = 0; i < count && a.Status() != FKStatus.STOP; i++ {
	pause:
		if a.IsPausing() {
			time.Sleep(time.Second)
			goto pause
		}
		// 从爬行队列取出空闲蜘蛛，并发执行
		c := a.CrawlerPool.AllocCrawlerFromPool()
		if c != nil {
			go func(i int, c FKCrawler.Crawler) {
				// 执行并返回结果消息
				c.Init(a.SpiderQueue.GetByIndex(i)).Run()
				// 任务结束后回收该蜘蛛
				a.RWMutex.RLock()
				if a.status != FKStatus.STOP {
					a.CrawlerPool.ReturnBackCrawlerToPool(c)
				}
				a.RWMutex.RUnlock()
			}(i, c)
		}
	}
	// 监控结束任务
	for ii := 0; ii < i; ii++ {
		s := <-FKStatus.GlobalRuntimeReportChan
		if (s.DataNum == 0) && (s.FileNum == 0) {
			FKLog.G_Log.App(" *     [任务小计：%s | KEYWORDS：%s]   无采集结果，用时 %v",
				s.SpiderName, s.Keyword, s.Time)
			continue
		}
		FKLog.G_Log.Informational(" * ")
		switch {
		case s.DataNum > 0 && s.FileNum == 0:
			FKLog.G_Log.App(" *     [任务小计：%s | KEYWORDS：%s]   共采集数据 %v 条，用时 %v",
				s.SpiderName, s.Keyword, s.DataNum, s.Time)
		case s.DataNum == 0 && s.FileNum > 0:
			FKLog.G_Log.App(" *     [任务小计：%s | KEYWORDS：%s]   共下载文件 %v 个，用时 %v",
				s.SpiderName, s.Keyword, s.FileNum, s.Time)
		default:
			FKLog.G_Log.App(" *     [任务小计：%s | KEYWORDS：%s]   共采集数据 %v 条 + 下载文件 %v 个，用时 %v",
				s.SpiderName, s.Keyword, s.DataNum, s.FileNum, s.Time)
		}

		a.sum[0] += s.DataNum
		a.sum[1] += s.FileNum
	}

	// 总耗时
	a.takeTime = time.Since(FKStatus.GlobalAppStartTime)
	var prefix = func() string {
		if a.Status() == FKStatus.STOP {
			return "任务中途取消："
		}
		return "本次"
	}()
	// 打印总结报告
	FKLog.G_Log.Informational(" * ")
	FKLog.G_Log.Informational(` *********************************************************************************************************************************** `)
	FKLog.G_Log.Informational(" * ")
	switch {
	case a.sum[0] > 0 && a.sum[1] == 0:
		FKLog.G_Log.App(" *                            —— %s合计采集【数据 %v 条】， 实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, a.sum[0], FKStatus.GetSuccessRequestPageNum(),
			FKStatus.GetFailedRequestPageNum(), FKStatus.GetTotalRequestPageNum(), a.takeTime)
	case a.sum[0] == 0 && a.sum[1] > 0:
		FKLog.G_Log.App(" *                            —— %s合计采集【文件 %v 个】， 实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, a.sum[1], FKStatus.GetSuccessRequestPageNum(),
			FKStatus.GetFailedRequestPageNum(), FKStatus.GetTotalRequestPageNum(), a.takeTime)
	case a.sum[0] == 0 && a.sum[1] == 0:
		FKLog.G_Log.App(" *                            —— %s无采集结果，实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, FKStatus.GetSuccessRequestPageNum(),
			FKStatus.GetFailedRequestPageNum(), FKStatus.GetTotalRequestPageNum(), a.takeTime)
	default:
		FKLog.G_Log.App(" *                            —— %s合计采集【数据 %v 条 + 文件 %v 个】，实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, a.sum[0], a.sum[1], FKStatus.GetSuccessRequestPageNum(),
			FKStatus.GetFailedRequestPageNum(), FKStatus.GetTotalRequestPageNum(), a.takeTime)
	}
	FKLog.G_Log.Informational(" * ")
	FKLog.G_Log.Informational(` *********************************************************************************************************************************** `)

	// 单机模式并发运行，需要标记任务结束
	if a.AppRuntimeConfig.Mode == FKStatus.OFFLINE {
		a.PauseLog()
		a.syncOnceExitApp.Do(func() { close(a.finish) })
	}
}

// 客户端向服务端反馈日志
func (a *App) socketLog() {
	for a.canSocketLog {
		_, msg, ok := FKLog.G_Log.Peek()
		if !ok {
			return
		}
		if a.Teleport.CountNodes() == 0 {
			// 与服务器失去连接后，抛掉返馈日志
			continue
		}
		a.Teleport.Request(msg, "log", "")
	}
}

func (a *App) checkPort() bool {
	if a.AppRuntimeConfig.MasterPort == 0 {
		FKLog.G_Log.Warning(" *     —— 分布式端口不能为空")
		return false
	}
	return true
}

func (a *App) checkAll() bool {
	if a.AppRuntimeConfig.MasterIP == "" || !a.checkPort() {
		FKLog.G_Log.Warning(" *     —— 服务器地址不能为空")
		return false
	}
	return true
}

// 设置任务运行时公共配置
func (a *App) setAppConf(task *FKDistributor.DistributeTask) {
	a.AppRuntimeConfig.MaxThreadNum = task.ThreadNum
	a.AppRuntimeConfig.MedianPauseTime = task.Pausetime
	a.AppRuntimeConfig.OutputType = task.OutType
	a.AppRuntimeConfig.DockerCap = task.DockerCap
	a.AppRuntimeConfig.IsInheritSuccess = task.SuccessInherit
	a.AppRuntimeConfig.IsInheritFailure = task.FailureInherit
	a.AppRuntimeConfig.RequestLimit = task.Limit
	a.AppRuntimeConfig.UpdateProxyIntervale = task.ProxyMinute
	a.AppRuntimeConfig.Keywords = task.Keywords
}
func (a *App) setTask(task *FKDistributor.DistributeTask) {
	task.ThreadNum = a.AppRuntimeConfig.MaxThreadNum
	task.Pausetime = a.AppRuntimeConfig.MedianPauseTime
	task.OutType = a.AppRuntimeConfig.OutputType
	task.DockerCap = a.AppRuntimeConfig.DockerCap
	task.SuccessInherit = a.AppRuntimeConfig.IsInheritSuccess
	task.FailureInherit = a.AppRuntimeConfig.IsInheritFailure
	task.Limit = a.AppRuntimeConfig.RequestLimit
	task.ProxyMinute = a.AppRuntimeConfig.UpdateProxyIntervale
	task.Keywords = a.AppRuntimeConfig.Keywords
}
