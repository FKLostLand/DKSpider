package FKSpider

import (
	"FKLog"
	"FKRequest"
	"FKScheduler"
	"FKStatus"
	"FKTimer"
	"math"
	"sync"
	"time"
)

type (
	//采集规则树
	RuleTree struct {
		Root  func(*Context)   // 根节点(执行入口)
		Trunk map[string]*Rule // 节点散列表(执行采集过程)
	}
	// 采集规则节点
	Rule struct {
		ItemFields []string                                           // 结果字段列表(选填，写上可保证字段顺序)
		ParseFunc  func(*Context)                                     // 内容解析函数
		AidFunc    func(*Context, map[string]interface{}) interface{} // 通用辅助函数
	}
	// 蜘蛛对象
	Spider struct {
		Name            string                                                  // 蜘蛛名称
		Description     string                                                  // 蜘蛛描述文字
		MedianPauseTime int64                                                   // 暂停时间中位数。优先使用配置参数，其次使用界面参数。
		RequestLimit    int64                                                   // 请求数限制。0为不限。优先使用配置参数，其次使用界面参数。
		Keywords        string                                                  // 自定义关键字，使用前必须设置规则初始值为  KEYWORDS
		EnableCookie    bool                                                    // 是否使用cookie记录
		NotDefaultField bool                                                    // 是否向输出结果中添加默认字段（例如url, downloadTime等）
		Namespace       func(s *Spider) string                                  // 命名空间，用来输出文件，路径命名
		SubNamespace    func(s *Spider, dataCell map[string]interface{}) string // 次级命名空间，用来输出文件，路径命名
		RuleTree        *RuleTree                                               // 采集规则树

		// 以下字段系统自动赋值
		id        int                 // 自动分配的SpiderQueue中的索引
		subName   string              // 由Keywords转换为的二级标识名
		reqMatrix *FKScheduler.Matrix // 请求矩阵
		timer     *FKTimer.Timer      // 定时器
		status    int                 // 执行状态
		lock      sync.RWMutex
		once      sync.Once
	}
)

// 添加自身到蜘蛛菜单
func (s Spider) RegisterToSpiderSpecies() *Spider {
	s.status = FKStatus.UNINIT
	return GlobalSpiderSpecies.Add(&s)
}

// 指定规则的获取结果的字段名列表
func (s *Spider) GetItemFields(rule *Rule) []string {
	return rule.ItemFields
}

// 返回结果字段名的值
// 不存在时返回空字符串
func (s *Spider) GetItemField(rule *Rule, index int) (field string) {
	if index > len(rule.ItemFields)-1 || index < 0 {
		return ""
	}
	return rule.ItemFields[index]
}

// 返回结果字段名的其索引
// 不存在时索引为-1
func (s *Spider) GetItemFieldIndex(rule *Rule, field string) (index int) {
	for idx, v := range rule.ItemFields {
		if v == field {
			return idx
		}
	}
	return -1
}

// 为指定Rule动态追加结果字段名，并返回索引位置
// 已存在时返回原来索引位置
func (s *Spider) UpsertItemField(rule *Rule, field string) (index int) {
	for i, v := range rule.ItemFields {
		if v == field {
			return i
		}
	}
	rule.ItemFields = append(rule.ItemFields, field)
	return len(rule.ItemFields) - 1
}

// 获取蜘蛛名称
func (s *Spider) GetName() string {
	return s.Name
}

// 获取蜘蛛二级标识名
func (s *Spider) GetSubName() string {
	s.once.Do(func() {
		s.subName = s.GetKeywords()
		// s.subName = FKBase.String2Hash(s.subName)
	})
	return s.subName
}

// 安全返回指定规则
func (s *Spider) GetRule(ruleName string) (*Rule, bool) {
	rule, found := s.RuleTree.Trunk[ruleName]
	return rule, found
}

// 返回指定规则
func (s *Spider) MustGetRule(ruleName string) *Rule {
	return s.RuleTree.Trunk[ruleName]
}

// 返回规则树
func (s *Spider) GetRules() map[string]*Rule {
	return s.RuleTree.Trunk
}

// 获取蜘蛛描述
func (s *Spider) GetDescription() string {
	return s.Description
}

// 获取蜘蛛ID
func (s *Spider) GetId() int {
	return s.id
}

// 设置蜘蛛ID
func (s *Spider) SetId(id int) {
	s.id = id
}

// 获取自定义配置信息
func (s *Spider) GetKeywords() string {
	return s.Keywords
}

// 设置自定义配置信息
func (s *Spider) SetKeywords(keywords string) {
	s.Keywords = keywords
}

// 获取采集上限
// <0 表示采用限制请求数的方案
// >0 表示采用规则中的自定义限制方案
func (s *Spider) GetLimit() int64 {
	return s.RequestLimit
}

// 设置采集上限
// <0 表示采用限制请求数的方案
// >0 表示采用规则中的自定义限制方案
func (s *Spider) SetLimit(max int64) {
	s.RequestLimit = max
}

// 控制所有请求是否使用cookie
func (s *Spider) GetEnableCookie() bool {
	return s.EnableCookie
}

