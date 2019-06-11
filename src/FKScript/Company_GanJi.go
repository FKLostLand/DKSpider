package FKScript

import (
	"strconv"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"FKSpider"
	"FKRequest"
)

func init() {
	Company_GanJi_Spider.RegisterToSpiderSpecies()
}

var Company_GanJi_Spider = &FKSpider.Spider{
	Name:        "数据站_赶集网企业名录",
	Description: "赶集网公司列表，抓取企业资讯 [http://sz.ganji.com/gongsi/o1]",
	EnableCookie: false,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.AddQueue(&FKRequest.Request{
				Url:  "http://sz.ganji.com/gongsi/o1",
				Rule: "请求列表",
				Temp: map[string]interface{}{"p": 1},
			})
		},

		Trunk: map[string]* FKSpider.Rule{

			"请求列表": {
				ParseFunc: func(ctx *FKSpider.Context) {
					var curr = ctx.GetTemp("p", int(0)).(int)
					if ctx.GetDom().Find(".linkOn span").Text() != strconv.Itoa(curr) {
						return
					}
					ctx.AddQueue(&FKRequest.Request{
						Url:         "http://sz.ganji.com/gongsi/o" + strconv.Itoa(curr+1),
						Rule:        "请求列表",
						Temp:        map[string]interface{}{"p": curr + 1},
						ConnTimeout: -1,
					})

					// 用指定规则解析响应流
					ctx.Parse("获取列表")
				},
			},

			"获取列表": {
				ParseFunc: func(ctx *FKSpider.Context) {
					ctx.GetDom().
						Find(".com-list-2 table a").
						Each(func(i int, s *goquery.Selection) {
						url, _ := s.Attr("href")
						ctx.AddQueue(&FKRequest.Request{
							Url:         url,
							Rule:        "输出结果",
							ConnTimeout: -1,
						})
					})
				},
			},

			"输出结果": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"公司",
					"联系人",
					"地址",
					"简介",
					"行业",
					"类型",
					"规模",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()

					var 公司, 规模, 行业, 类型, 联系人, 地址 string

					query.Find(".c-introduce li").Each(func(i int, s *goquery.Selection) {
						em := s.Find("em").Text()
						t := strings.Split(s.Text(), `   `)[0]
						t = strings.Replace(t, em, "", -1)
						t = strings.Trim(t, " ")

						switch em {
						case "公司名称：":
							公司 = t

						case "公司规模：":
							规模 = t

						case "公司行业：":
							行业 = t

						case "公司类型：":
							类型 = t

						case "联 系 人：":
							联系人 = t

						case "联系电话：":
							if img, ok := s.Find("img").Attr("src"); ok {
								ctx.AddQueue(&FKRequest.Request{
									Url:         "http://www.ganji.com" + img,
									Rule:        "联系方式",
									Temp:        map[string]interface{}{"n": 公司 + "(" + 联系人 + ").png"},
									Priority:    1,
									ConnTimeout: -1,
								})
							}

						case "公司地址：":
							地址 = t
						}
					})

					简介 := query.Find("#company_description").Text()
					ctx.Output(map[int]interface{}{
						0: 公司,
						1: 联系人,
						2: 地址,
						3: 简介,
						4: 行业,
						5: 类型,
						6: 规模,
					})
				},
			},

			"联系方式": {
				ParseFunc: func(ctx *FKSpider.Context) {
					ctx.FileOutput(ctx.GetTemp("n", "").(string))
				},
			},
		},
	},
}
