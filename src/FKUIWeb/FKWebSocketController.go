package FKUIWeb

import (
	"FKApp"
	"FKBase"
	"FKConfig"
	"FKSpider"
	"FKStatus"
	ws "github.com/golang/net/websocket"
	"math"
	"sync"
)

type SocketController struct {
	connPool     map[string]*ws.Conn
	wchanPool    map[string]*Wchan
	connRWMutex  sync.RWMutex
	wchanRWMutex sync.RWMutex
}

func (self *SocketController) GetConn(sessID string) *ws.Conn {
	self.connRWMutex.RLock()
	defer self.connRWMutex.RUnlock()
	return self.connPool[sessID]
}

func (self *SocketController) GetWchan(sessID string) *Wchan {
	self.wchanRWMutex.RLock()
	defer self.wchanRWMutex.RUnlock()
	return self.wchanPool[sessID]
}

func (self *SocketController) Add(sessID string, conn *ws.Conn) {
	self.connRWMutex.Lock()
	self.wchanRWMutex.Lock()
	defer self.connRWMutex.Unlock()
	defer self.wchanRWMutex.Unlock()

	self.connPool[sessID] = conn
	self.wchanPool[sessID] = newWchan()
}

func (self *SocketController) Remove(sessID string, conn *ws.Conn) {
	self.connRWMutex.Lock()
	self.wchanRWMutex.Lock()
	defer self.connRWMutex.Unlock()
	defer self.wchanRWMutex.Unlock()

	if self.connPool[sessID] == nil {
		return
	}
	wc := self.wchanPool[sessID]
	close(wc.wchan)
	conn.Close()
	delete(self.connPool, sessID)
	delete(self.wchanPool, sessID)
}

func (self *SocketController) Write(sessID string, void map[string]interface{}, to ...int) {
	self.wchanRWMutex.RLock()
	defer self.wchanRWMutex.RUnlock()

	// to为1时，只向当前连接发送；to为-1时，向除当前连接外的其他所有连接发送；to为0时或为空时，向所有连接发送
	var t int = 0
	if len(to) > 0 {
		t = to[0]
	}

	void["mode"] = FKApp.G_App.GetAppConfig("Mode").(int)

	switch t {
	case 1:
		wc := self.wchanPool[sessID]
		if wc == nil {
			return
		}
		void["initiative"] = true
		wc.wchan <- void

	case 0, -1:
		l := len(self.wchanPool)
		for _sessID, wc := range self.wchanPool {
			if t == -1 && _sessID == sessID {
				continue
			}
			_void := make(map[string]interface{}, l)
			for k, v := range void {
				_void[k] = v
			}
			if _sessID == sessID {
				_void["initiative"] = true
			} else {
				_void["initiative"] = false
			}
			wc.wchan <- _void
		}
	}
}

type Wchan struct {
	wchan chan interface{}
}

func newWchan() *Wchan {
	return &Wchan{
		wchan: make(chan interface{}, 1024),
	}
}

var (
	wsApi = map[string]func(string, map[string]interface{}){}
	Sc    = &SocketController{
		connPool:  make(map[string]*ws.Conn),
		wchanPool: make(map[string]*Wchan),
	}
)

func init() {
	// 初始化运行
	wsApi["refresh"] = func(sessID string, req map[string]interface{}) {
		// 写入发送通道
		Sc.Write(sessID, tplData(FKApp.G_App.GetAppConfig("Mode").(int)), 1)
	}

	// 初始化运行
	wsApi["init"] = func(sessID string, req map[string]interface{}) {
		var mode = FKBase.Atoi(req["mode"])
		var port = FKBase.Atoi(req["port"])
		var master = FKBase.Atoa(req["ip"]) //服务器(主节点)地址，不含端口
		currMode := FKApp.G_App.GetAppConfig("Mode").(int)
		if currMode == FKStatus.UNSET {
			FKApp.G_App.Init(mode, port, master, G_LSC) // 运行模式初始化，设置log输出目标
		} else {
			FKApp.G_App = FKApp.G_App.ReInit(mode, port, master) // 切换运行模式
		}

		if mode == FKStatus.CLIENT {
			go FKApp.G_App.Run()
		}

		// 写入发送通道
		Sc.Write(sessID, tplData(mode))
	}

	wsApi["run"] = func(sessID string, req map[string]interface{}) {
		if FKApp.G_App.GetAppConfig("Mode").(int) != FKStatus.CLIENT {
			setConf(req)
		}

		if FKApp.G_App.GetAppConfig("Mode").(int) == FKStatus.OFFLINE {
			Sc.Write(sessID, map[string]interface{}{"operate": "run"})
		}

		go func() {
			FKApp.G_App.Run()
			if FKApp.G_App.GetAppConfig("Mode").(int) == FKStatus.OFFLINE {
				Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
			}
		}()
	}

	// 终止当前任务，现仅支持单机模式
	wsApi["stop"] = func(sessID string, req map[string]interface{}) {
		if FKApp.G_App.GetAppConfig("Mode").(int) != FKStatus.OFFLINE {
			Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
			return
		} else {
			FKApp.G_App.Stop()
			Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
		}
	}

	// 任务暂停与恢复，目前仅支持单机模式
	wsApi["pauseRecover"] = func(sessID string, req map[string]interface{}) {
		if FKApp.G_App.GetAppConfig("Mode").(int) != FKStatus.OFFLINE {
			return
		}
		FKApp.G_App.PauseRecover()
		Sc.Write(sessID, map[string]interface{}{"operate": "pauseRecover"})
	}

	// 退出当前模式
	wsApi["exit"] = func(sessID string, req map[string]interface{}) {
		FKApp.G_App = FKApp.G_App.ReInit(FKStatus.UNSET, 0, "")
		Sc.Write(sessID, map[string]interface{}{"operate": "exit"})
	}
}