// 自定义暂停时间 pause[0]~(pause[0]+pause[1])，优先级高于外部传参
// 当且仅当runtime[0]为true时可覆盖现有值
func (s *Spider) SetPausetime(pause int64, runtime ...bool) {
	if s.MedianPauseTime == 0 || len(runtime) > 0 && runtime[0] {
		s.MedianPauseTime = pause
	}
}

// 设置定时器
// @id为定时器唯一标识
// @bell==nil时为倒计时器，此时@tol为睡眠时长
// @bell!=nil时为闹铃，此时@tol用于指定醒来时刻（从now起遇到的第tol个bell）
func (s *Spider) SetTimer(id string, tol time.Duration, bell *FKTimer.Bell) bool {
	if s.timer == nil {
		s.timer = FKTimer.CreateTimer()
	}
	return s.timer.Set(id, tol, bell)
}

// 启动定时器，并返回定时器是否可以继续使用
func (s *Spider) RunTimer(id string) bool {
	if s.timer == nil {
		return false
	}
	return s.timer.Sleep(id)
}

// 返回一个自身复制品
func (s *Spider) Copy() *Spider {
	ghost := &Spider{}
	ghost.Name = s.Name
	ghost.subName = s.subName

	ghost.RuleTree = &RuleTree{
		Root:  s.RuleTree.Root,
		Trunk: make(map[string]*Rule, len(s.RuleTree.Trunk)),
	}
	for k, v := range s.RuleTree.Trunk {
		ghost.RuleTree.Trunk[k] = new(Rule)

		ghost.RuleTree.Trunk[k].ItemFields = make([]string, len(v.ItemFields))
		copy(ghost.RuleTree.Trunk[k].ItemFields, v.ItemFields)

		ghost.RuleTree.Trunk[k].ParseFunc = v.ParseFunc
		ghost.RuleTree.Trunk[k].AidFunc = v.AidFunc
	}

	ghost.Description = s.Description
	ghost.MedianPauseTime = s.MedianPauseTime
	ghost.EnableCookie = s.EnableCookie
	ghost.RequestLimit = s.RequestLimit
	ghost.Keywords = s.Keywords

	ghost.NotDefaultField = s.NotDefaultField
	ghost.Namespace = s.Namespace
	ghost.SubNamespace = s.SubNamespace

	ghost.timer = s.timer
	ghost.status = s.status

	return ghost
}

func (s *Spider) ReqmatrixInit() *Spider {
	if s.RequestLimit < 0 {
		s.reqMatrix = FKScheduler.AddMatrix(s.GetName(), s.GetSubName(), s.RequestLimit)
		s.SetLimit(0)
	} else {
		s.reqMatrix = FKScheduler.AddMatrix(s.GetName(), s.GetSubName(), math.MinInt64)
	}
	return s
}

// 返回是否作为新的失败请求被添加至队列尾部
func (s *Spider) DoHistory(req *FKRequest.Request, ok bool) bool {
	return s.reqMatrix.DoHistory(req, ok)
}

func (s *Spider) RequestPush(req *FKRequest.Request) {
	s.reqMatrix.Push(req)
}

func (s *Spider) RequestPull() *FKRequest.Request {
	return s.reqMatrix.Pull()
}

func (s *Spider) RequestUse() {
	s.reqMatrix.Use()
}

func (s *Spider) RequestFree() {
	s.reqMatrix.Free()
}

func (s *Spider) RequestLen() int {
	return s.reqMatrix.Len()
}

func (s *Spider) TryFlushSuccess() {
	s.reqMatrix.TryFlushSuccess()
}

func (s *Spider) TryFlushFailure() {
	s.reqMatrix.TryFlushFailure()
}

// 开始执行蜘蛛
func (s *Spider) Start() {
	defer func() {
		if p := recover(); p != nil {
			FKLog.G_Log.Error(" *     Panic  [root]: %v\n", p)
		}
		s.lock.Lock()
		s.status = FKStatus.RUN
		s.lock.Unlock()
	}()
	s.RuleTree.Root(GetContext(s, nil))
}

// 主动崩溃爬虫运行协程
func (s *Spider) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.status == FKStatus.STOP {
		return
	}
	s.status = FKStatus.STOP
	// 取消所有定时器
	if s.timer != nil {
		s.timer.Drop()
		s.timer = nil
	}
}

func (s *Spider) CanStop() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.status != FKStatus.UNINIT && s.reqMatrix.CanStop()
}

func (s *Spider) IsStopping() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.status == FKStatus.STOP
}

// 若已主动终止任务，则崩溃爬虫协程
func (s *Spider) tryPanic() {
	if s.IsStopping() {
		panic("Proactive stop sprider.")
	}
}

// 退出任务前收尾工作
func (s *Spider) Defer() {
	// 取消所有定时器
	if s.timer != nil {
		s.timer.Drop()
		s.timer = nil
	}
	// 等待处理中的请求完成
	s.reqMatrix.Wait()
	// 更新失败记录
	s.reqMatrix.TryFlushFailure()
}

// 是否输出默认添加的字段 Url/ParentUrl/DownloadTime
func (s *Spider) OutDefaultField() bool {
	return !s.NotDefaultField
}
