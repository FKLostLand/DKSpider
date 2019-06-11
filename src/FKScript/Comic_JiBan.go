package FKScript

import (
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
	"FKSpider"
	"FKRequest"
)

func init() {
	Comic_JiBan_Spider.RegisterToSpiderSpecies()
}

var Comic_JiBan_Spider = &FKSpider.Spider{
	Name:         "新闻站_羁绊动漫",
	Description:  "羁绊二次元动漫新闻资讯，抓取内容和图片 [http://www.005.tv/zx/]",
	EnableCookie: true,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.AddQueue(&FKRequest.Request{
				Url:         "http://www.005.tv/zx/list_526_1.html",
				Rule:        "请求",
				Temp:        map[string]interface{}{"p": 1},
				ConnTimeout: -1,
				Reloadable:  true,
			})

		},
		Trunk: map[string]*FKSpider.Rule{
			"请求": {
				ParseFunc: func(ctx *FKSpider.Context) {
					var curr = ctx.GetTemp("p", int(0)).(int)
					ctx.GetDom().Find(".pages .dede_pages  .pagelist  .thisclass a").Each(func(ii int, iio *goquery.Selection) {
						url2, _ := iio.Attr("href")
						if url2 != "javascript:void(0);" {
							if curr > 100 {
								return
							}
						}
					})
					ctx.AddQueue(&FKRequest.Request{
						Url:         "http://www.005.tv/zx/list_526_" + strconv.Itoa(curr+1) + ".html",
						Rule:        "请求",
						Temp:        map[string]interface{}{"p": curr + 1},
						ConnTimeout: -1,
						Reloadable:  true,
					})
					ctx.Parse("获取列表")
				},
			},

			"获取列表": {
				ParseFunc: func(ctx *FKSpider.Context) {
					ctx.GetDom().
						Find(".article-list ul li .xs-100 div h3 a").
						Each(func(i int, s *goquery.Selection) {
						url, _ := s.Attr("href")
						ctx.AddQueue(&FKRequest.Request{
							Url:         url,
							Rule:        "news",
							ConnTimeout: -1,
						})
					})
				},
			},

			"news": {
				ItemFields: []string{
					"title",
					"time",
					"img_url",
					"content",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					var title, time, img_url, content string
					query.Find(".article-list-wrap").
						Each(func(j int, jo *goquery.Selection) {
						title = jo.Find(".articleTitle-name").Text()
						time = jo.Find("span.time").Text()
						jo.Find(".articleContent img").Each(func(x int, xo *goquery.Selection) {
							if img, ok := xo.Attr("src"); ok {
								img_url = img_url + img + ","
							}
						})
						jo.Find(".articleContent img").ReplaceWithHtml("#image#")
						jo.Find(".articleContent img").Remove()
						content, _ = jo.Find(".articleContent").Html()
						content = strings.Replace(content, `"`, `'`, -1)
					})
					ctx.Output(map[int]interface{}{
						0: title,
						1: time,
						2: img_url,
						3: content,
					})
				},
			},
		},
	},
}
