package FKScript

import (
	"FKSpider"
	"FKRequest"
)

func init() {
	FileTest_Spider.RegisterToSpiderSpecies()
}

var FileTest_Spider = &FKSpider.Spider{
	Name:        "测试站_网络下载测试",
	Description: "进行网络测试和最基本的爬虫下载测试，抓取 百度的一张Logo图片,以及github的google HTML页面]",
	EnableCookie: false,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.AddQueue(&FKRequest.Request{
				Url:          "https://www.baidu.com/img/bd_logo1.png",
				Rule:         "百度图片",
				ConnTimeout:  -1,
				DownloaderID: 0, //图片等多媒体文件必须使用0（surfer surf go原生下载器）
			})
			ctx.AddQueue(&FKRequest.Request{
				Url:          "https://github.com/google",
				Rule:         "Github页面",
				ConnTimeout:  -1,
				DownloaderID: 0, //文本文件可使用0或者1（0：surfer surf go原生下载器；1：surfer plantomjs内核）
			})
		},

		Trunk: map[string]* FKSpider.Rule{

			"百度图片": {
				ParseFunc: func(ctx * FKSpider.Context) {
					ctx.FileOutput("baidu") // 等价于ctx.AddFile("baidu")
				},
			},
			"Github页面": {
				ParseFunc: func(ctx *FKSpider.Context) {
					ctx.FileOutput() // 等价于ctx.AddFile()
				},
			},
		},
	},
}