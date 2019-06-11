package FKCrawler

import (
	"FKDownloader"
	"FKSpider"
)

// 采集引擎
type (
	Crawler interface {
		Init(*FKSpider.Spider) Crawler // 初始化采集引擎
		Run()                          // 运行采集引擎
		Stop()                         // 主动终止采集引擎
		CanStop() bool                 // 能否终止采集引擎
		GetId() int                    // 获取引擎ID
	}
)

// 创建一个采集引擎对象
// 参数Id: 该采集引擎的唯一ID
func CreateCrawler(id int) Crawler {
	return &crawler{
		id:         id,
		Downloader: FKDownloader.GlobalMixDownloader,
	}
}
