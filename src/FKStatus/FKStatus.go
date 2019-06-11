package FKStatus

import (
	"sync/atomic"
	"time"
)

// 本APP运行模式
const (
	UNSET int = iota - 1
	OFFLINE
	SERVER
	CLIENT
)

// 运行状态
const (
	UNINIT = iota - 1
	STOP
	RUN
	PAUSE
)

// 本APP运行时配置
type AppRuntimeConfig struct {
	Mode                 int    // 本节点角色
	MasterPort           int    // 主节点端口
	MasterIP             string // 主节点IP
	MaxThreadNum         int    // 全局最大并发量
	MedianPauseTime      int64  // 暂停时长(ms)
	OutputType           string // 文本输出方式
	DockerCap            int    // 分段转储容器容量
	RequestLimit         int64  // 采集数量上限
	UpdateProxyIntervale int64  // 更换代理IP间隔时间(m)
	IsInheritSuccess     bool   // 是否继承历史成功纪录
	IsInheritFailure     bool   // 是否继承历史失败纪录
	Keywords             string // 关键字列表
}

// 本APP任务报告
type AppRuntimeReport struct {
	SpiderName string        // 蜘蛛名
	Keyword    string        // 关键字
	DataNum    uint64        // 数据量
	FileNum    uint64        // 文件量
	Time       time.Duration // 报告时间
}

var (
	GlobalRuntimeTaskConfig *AppRuntimeConfig      // 当前运行的任务状态
	GlobalAppStartTime      time.Time              // 开始运行App的时间
	GlobalRuntimeReportChan chan *AppRuntimeReport // 全局监控数据报告
	SuccessRequestPageNum   uint64                 // 成功的请求页面数量
	FailedRequestPageNum    uint64                 // 失败的页面请求数量
)

// 重置页面请求计数
func ResetRequestPageNum() {
	SuccessRequestPageNum = 0
	FailedRequestPageNum = 0
}

func GetSuccessRequestPageNum() uint64 {
	return SuccessRequestPageNum
}
func GetFailedRequestPageNum() uint64 {
	return FailedRequestPageNum
}
func GetTotalRequestPageNum() uint64 {
	return SuccessRequestPageNum + FailedRequestPageNum
}

// 添加成功的请求页面数量
func AddSuccessRequestPageNum() {
	atomic.AddUint64(&SuccessRequestPageNum, 1)
}

// 添加失败的请求页面数量
func AddFailedRequestPageNum() {
	atomic.AddUint64(&FailedRequestPageNum, 1)
}

// 入口
func init() {
	GlobalRuntimeTaskConfig = new(AppRuntimeConfig)
	// 任务报告
	GlobalRuntimeReportChan = make(chan *AppRuntimeReport)
}
