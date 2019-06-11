package FKScript

import (
	"github.com/PuerkitoBio/goquery"
	"FKSpider"
	"FKRequest"
)

func init() {
	Retail_KaoLa_Spider.RegisterToSpiderSpecies()
}

var Retail_KaoLa_Spider = &FKSpider.Spider{
	Name:        "数据站_考拉海淘",
	Description: "考拉海淘商品数据，抓取商品名称,价格,评论等 [www.kaola.com]",
	EnableCookie: false,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.AddQueue(&FKRequest.Request{
				Url: "http://www.kaola.com",
				Rule: "获取版块URL",
				})
		},

		Trunk: map[string]* FKSpider.Rule{

			"获取版块URL": {
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					lis := query.Find("#funcTab li a")
					lis.Each(func(i int, s *goquery.Selection) {
						if i == 0 {
							return
						}
						if url, ok := s.Attr("href"); ok {
							ctx.AddQueue(&FKRequest.Request{
								Url: url,
								Rule: "商品列表",
								Temp: map[string]interface{}{"goodsType": s.Text()},
								})
						}
					})
				},
			},

			"商品列表": {
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					query.Find(".proinfo").Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Find("a").Attr("href"); ok {
							ctx.AddQueue(&FKRequest.Request{
								Url:  "http://www.kaola.com" + url,
								Rule: "商品详情",
								Temp: map[string]interface{}{"goodsType": ctx.GetTemp("goodsType", "").(string)},
							})
						}
					})
				},
			},

			"商品详情": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"标题",
					"价格",
					"品牌",
					"采购地",
					"评论数",
					"类别",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					// 获取标题
					title := query.Find(".product-title").Text()

					// 获取价格
					price := query.Find("#js_currentPrice span").Text()

					// 获取品牌
					brand := query.Find(".goods_parameter li").Eq(0).Text()

					// 获取采购地
					from := query.Find(".goods_parameter li").Eq(1).Text()

					// 获取评论数
					discussNum := query.Find("#commentCounts").Text()

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: price,
						2: brand,
						3: from,
						4: discussNum,
						5: ctx.GetTemp("goodsType", ""),
					})
				},
			},
		},
	},
}
