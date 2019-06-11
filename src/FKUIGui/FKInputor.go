package FKUIGui

import "FKStatus"

// GUI输入
type Inputor struct {
	Spiders []*GUISpider
	*FKStatus.AppRuntimeConfig
	Pausetime   int64
	ProxyMinute int64
}

var globalInputor = &Inputor{
	AppRuntimeConfig: FKStatus.GlobalRuntimeTaskConfig,
	Pausetime:        FKStatus.GlobalRuntimeTaskConfig.MedianPauseTime,
	ProxyMinute:      FKStatus.GlobalRuntimeTaskConfig.UpdateProxyIntervale,
}
