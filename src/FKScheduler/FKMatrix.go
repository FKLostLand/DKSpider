package FKScheduler

import (
	"FKHistory"
	"FKLog"
	"FKRequest"
	"FKStatus"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// 一个蜘蛛实例的请求矩阵
type Matrix struct {
	maxPage            int64                         // 最大采集页数，以负数形式表示
	resCount           int32                         // 资源使用情况计数
	spiderName         string                        // 所属Spider名称
	requestPriorityMap map[int][]*FKRequest.Request  // [优先级]队列，优先级默认为0
	priorities         []int                         // 优先级顺序，从低到高
	history            FKHistory.Historier           // 历史记录
	tempHistory        map[string]bool               // 临时记录 [FKRequest.Unique(url+method)]true
	failures           map[string]*FKRequest.Request // 历史及本次失败请求
	tempHistoryLock    sync.RWMutex
	failureLock        sync.Mutex
	sync.Mutex
}

func createMatrix(spiderName, spiderSubName string, maxPage int64) *Matrix {
	matrix := &Matrix{
		spiderName:         spiderName,
		maxPage:            maxPage,
		requestPriorityMap: make(map[int][]*FKRequest.Request),
		priorities:         []int{},
		history:            FKHistory.CreateHistorier(spiderName, spiderSubName),
		tempHistory:        make(map[string]bool),
		failures:           make(map[string]*FKRequest.Request),
	}
	if FKStatus.GlobalRuntimeTaskConfig.Mode != FKStatus.SERVER {
		matrix.history.ReadSuccess(FKStatus.GlobalRuntimeTaskConfig.OutputType, FKStatus.GlobalRuntimeTaskConfig.IsInheritSuccess)
		matrix.history.ReadFailure(FKStatus.GlobalRuntimeTaskConfig.OutputType, FKStatus.GlobalRuntimeTaskConfig.IsInheritFailure)
		matrix.setFailures(matrix.history.PullFailureList())
	}
	return matrix
}

// 添加请求到队列，并发安全
func (m *Matrix) Push(req *FKRequest.Request) {
	// 禁止并发，降低请求积存量
	m.Lock()
	defer m.Unlock()

	if GlobalScheduler.checkStatus(FKStatus.STOP) {
		return
	}

	// 达到请求上限，停止该规则运行
	if m.maxPage >= 0 {
		return
	}

	// 暂停状态时等待，降低请求积存量
	waited := false
	for GlobalScheduler.checkStatus(FKStatus.PAUSE) {
		waited = true
		time.Sleep(time.Second)
	}
	if waited && GlobalScheduler.checkStatus(FKStatus.STOP) {
		return
	}

	// 资源使用过多时等待，降低请求积存量
	waited = false
	for atomic.LoadInt32(&m.resCount) > GlobalScheduler.avgRes() {
		waited = true
		time.Sleep(100 * time.Millisecond)
	}
	if waited && GlobalScheduler.checkStatus(FKStatus.STOP) {
		return
	}

	// 不可重复下载的req
	if !req.IsReloadable() {
		// 已存在成功记录时退出
		if m.hasHistory(req.Unique()) {
			return
		}
		// 添加到临时记录
		m.insertTempHistory(req.Unique())
	}

	var priority = req.GetPriority()

	// 初始化该蜘蛛下该优先级队列
	if _, found := m.requestPriorityMap[priority]; !found {
		m.priorities = append(m.priorities, priority)
		sort.Ints(m.priorities) // 从小到大排序
		m.requestPriorityMap[priority] = []*FKRequest.Request{}
	}

	// 添加请求到队列
	m.requestPriorityMap[priority] = append(m.requestPriorityMap[priority], req)

	// 大致限制加入队列的请求量，并发情况下应该会比maxPage多
	atomic.AddInt64(&m.maxPage, 1)
}

// 从队列取出请求，不存在时返回nil，并发安全
func (m *Matrix) Pull() (req *FKRequest.Request) {
	m.Lock()
	defer m.Unlock()
	if !GlobalScheduler.checkStatus(FKStatus.RUN) {
		return
	}
	// 按优先级从高到低取出请求
	for i := len(m.requestPriorityMap) - 1; i >= 0; i-- {
		idx := m.priorities[i]
		if len(m.requestPriorityMap[idx]) > 0 {
			req = m.requestPriorityMap[idx][0]
			m.requestPriorityMap[idx] = m.requestPriorityMap[idx][1:]
			if GlobalScheduler.useProxy {
				req.SetProxy(GlobalScheduler.proxy.GetOne(req.GetUrl()))
			} else {
				req.SetProxy("")
			}
			return
		}
	}
	return
}

func (m *Matrix) Use() {
	defer func() {
		recover()
	}()
	GlobalScheduler.count <- true
	atomic.AddInt32(&m.resCount, 1)
}

func (m *Matrix) Free() {
	<-GlobalScheduler.count
	atomic.AddInt32(&m.resCount, -1)
}

// 返回是否作为新的失败请求被添加至队列尾部
func (m *Matrix) DoHistory(req *FKRequest.Request, ok bool) bool {
	if !req.IsReloadable() {
		m.tempHistoryLock.Lock()
		delete(m.tempHistory, req.Unique())
		m.tempHistoryLock.Unlock()

		if ok {
			m.history.UpsertSuccess(req.Unique())
			return false
		}
	}

	if ok {
		return false
	}

	m.failureLock.Lock()
	defer m.failureLock.Unlock()
	if _, ok := m.failures[req.Unique()]; !ok {
		// 首次失败时，在任务队列末尾重新执行一次
		m.failures[req.Unique()] = req
		FKLog.G_Log.Informational(" *     + 失败请求: [%v]\n", req.GetUrl())
		return true
	}
	// 失败两次后，加入历史失败记录
	m.history.UpsertFailure(req)
	return false
}

func (m *Matrix) CanStop() bool {
	if GlobalScheduler.checkStatus(FKStatus.STOP) {
		return true
	}
	if m.maxPage >= 0 {
		return true
	}
	if atomic.LoadInt32(&m.resCount) != 0 {
		return false
	}
	if m.Len() > 0 {
		return false
	}

	m.failureLock.Lock()
	defer m.failureLock.Unlock()
	if len(m.failures) > 0 {
		// 重新下载历史记录中失败的请求
		var goon bool
		for reqUnique, req := range m.failures {
			if req == nil {
				continue
			}
			m.failures[reqUnique] = nil
			goon = true
			FKLog.G_Log.Informational(" *     - 失败请求: [%v]\n", req.GetUrl())
			m.Push(req)
		}
		if goon {
			return false
		}
	}
	return true
}

// 非服务器模式下保存历史成功记录
func (m *Matrix) TryFlushSuccess() {
	if FKStatus.GlobalRuntimeTaskConfig.Mode != FKStatus.SERVER && FKStatus.GlobalRuntimeTaskConfig.IsInheritSuccess {
		m.history.FlushSuccess(FKStatus.GlobalRuntimeTaskConfig.OutputType)
	}
}

// 非服务器模式下保存历史失败记录
func (m *Matrix) TryFlushFailure() {
	if FKStatus.GlobalRuntimeTaskConfig.Mode != FKStatus.SERVER && FKStatus.GlobalRuntimeTaskConfig.IsInheritFailure {
		m.history.FlushFailure(FKStatus.GlobalRuntimeTaskConfig.OutputType)
	}
}

// 等待处理中的请求完成
func (m *Matrix) Wait() {
	if GlobalScheduler.checkStatus(FKStatus.STOP) {
		return
	}
	for atomic.LoadInt32(&m.resCount) != 0 {
		time.Sleep(500 * time.Millisecond)
	}
}

func (m *Matrix) Len() int {
	m.Lock()
	defer m.Unlock()
	var l int
	for _, reqs := range m.requestPriorityMap {
		l += len(reqs)
	}
	return l
}

func (m *Matrix) hasHistory(reqUnique string) bool {
	if m.history.HasSuccess(reqUnique) {
		return true
	}
	m.tempHistoryLock.RLock()
	has := m.tempHistory[reqUnique]
	m.tempHistoryLock.RUnlock()
	return has
}

func (m *Matrix) insertTempHistory(reqUnique string) {
	m.tempHistoryLock.Lock()
	m.tempHistory[reqUnique] = true
	m.tempHistoryLock.Unlock()
}

func (m *Matrix) setFailures(reqs map[string]*FKRequest.Request) {
	m.failureLock.Lock()
	defer m.failureLock.Unlock()
	for key, req := range reqs {
		m.failures[key] = req
		FKLog.G_Log.Informational(" *     * 失败请求: [%v]\n", req.GetUrl())
	}
}