func tplData(mode int) map[string]interface{} {
	var info = map[string]interface{}{"operate": "init", "mode": mode}

	// 运行模式标题
	switch mode {
	case FKStatus.OFFLINE:
		info["title"] = FKConfig.APP_FULL_NAME + "                                                          【 运行模式 ->  单机 】"
	case FKStatus.SERVER:
		info["title"] = FKConfig.APP_FULL_NAME + "                                                          【 运行模式 ->  服务端 】"
	case FKStatus.CLIENT:
		info["title"] = FKConfig.APP_FULL_NAME + "                                                          【 运行模式 ->  客户端 】"
	}

	if mode == FKStatus.CLIENT {
		return info
	}

	// 蜘蛛家族清单
	info["spiders"] = map[string]interface{}{
		"menu": spiderMenu,
		"curr": func() interface{} {
			l := FKApp.G_App.GetSpiderQueue().Len()
			if l == 0 {
				return 0
			}
			var curr = make(map[string]bool, l)
			for _, sp := range FKApp.G_App.GetSpiderQueue().GetAll() {
				curr[sp.GetName()] = true
			}

			return curr
		}(),
	}

	// 输出方式清单
	info["OutType"] = map[string]interface{}{
		"menu": FKApp.G_App.GetOutputTypeList(),
		"curr": FKApp.G_App.GetAppConfig("OutputType"),
	}

	// 并发协程上限
	info["ThreadNum"] = map[string]int{
		"max":  999999,
		"min":  1,
		"curr": FKApp.G_App.GetAppConfig("MaxThreadNum").(int),
	}

	// 暂停区间/ms(随机: Pausetime/2 ~ Pausetime*2)
	info["Pausetime"] = map[string][]int64{
		"menu": {0, 100, 300, 500, 1000, 3000, 5000, 10000, 15000, 20000, 30000, 60000},
		"curr": []int64{FKApp.G_App.GetAppConfig("MedianPauseTime").(int64)},
	}

	// 代理IP更换的间隔分钟数
	info["ProxyMinute"] = map[string][]int64{
		"menu": {0, 1, 3, 5, 10, 15, 20, 30, 45, 60, 120, 180},
		"curr": []int64{FKApp.G_App.GetAppConfig("UpdateProxyIntervale").(int64)},
	}

	// 分批输出的容量
	info["DockerCap"] = map[string]int{
		"min":  1,
		"max":  5000000,
		"curr": FKApp.G_App.GetAppConfig("DockerCap").(int),
	}

	// 采集上限
	if FKApp.G_App.GetAppConfig("RequestLimit").(int64) == math.MaxInt64 {
		info["RequestLimit"] = 0
	} else {
		info["RequestLimit"] = FKApp.G_App.GetAppConfig("RequestLimit")
	}

	// 自定义配置
	info["Keywords"] = FKApp.G_App.GetAppConfig("Keywords")

	// 继承历史记录
	info["IsInheritSuccess"] = FKApp.G_App.GetAppConfig("IsInheritSuccess")
	info["IsInheritFailure"] = FKApp.G_App.GetAppConfig("IsInheritFailure")

	// 运行状态
	info["status"] = FKApp.G_App.Status()

	return info
}

// 配置运行参数
func setConf(req map[string]interface{}) {
	if tn := FKBase.Atoi(req["ThreadNum"]); tn == 0 {
		FKApp.G_App.SetAppConfig("MaxThreadNum", 1)
	} else {
		FKApp.G_App.SetAppConfig("MaxThreadNum", tn)
	}

	FKApp.G_App.
		SetAppConfig("MedianPauseTime", int64(FKBase.Atoi(req["Pausetime"]))).
		SetAppConfig("UpdateProxyIntervale", int64(FKBase.Atoi(req["ProxyMinute"]))).
		SetAppConfig("OutputType", FKBase.Atoa(req["OutType"])).
		SetAppConfig("DockerCap", FKBase.Atoi(req["DockerCap"])).
		SetAppConfig("RequestLimit", int64(FKBase.Atoi(req["Limit"]))).
		SetAppConfig("Keywords", FKBase.Atoa(req["Keywords"])).
		SetAppConfig("IsInheritSuccess", req["SuccessInherit"] == "true").
		SetAppConfig("IsInheritFailure", req["FailureInherit"] == "true")

	setSpiderQueue(req)
}

func setSpiderQueue(req map[string]interface{}) {
	spNames, ok := req["spiders"].([]interface{})
	if !ok {
		return
	}
	spiders := []*FKSpider.Spider{}
	for _, sp := range FKApp.G_App.GetSpiderTypeList() {
		for _, spName := range spNames {
			if FKBase.Atoa(spName) == sp.GetName() {
				spiders = append(spiders, sp.Copy())
			}
		}
	}
	FKApp.G_App.SpiderPrepare(spiders)
}
