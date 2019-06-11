package FKScript

import (
	"encoding/json"
	"log"
	"FKSpider"
	"FKRequest"
	"strings"
)

func init() {
	News_RenMinWang_Spider.RegisterToSpiderSpecies()
}

type Item_RenMinWang struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Url      string `json:"url"`
	Date     string `json:"date"`
	NodeId   string `json:"nodeId"`
	ImgCount string `json:"imgCount"`
}
type News_RenMinWang struct {
	Items []Item_RenMinWang `json:"items"`
}

var news_RenMinWang News_RenMinWang

var News_RenMinWang_Spider = &FKSpider.Spider{
	Name:        "新闻站_人民网新闻",
	Description: "人民网最新分类新闻，抓取标题和内容 [http://news.people.com.cn/]",
	EnableCookie: false,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.AddQueue(&FKRequest.Request{
				Method: "GET",
				Url:    "http://news.people.com.cn/210801/211150/index.js?cache=false",
				Rule:   "新闻列表",
			})
		},

		Trunk: map[string]* FKSpider.Rule{
			"新闻列表": {
				ParseFunc: func(ctx *FKSpider.Context) {
					str := ctx.GetText()

					err := json.Unmarshal([]byte(str), &news_RenMinWang)
					if err != nil {
						log.Printf("解析错误： %v\n", err)
						return
					}

					newsLength := len(news_RenMinWang.Items)
					for i := 0; i < newsLength; i++ {
						ctx.AddQueue(&FKRequest.Request{
							Url:  news_RenMinWang.Items[i].Url,
							Rule: "热点新闻",
							Temp: map[string]interface{}{
								"id":       news_RenMinWang.Items[i].Id,
								"title":    news_RenMinWang.Items[i].Title,
								"date":     news_RenMinWang.Items[i].Date,
								"newsType": news_RenMinWang.Items[i].NodeId,
							},
						})
					}
				},
			},

			"热点新闻": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"ID",
					"标题",
					"内容",
					"类别",
					"ReleaseTime",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()

					// 获取内容
					content := query.Find("#p_content").Text()
					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: ctx.GetTemp("id", ""),
						1: ctx.GetTemp("title", ""),
						2: strings.TrimSpace(content),
						3: ctx.GetTemp("newsType", ""),
						4: ctx.GetTemp("date", ""),
					})
				},
			},
		},
	},
}
