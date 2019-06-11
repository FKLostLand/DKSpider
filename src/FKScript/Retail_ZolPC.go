package FKScript

import (
	"strconv"
	"github.com/PuerkitoBio/goquery"
	"FKSpider"
	"FKRequest"
)

func init() {
	Retail_ZolPC_Spider.RegisterToSpiderSpecies()
}

var Retail_ZolPC_Spider = &FKSpider.Spider{
	Name:        "论坛站_中关村笔记本",
	Description: "笔记本论坛，抓取用户帖子 [http://bbs.zol.com.cn/nbbbs/]",
	EnableCookie: false,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.Aid(map[string]interface{}{"loop": [2]int{1, 720}, "Rule": "生成请求"}, "生成请求")
		},

		Trunk: map[string]* FKSpider.Rule{
			"生成请求": {
				AidFunc: func(ctx *FKSpider.Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&FKRequest.Request{
							Url:  "http://bbs.zol.com.cn/nbbbs/p" + strconv.Itoa(loop[0]) + ".html#c",
							Rule: aid["Rule"].(string),
						})
					}
					return nil
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					ss := query.Find("tbody").Find("tr[id]")
					ss.Each(func(i int, goq *goquery.Selection) {
						ctx.SetTemp("html", goq)
						ctx.Parse("获取结果")
					})
				},
			},

			"获取结果": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"机型",
					"链接",
					"主题",
					"发表者",
					"发表时间",
					"总回复",
					"总查看",
					"最后回复者",
					"最后回复时间",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					var selectObj = ctx.GetTemp("html", &goquery.Selection{}).(*goquery.Selection)

					//url
					outUrls := selectObj.Find("td").Eq(1)
					outUrl, _ := outUrls.Attr("data-url")
					outUrl = "http://bbs.zol.com.cn/" + outUrl
					//title type
					outTitles := selectObj.Find("td").Eq(1)
					outType := outTitles.Find(".iclass a").Text()
					outTitle := outTitles.Find("div a").Text()

					//author stime
					authors := selectObj.Find("td").Eq(2)
					author := authors.Find("a").Text()
					stime := authors.Find("span").Text()

					//reply read
					replys := selectObj.Find("td").Eq(3)
					reply := replys.Find("span").Text()
					read := replys.Find("i").Text()

					//ereply etime
					etimes := selectObj.Find("td").Eq(4)
					ereply := etimes.Find("a").Eq(0).Text()
					etime := etimes.Find("a").Eq(1).Text()

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: outType,
						1: outUrl,
						2: outTitle,
						3: author,
						4: stime,
						5: reply,
						6: read,
						7: ereply,
						8: etime,
					})
				},
			},
		},
	},
}
