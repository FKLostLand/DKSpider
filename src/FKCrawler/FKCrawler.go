package FKCrawler

import (
	"FKDownloader"
	"FKLog"
	"FKPipeline"
	"FKRequest"
	"FKSpider"
	"FKStatus"
	"bytes"
	"math/rand"
	"runtime"
	"time"
)

type (
	crawler struct {
		*FKSpider.Spider                 // 执行的采集规则
		FKDownloader.Downloader          // 全局公用的下载器
		FKPipeline.Pipeline              // 结果收集与输出管道
		id                      int      // 引擎ID
		pause                   [2]int64 // [请求间隔的最短时长,请求间隔的增幅时长]
	}
)

// 初始化
func (c *crawler) Init(sp *FKSpider.Spider) Crawler {
	c.Spider = sp.ReqmatrixInit()
	c.Pipeline = FKPipeline.CreatePipeline(sp)
	c.pause[0] = sp.MedianPauseTime / 2
	if c.pause[0] > 0 {
		c.pause[1] = c.pause[0] * 3
	} else {
		c.pause[1] = 1
	}
	return c
}

// 任务执行入口
func (c *crawler) Run() {
	// 预先启动数据收集/输出管道
	c.Pipeline.Start()

	// 运行处理协程
	bChan := make(chan bool)
	go func() {
		c.run()
		close(bChan)
	}()

	// 启动任务
	c.Spider.Start()

	<-bChan // 等待处理协程退出

	// 停止数据收集/输出管道
	c.Pipeline.Stop()
}

// 主动终止
func (c *crawler) Stop() {
	// 主动崩溃爬虫运行协程
	c.Spider.Stop()
	c.Pipeline.Stop()
}

// 是否可以停止
func (c *crawler) CanStop() bool {
	return c.Spider.CanStop()
}

// 设置采集引擎ID
func (c *crawler) GetId() int {
	return c.id
}

// 单独一个协程进行采集处理
func (c *crawler) run() {
	for {
		// 队列中取出一条请求并处理
		req := c.getOneQuest()
		if req == nil {
			// 停止任务
			if c.Spider.CanStop() {
				break
			}
			time.Sleep(20 * time.Millisecond)
			continue
		}

		// 执行请求
		c.useOneQuest()
		go func() {
			defer func() {
				c.freeOneQuest()
			}()
			FKLog.G_Log.Debug(" *     Start: %v", req.GetUrl())
			c.process(req)
		}()

		// 随机等待
		c.sleep()
	}

	// 等待处理中的任务完成
	c.Spider.Defer()
}

// 真正的请求处理
func (c *crawler) process(req *FKRequest.Request) {
	var (
		downUrl = req.GetUrl()
		sp      = c.Spider
	)
	defer func() {
		if p := recover(); p != nil {
			if sp.IsStopping() {
				return
			}
			// 返回是否作为新的失败请求被添加至队列尾部
			if sp.DoHistory(req, false) {
				// 统计失败数
				FKStatus.AddFailedRequestPageNum()
			}
			// 提示错误
			stack := make([]byte, 4<<10) //4KB
			length := runtime.Stack(stack, true)
			start := bytes.Index(stack, []byte("/src/runtime/panic.go"))
			stack = stack[start:length]
			start = bytes.Index(stack, []byte("\n")) + 1
			stack = stack[start:]
			if end := bytes.Index(stack, []byte("\ngoroutine ")); end != -1 {
				stack = stack[:end]
			}
			stack = bytes.Replace(stack, []byte("\n"), []byte("\r\n"), -1)
			FKLog.G_Log.Error(" *     Panic  [process][%s]: %s\r\n[TRACE]\r\n%s", downUrl, p, stack)
		}
	}()

	var ctx = c.Downloader.Download(sp, req) // download page

	if err := ctx.GetError(); err != nil {
		// 返回是否作为新的失败请求被添加至队列尾部
		if sp.DoHistory(req, false) {
			// 统计失败数
			FKStatus.AddFailedRequestPageNum()
		}
		// 提示错误
		FKLog.G_Log.Error(" *     Fail  [download][%v]: %v", downUrl, err)
		return
	}

	// 过程处理，提炼数据
	ctx.Parse(req.GetRuleName())

	// 该条请求文件结果存入pipeline
	for _, f := range ctx.PullFiles() {
		if c.Pipeline.CollectFile(f) != nil {
			break
		}
	}
	// 该条请求文本结果存入pipeline
	for _, item := range ctx.PullItems() {
		if c.Pipeline.CollectData(item) != nil {
			break
		}
	}

	// 处理成功请求记录
	sp.DoHistory(req, true)

	// 统计成功页数
	FKStatus.AddSuccessRequestPageNum()

	// 提示抓取成功
	FKLog.G_Log.Informational(" *     Success: %v", downUrl)

	// 释放ctx准备复用
	FKSpider.PutContext(ctx)
}

// 采集引擎进行休眠
func (c *crawler) sleep() {
	sleeptime := c.pause[0] + rand.Int63n(c.pause[1])
	time.Sleep(time.Duration(sleeptime) * time.Millisecond)
}

// 从调度读取一个请求
func (c *crawler) getOneQuest() *FKRequest.Request {
	return c.Spider.RequestPull()
}

//从调度使用一个资源空位
func (c *crawler) useOneQuest() {
	c.Spider.RequestUse()
}

//从调度释放一个资源空位
func (c *crawler) freeOneQuest() {
	c.Spider.RequestFree()
}
