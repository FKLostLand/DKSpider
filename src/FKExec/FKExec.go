package exec

import (
	"FKApp"
	"FKConfig"
	"FKGc"
	"FKPipeline"
	"FKStatus"
	"FKSystem"
	"flag"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"FKAntiDebug"
	"FKRegistionCheck"
)

var (
	flagUIType               *string
	flagNodeMode             *int
	flagServerIP             *string
	flagServerPort           *int
	flagKeywords             *string
	flagRequestLimit         *int64
	flagOutputType           *string
	flagMaxThreadNum         *int
	flagMedianPauseTime      *int64
	flagUpdateProxyIntervale *int64
	flagDockerCap            *int
	flagIsInheritSuccess     *bool
	flagIsInheritFailure     *bool
)

func init() {
	// 开启最大CPU核心执行
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 开启手动GC协程
	FKGc.StartManualGCThread()
}


func Init(isDevMode bool) {
	// 反DEBUG嵌入
	if !isDevMode {
		FKAntiDebug.AntiCheckSleep()
		FKAntiDebug.AntiDebugCheck()
		FKAntiDebug.AntiVirus()
	}

	// 注册检查
	FKRegistionCheck.RegistionCheck()

	// 输出功能检查
	for out := range FKPipeline.G_DataOutput {
		FKPipeline.G_DataOutputLib = append(FKPipeline.G_DataOutputLib, out)
	}
	sort.Strings(FKPipeline.G_DataOutputLib)
}

func Run() {
	fmt.Printf("%v\n\n", FKConfig.APP_FULL_NAME)
	readFlag()
	parseDifferentUIFlag()

	flag.Parse()
	writeFlag()

	runPlatform(*flagUIType)
}

// 读取命令行参数
func readFlag() {
	OSType := FKSystem.GetSystemOSType()
	defaultUIType := "cmd"
	switch OSType {
	case FKSystem.EnumSystemLinux:
		defaultUIType = "cmd"
		break
	case FKSystem.EnumSystemWindows:
		defaultUIType = "gui"
		break
	case FKSystem.EnumSystemMacOS:
		defaultUIType = "web"
		break
	default:
		break
	}
	flagUIType = flag.String("UIType", defaultUIType,
		"   <选择UI操作界面> [web] 网页版UI    [gui] 软件版UI    [cmd] 命令行UI")
	flagNodeMode = flag.Int("NodeMode", FKStatus.GlobalRuntimeTaskConfig.Mode,
		"   <选择本节点运行模式: ["+strconv.Itoa(FKStatus.OFFLINE)+"] 单机模式    ["+strconv.Itoa(FKStatus.SERVER)+"] 服务端节点    ["+strconv.Itoa(FKStatus.CLIENT)+"] 客户端节点>")
	flagServerIP = flag.String("MasterIP", FKStatus.GlobalRuntimeTaskConfig.MasterIP,
		"   <服务端IP: 不含端口，仅客户端模式下使用>")
	flagServerPort = flag.Int("MasterPort", FKStatus.GlobalRuntimeTaskConfig.MasterPort,
		"   <端口号: 只填写数字即可，不含冒号，单机模式无需填写>")
	flagKeywords = flag.String("Keywords", FKStatus.GlobalRuntimeTaskConfig.Keywords,
		"   <自定义关键字配置: 多关键字请分别多包一层“<>”>")
	flagRequestLimit = flag.Int64("RequestLimit", FKStatus.GlobalRuntimeTaskConfig.RequestLimit,
		"   <总计采集数量上限（最大限制URL数）> [>=0]")
	flagOutputType = flag.String("OutputType", FKStatus.GlobalRuntimeTaskConfig.OutputType,
		func() string {
			var tmp string
			for _, v := range FKApp.G_App.GetOutputTypeList() {
				tmp += "[" + v + "] "
			}
			return "   <文本输出方式: > " + strings.TrimRight(tmp, " ")
		}())
	flagMaxThreadNum = flag.Int("MaxThreadNum", FKStatus.GlobalRuntimeTaskConfig.MaxThreadNum,
		"   <同时并发协程数量> [1~99999]")
	flagMedianPauseTime = flag.Int64("MedianPauseTime", FKStatus.GlobalRuntimeTaskConfig.MedianPauseTime,
		"   <平均间隔暂停时间（毫秒）> [>=100]")
	flagUpdateProxyIntervale = flag.Int64("UpdateProxyIntervale", FKStatus.GlobalRuntimeTaskConfig.UpdateProxyIntervale,
		"   <代理IP更换频率（分），为0时表示不使用代理> [>=0]")
	flagDockerCap = flag.Int("DockerCap", FKStatus.GlobalRuntimeTaskConfig.DockerCap,
		"   <文件转储容器容量> [1~5000000]")
	flagIsInheritSuccess = flag.Bool("IsInheritSuccess", FKStatus.GlobalRuntimeTaskConfig.IsInheritSuccess,
		"   <是否继承并保存成功记录> [true] [false]")
	flagIsInheritFailure = flag.Bool("IsInheritFailure", FKStatus.GlobalRuntimeTaskConfig.IsInheritFailure,
		"   <是否继承并保存失败记录> [true] [false]")
}

// 写入命令行参数
func 	writeFlag() {
	FKStatus.GlobalRuntimeTaskConfig.Mode = *flagNodeMode
	FKStatus.GlobalRuntimeTaskConfig.MasterPort = *flagServerPort
	FKStatus.GlobalRuntimeTaskConfig.MasterIP = *flagServerIP
	FKStatus.GlobalRuntimeTaskConfig.MaxThreadNum = *flagMaxThreadNum
	FKStatus.GlobalRuntimeTaskConfig.MedianPauseTime = *flagMedianPauseTime
	FKStatus.GlobalRuntimeTaskConfig.OutputType = *flagOutputType
	FKStatus.GlobalRuntimeTaskConfig.DockerCap = *flagDockerCap
	FKStatus.GlobalRuntimeTaskConfig.RequestLimit = *flagRequestLimit
	FKStatus.GlobalRuntimeTaskConfig.UpdateProxyIntervale = *flagUpdateProxyIntervale
	FKStatus.GlobalRuntimeTaskConfig.IsInheritSuccess = *flagIsInheritSuccess
	FKStatus.GlobalRuntimeTaskConfig.IsInheritFailure = *flagIsInheritFailure
	FKStatus.GlobalRuntimeTaskConfig.Keywords = *flagKeywords
}
