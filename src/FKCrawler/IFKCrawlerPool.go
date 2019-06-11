package FKCrawler

// 采集引擎池
type (
	CrawlerPool interface {
		// 重置采集引擎池大小
		// 返回值int： 实际分配的采集引擎池大小
		ResetPoolSize(crawlerNum int) int
		// 请求从池内分配一个采集引擎
		AllocCrawlerFromPool() Crawler
		// 归还一个采集引擎到池内
		ReturnBackCrawlerToPool(Crawler)
		// 主动终止所有爬行任务
		StopAllCrawlers()
	}
)
