package FKCrawler

import (
	"FKConfig"
	"FKStatus"
	"sync"
	"time"
)

type (
	crawlerPool struct {
		poolSize     int // 采集引擎池实际大小
		usingCount   int // 已被使用的采集引擎对象数量
		usable       chan Crawler
		crawlerArray []Crawler // 所有的采集引擎对象数组
		status       int       // 当前采集引擎池的状态
		sync.RWMutex           // 读写锁
	}
)

func CreateCrawlerPool() CrawlerPool {
	return &crawlerPool{
		status:       FKStatus.RUN,
		crawlerArray: make([]Crawler, 0, FKConfig.CONFIG_CRAWL_CAP),
	}
}

// 根据要执行的蜘蛛数量设置CrawlerPool大小
func (p *crawlerPool) ResetPoolSize(crawlerNum int) int {
	p.Lock()
	defer p.Unlock()

	var wantNum int
	if crawlerNum < FKConfig.CONFIG_CRAWL_CAP {
		wantNum = crawlerNum
	} else {
		wantNum = FKConfig.CONFIG_CRAWL_CAP
	}
	if wantNum <= 0 {
		wantNum = 1
	}
	p.poolSize = wantNum
	p.usingCount = 0
	p.usable = make(chan Crawler, wantNum)
	for _, crawler := range p.crawlerArray {
		if p.usingCount < p.poolSize {
			p.usable <- crawler
			p.usingCount++
		}
	}
	p.status = FKStatus.RUN
	return wantNum
}

// 请求从池内分配一个采集引擎
func (p *crawlerPool) AllocCrawlerFromPool() Crawler {
	var crawler Crawler
	for {
		p.Lock()
		if p.status == FKStatus.STOP {
			p.Unlock()
			return nil
		}
		select {
		case crawler = <-p.usable:
			p.Unlock()
			return crawler
		default:
			if p.usingCount < p.poolSize {
				crawler = CreateCrawler(p.usingCount)
				p.crawlerArray = append(p.crawlerArray, crawler)
				p.usingCount++
				p.Unlock()
				return crawler
			}
		}
		p.Unlock()
		time.Sleep(time.Second)
	}
	return nil
}

// 归还一个采集引擎到池内
func (p *crawlerPool) ReturnBackCrawlerToPool(crawler Crawler) {
	p.RLock()
	defer p.RUnlock()

	if p.status == FKStatus.STOP || !crawler.CanStop() {
		return
	}
	p.usable <- crawler
}

// 主动终止所有爬行任务
func (p *crawlerPool) StopAllCrawlers() {
	p.Lock()
	if p.status == FKStatus.STOP {
		p.Unlock()
		return
	}
	p.status = FKStatus.STOP
	close(p.usable)
	p.usable = nil
	p.Unlock()

	for _, crawler := range p.crawlerArray {
		crawler.Stop()
	}
}
