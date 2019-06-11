package FKPipeline

import (
	"FKKafka"
	"FKMySQL"
	"FKSpider"
	"FKStatus"
	"FKTempDataPool"
)

// 数据收集/输出管道
type Pipeline interface {
	Start()                                    //启动
	Stop()                                     //停止
	CollectData(FKTempDataPool.DataCell) error //收集数据单元
	CollectFile(FKTempDataPool.FileCell) error //收集文件
}

func CreatePipeline(spider *FKSpider.Spider) Pipeline {
	var self = &pipeline{}
	self.Spider = spider
	self.outType = FKStatus.GlobalRuntimeTaskConfig.OutputType
	if FKStatus.GlobalRuntimeTaskConfig.DockerCap < 1 {
		FKStatus.GlobalRuntimeTaskConfig.DockerCap = 1
	}
	self.DataChan = make(chan FKTempDataPool.DataCell, FKStatus.GlobalRuntimeTaskConfig.DockerCap)
	self.FileChan = make(chan FKTempDataPool.FileCell, FKStatus.GlobalRuntimeTaskConfig.DockerCap)
	self.dataDocker = make([]FKTempDataPool.DataCell, 0, FKStatus.GlobalRuntimeTaskConfig.DockerCap)
	self.sum = [4]uint64{}
	self.dataBatch = 0
	self.fileBatch = 0
	return self
}

// 刷新输出方式的状态
func RefreshOutput() {
	switch FKStatus.GlobalRuntimeTaskConfig.OutputType {
	case "mysql":
		FKMySQL.Refresh()
	case "kafka":
		FKKafka.Refresh()
	}
}
