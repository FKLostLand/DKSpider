package FKScript

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
	"FKSpider"
	"FKRequest"
)

/*
## 中国新闻网-滚动新闻栏目

### 说明

	只是爬取滚动新闻栏目（共10页）

### 代码说明

	1.直接访问滚动新闻栏目地址（http://www.chinanews.com/scroll-news/news1.html）
	2.获取分页导航
	3.获取分页链接
 */


func init() {
	News_ZhongGuoXinWen_Spider.RegisterToSpiderSpecies()
}

var News_ZhongGuoXinWen_Spider = &FKSpider.Spider{
	Name:        "新闻站_中国新闻网",
	Description: "爬取滚动新闻栏目 [http://www.chinanews.com/scroll-news/news1.html]",
	EnableCookie: false,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.AddQueue(&FKRequest.Request{
				Url:          "http://www.chinanews.com/scroll-news/news1.html",
				Rule:         "滚动新闻",
			})
		},

		Trunk: map[string]* FKSpider.Rule{
			"滚动新闻": {
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					//获取分页导航
					navBox := query.Find(".pagebox a")
					navBox.Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Attr("href"); ok {
							ctx.AddQueue(&FKRequest.Request{
								Url:  "http://www.chinanews.com" +  url,
								Rule: "新闻列表",

							})
						}

					})

				},
			},

			"新闻列表": {
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					//获取新闻列表
					newList := query.Find(".content_list li")
					newList.Each(func(i int, s *goquery.Selection) {
						//新闻类型
						newsType := s.Find(".dd_lm a").Text()
						//标题
						newsTitle := s.Find(".dd_bt a").Text()
						//时间
						newsTime := s.Find(".dd_time").Text()
						if url, ok := s.Find(".dd_bt a").Attr("href"); ok {
							ctx.AddQueue(&FKRequest.Request{
								Url:  "http://" + url[2:len(url)],
								Rule: "新闻内容",
								Temp: map[string]interface{}{
									"newsType":  newsType,
									"newsTitle": newsTitle,
									"newsTime":  newsTime,
								},
							})
						}

					})

				},
			},

			"新闻内容": {
				ItemFields: []string{
					"类别",
					"来源",
					"标题",
					"内容",
					"时间",
				},

				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					//正文
					content := query.Find(".left_zw").Text()
					//来源
					from := query.Find(".left-t").Text()
					i := strings.LastIndex(from,"来源")
					//来源字符串特殊处理
					if i == -1{
						from = "未知"
					}else{
						from = from[i+9:len(from)]
						from = strings.Replace(from,"参与互动","",1)
						if from=="" {
							from = query.Find(".left-t").Eq(2).Text()
							from = strings.Replace(from,"参与互动","",1)
						}
					}

					//输出格式
					ctx.Output(map[int]interface{}{
						0: ctx.GetTemp("newsType",""),
						1: from,
						2: ctx.GetTemp("newsTitle",""),
						3: content,
						4: ctx.GetTemp("newsTime", ""),
					})
				},
			},
		},
	},
}
