package FKScheduler

import (
	"FKLog"
	"FKProxy"
	"FKStatus"
	"sync"
	"time"
)

// 调度器
type scheduler struct {
	status       int            // 运行状态
	count        chan bool      // 总并发量计数
	useProxy     bool           // 标记是否使用代理IP
	proxy        *FKProxy.Proxy // 全局代理IP
	matrices     []*Matrix      // Spider实例的请求矩阵列表
	sync.RWMutex                // 全局读写锁
}

// 定义全局调度器
var GlobalScheduler = &scheduler{
	status: FKStatus.RUN,
	count:  make(chan bool, FKStatus.GlobalRuntimeTaskConfig.MaxThreadNum),
	proxy:  FKProxy.CreateProxy(),
}

func Init() {
	for GlobalScheduler.proxy == nil {
		time.Sleep(100 * time.Millisecond)
	}
	GlobalScheduler.matrices = []*Matrix{}
	GlobalScheduler.count = make(chan bool, FKStatus.GlobalRuntimeTaskConfig.MaxThreadNum)

	if FKStatus.GlobalRuntimeTaskConfig.UpdateProxyIntervale > 0 {
		if GlobalScheduler.proxy.Count() > 0 {
			GlobalScheduler.useProxy = true
			GlobalScheduler.proxy.UpdateTicker(FKStatus.GlobalRuntimeTaskConfig.UpdateProxyIntervale)
			FKLog.G_Log.Informational(" *     使用代理IP，代理IP更换频率为 %v 分钟", FKStatus.GlobalRuntimeTaskConfig.UpdateProxyIntervale)
		} else {
			GlobalScheduler.useProxy = false
			FKLog.G_Log.Informational(" *     在线代理IP列表为空，无法使用代理IP")
		}
	} else {
		GlobalScheduler.useProxy = false
		FKLog.G_Log.Informational(" *     不使用代理IP")
	}

	GlobalScheduler.status = FKStatus.RUN
}

// 注册资源队列
func AddMatrix(spiderName, spiderSubName string, maxPage int64) *Matrix {
	matrix := createMatrix(spiderName, spiderSubName, maxPage)
	GlobalScheduler.RLock()
	defer GlobalScheduler.RUnlock()
	GlobalScheduler.matrices = append(GlobalScheduler.matrices, matrix)
	return matrix
}

// 暂停\恢复所有爬行任务
func PauseRecover() {
	GlobalScheduler.Lock()
	defer GlobalScheduler.Unlock()
	switch GlobalScheduler.status {
	case FKStatus.PAUSE:
		GlobalScheduler.status = FKStatus.RUN
	case FKStatus.RUN:
		GlobalScheduler.status = FKStatus.PAUSE
	}
}

// 终止任务
func Stop() {
	GlobalScheduler.Lock()
	defer GlobalScheduler.Unlock()

	GlobalScheduler.status = FKStatus.STOP
	// 清空
	defer func() {
		recover()
	}()
	// for _, matrix := range sdl.matrices {
	// 	matrix.windup()
	// }
	close(GlobalScheduler.count)
	GlobalScheduler.matrices = []*Matrix{}
}

// 每个spider实例分配到的平均资源量
func (s *scheduler) avgRes() int32 {
	avg := int32(cap(GlobalScheduler.count) / len(GlobalScheduler.matrices))
	if avg == 0 {
		avg = 1
	}
	return avg
}

func (s *scheduler) checkStatus(status int) bool {
	s.RLock()
	b := s.status == status
	s.RUnlock()
	return b
}
