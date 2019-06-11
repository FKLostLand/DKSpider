package FKScript

import (
	"strconv"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"FKSpider"
	"FKRequest"
)

func init() {
	Bus_QiCheZhiJia_Spider.RegisterToSpiderSpecies()
}

var Bus_QiCheZhiJia_Spider = &FKSpider.Spider{
	Name:        "论坛站_汽车之家论坛",
	Description: "抓取用户帖子和相关信息 [http://club.autohome.com.cn/]",
	EnableCookie: false,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.AddQueue(&FKRequest.Request{
				Url:  "http://club.autohome.com.cn/bbs/forum-o-200042-1.html?qaType=-1#pvareaid=101061",
				Rule: "请求列表",
				Temp: map[string]interface{}{"p": 1},
			})
		},

		Trunk: map[string]* FKSpider.Rule{

			"请求列表": {
				ParseFunc: func(ctx *FKSpider.Context) {
					var curr = ctx.GetTemp("p", 0).(int)
					if c := ctx.GetDom().Find(".pages .cur").Text(); c != strconv.Itoa(curr) {
						return
					}
					ctx.AddQueue(&FKRequest.Request{
						Url:  "http://club.autohome.com.cn/bbs/forum-o-200042-" + strconv.Itoa(curr+1) + ".html?qaType=-1#pvareaid=101061",
						Rule: "请求列表",
						Temp: map[string]interface{}{"p": curr + 1},
					})

					// 用指定规则解析响应流
					ctx.Parse("获取列表")
				},
			},

			"获取列表": {
				ParseFunc: func(ctx *FKSpider.Context) {
					ctx.GetDom().
						Find(".list_dl").
						Each(func(i int, s *goquery.Selection) {
						url, _ := s.Find("dt a").Attr("href")
						ctx.AddQueue(&FKRequest.Request{
							Url:      "http://club.autohome.com.cn" + url,
							Rule:     "输出结果",
							Priority: 1,
						})
					})
				},
			},

			"输出结果": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{

					"当前积分",
					"帖子数",
					"关注的车",
					"注册时间",
					"作者",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()

					var 当前积分, 帖子数, 关注的车, 注册时间, 作者 string

					积分 := strings.Split(query.Find(".lv-curr").First().Text(), "当前积分：")
					if len(积分) > 1 {
						当前积分 = 积分[1]
					}

					info := query.Find(".conleft").Eq(0).Find(".leftlist li")

					if len(info.Eq(3).Nodes) > 0 {
						帖子数 = strings.Split(info.Eq(3).Find("a").Text(), "帖")[0]
					}

					for i := 6; !info.Eq(i).HasClass("leftimgs") &&
						len(info.Eq(i).Nodes) > 0 &&
						len(info.Eq(i).Find("a").Nodes) > 0; i++ {
						if strings.Contains(info.Eq(i).Text(), "所属：") {
							continue
						}

						fs := info.Eq(i).Find("a")
						var f string
						if len(fs.Nodes) > 1 {
							f, _ = info.Eq(i).Find("a").Eq(1).Attr("title")
						} else {
							f, _ = info.Eq(i).Find("a").First().Attr("title")
						}
						if f == "" {
							continue
						}
						关注的车 += f + "|"
					}

					关注的车 = strings.Trim(关注的车, "|")

					if len(info.Eq(4).Nodes) > 0 {
						注册 := strings.Split(info.Eq(4).Text(), "注册：")
						if len(注册) > 1 {
							注册时间 = 注册[1]
						}
					}
					作者 = query.Find(".conleft").Eq(0).Find("a").Text()
					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: 当前积分,
						1: 帖子数,
						2: 关注的车,
						3: 注册时间,
						4: 作者,
					})
				},
			},
		},
	},
}
