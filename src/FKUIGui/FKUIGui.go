package FKUIGui

import (
	"FKApp"
	"FKSpider"
	"FKStatus"
	"fmt"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"log"
)

// 解析参数
func ParseFlag() {

}

// 输出选项
var (
	radioBtnList    []declarative.RadioButton
	runStopBtn      *walk.PushButton
	pauseRecoverBtn *walk.PushButton
	setting         *walk.Composite
	mw              *walk.MainWindow
	runMode         *walk.GroupBox
	db              *walk.DataBinder
	ep              walk.ErrorPresenter
	mode            *walk.GroupBox
	host            *walk.Splitter
	spiderMenu      *SpiderMenu
)

// 执行入口
func Main() {
	fmt.Println("正在运行Windows界面模式...")

	FKApp.G_App.SetAppConfig("Mode", FKStatus.OFFLINE)

	radioBtnList = func() (o []declarative.RadioButton) {
		// 设置默认选择
		globalInputor.AppRuntimeConfig.OutputType = FKApp.G_App.GetOutputTypeList()[0]
		// 获取输出选项
		for _, out := range FKApp.G_App.GetOutputTypeList() {
			o = append(o, declarative.RadioButton{Text: out, Value: out})
		}
		return
	}()

	spiderMenu = CreateSpiderMenu(FKSpider.GlobalSpiderSpecies)

	showLoginWindow()
}

func Init() {
	FKApp.G_App.Init(globalInputor.Mode, globalInputor.MasterPort, globalInputor.MasterIP)
}

func SetTaskConf() {
	// 纠正协程数
	if globalInputor.MaxThreadNum == 0 {
		globalInputor.MaxThreadNum = 1
	}
	FKApp.G_App.SetAppConfig("MaxThreadNum", globalInputor.MaxThreadNum).
		SetAppConfig("MedianPauseTime", globalInputor.MedianPauseTime).
		SetAppConfig("UpdateProxyIntervale", globalInputor.UpdateProxyIntervale).
		SetAppConfig("OutputType", globalInputor.OutputType).
		SetAppConfig("DockerCap", globalInputor.DockerCap).
		SetAppConfig("RequestLimit", globalInputor.RequestLimit).
		SetAppConfig("Keywords", globalInputor.Keywords)
}

func SpiderPrepare() {
	sps := []*FKSpider.Spider{}
	for _, sp := range globalInputor.Spiders {
		sps = append(sps, sp.Spider)
	}
	FKApp.G_App.SpiderPrepare(sps)
}

func SpiderNames() (names []string) {
	for _, sp := range globalInputor.Spiders {
		names = append(names, sp.Spider.GetName())
	}
	return
}

func CreateLogWindow() {
	// 绑定log输出界面
	lv, err := CreateLogView(mw)
	if err != nil {
		panic(err)
	}
	FKApp.G_App.SetLog(lv)
	log.SetOutput(lv)
	// 设置左上角图标
	if icon, err := walk.NewIconFromResourceId(0); err == nil {
		mw.SetIcon(icon)
	}
}
