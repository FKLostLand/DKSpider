package FKUICmd

import (
	"FKApp"
	"FKLog"
	"FKSpider"
	"FKStatus"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var (
	flagSpiders *string
)

// 参数解析
func ParseFlag() {
	flagSpiders = flag.String(
		"Spiders",
		"",
		func() string {
			var spiderTypelist string
			for k, v := range FKApp.G_App.GetSpiderTypeList() {
				spiderTypelist += "   [" + strconv.Itoa(k) + "] " + v.GetName() + "  " + v.GetDescription() + "\r\n"
			}
			return "   <蜘蛛列表: 选择多蜘蛛以 \",\" 间隔>\r\n" + spiderTypelist
		}())
}

func showUsageCommit() {
	fmt.Println(
		"用法Example: FKAutoSpiderPool_windows.exe -UIType=cmd -NodeMode=0 -Spiders=1,2 -OutputType=csv -MaxThreadNum=20 -DockerCap=10000 -MedianPauseTime=300 -RequestLimit=0 -UpdateProxyIntervale=0 Keywords=\"\" IsInheritSuccess=true IsInheritFailure=true")
	fmt.Println(
		"用法Example: FKAutoSpiderPool_windows.exe -UIType=web")
	fmt.Println(
		"用法Example: FKAutoSpiderPool_windows.exe -UIType=gui")
}

// 执行入口
func Main() {
	fmt.Println("正在运行命令行模式...")
	FKApp.G_App.Init(FKStatus.GlobalRuntimeTaskConfig.Mode,
		FKStatus.GlobalRuntimeTaskConfig.MasterPort, FKStatus.GlobalRuntimeTaskConfig.MasterIP)
	if FKStatus.GlobalRuntimeTaskConfig.Mode == FKStatus.UNSET {
		// 展示用法说明
		showUsageCommit()
		return
	}
	switch FKApp.G_App.GetAppConfig("Mode").(int) {
	case FKStatus.SERVER:
		for {
			parseInput()
			run()
		}
	case FKStatus.CLIENT:
		run()
		select {}
	default:
		run()
	}
}

// 运行
func run() {
	// 创建蜘蛛队列
	sps := []*FKSpider.Spider{}
	*flagSpiders = strings.TrimSpace(*flagSpiders)
	if *flagSpiders == "*" {
		sps = FKApp.G_App.GetSpiderTypeList()
		for i, s := range FKApp.G_App.GetSpiderTypeList(){
			FKLog.G_Log.Informational("使用蜘蛛：%d - %s", i, s.Name)
		}
	} else {
		for _, idx := range strings.Split(*flagSpiders, ",") {
			idx = strings.TrimSpace(idx)
			if idx == "" {
				continue
			}
			i, _ := strconv.Atoi(idx)
			sps = append(sps, FKApp.G_App.GetSpiderTypeList()[i])
			FKLog.G_Log.Informational("使用蜘蛛：%d - %s", i, FKApp.G_App.GetSpiderTypeList()[i].Name)
		}
	}

	FKApp.G_App.SpiderPrepare(sps).Run()
}

// 输出提示文字格式
func outputInformations(){
	FKLog.G_Log.Informational("\n请添加任务:\n必填参数：%v\n可选参数：%v\n格式参考：-Spiders=1,2,3 -RequestLimit=0 -MaxThreadNum=20",
		"-Spiders", []string{
			"-Keywords",
			"-RequestLimit",
			"-OutputType",
			"-MaxThreadNum",
			"-MedianPauseTime",
			"-UpdateProxyIntervale",
			"-DockerCap",
			"-IsInheritSuccess",
			"-IsInheritFailure"})
	FKLog.G_Log.Informational("当前支持的蜘蛛类型包括:")
	for i, s := range FKApp.G_App.GetSpiderTypeList(){
		FKLog.G_Log.Informational("[ %d ] - %s", i, s.Name)
	}
	FKLog.G_Log.Informational("\n请输入任务参数...")
}

// 服务器模式下接收添加任务的参数
func parseInput() {
	outputInformations()
retry:
	*flagSpiders = ""
	input := [12]string{}
	fmt.Scanln(&input[0], &input[1], &input[2], &input[3], &input[4], &input[5], &input[6], &input[7], &input[8], &input[9])
	if strings.Index(input[0], "=") < 4 {
		FKLog.G_Log.Warning("\n参数格式不正确，请查看格式参考")
		outputInformations()
		goto retry
	}
	for _, v := range input {
		i := strings.Index(v, "=")
		if i < 4 {
			continue
		}
		key, value := v[:i], v[i+1:]
		switch key {
		case "-Keywords":
			FKStatus.GlobalRuntimeTaskConfig.Keywords = value
		case "-RequestLimit":
			limit, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				break
			}
			FKStatus.GlobalRuntimeTaskConfig.RequestLimit = limit
		case "-OutputType":
			FKStatus.GlobalRuntimeTaskConfig.OutputType = value
		case "-MaxThreadNum":
			thread, err := strconv.Atoi(value)
			if err != nil {
				break
			}
			FKStatus.GlobalRuntimeTaskConfig.MaxThreadNum = thread
		case "-MedianPauseTime":
			pause, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				break
			}
			FKStatus.GlobalRuntimeTaskConfig.MedianPauseTime = pause
		case "-UpdateProxyIntervale":
			proxyminute, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				break
			}
			FKStatus.GlobalRuntimeTaskConfig.UpdateProxyIntervale = proxyminute
		case "-DockerCap":
			dockercap, err := strconv.Atoi(value)
			if err != nil {
				break
			}
			if dockercap < 1 {
				dockercap = 1
			}
			FKStatus.GlobalRuntimeTaskConfig.DockerCap = dockercap
		case "-IsInheritSuccess":
			if value == "true" {
				FKStatus.GlobalRuntimeTaskConfig.IsInheritSuccess = true
			} else if value == "false" {
				FKStatus.GlobalRuntimeTaskConfig.IsInheritSuccess = false
			}
		case "-IsInheritFailure":
			if value == "true" {
				FKStatus.GlobalRuntimeTaskConfig.IsInheritFailure = true
			} else if value == "false" {
				FKStatus.GlobalRuntimeTaskConfig.IsInheritFailure = false
			}
		case "-Spiders":
			*flagSpiders = value
		default:
			FKLog.G_Log.Warning("\n解析失败，含有未知参数")
			outputInformations()
			goto retry
		}
	}
}
