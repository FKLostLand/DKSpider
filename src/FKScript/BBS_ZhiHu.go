package FKScript

import (
	"github.com/PuerkitoBio/goquery"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"strings"
	"regexp"
	"FKSpider"
	"FKRequest"
)

/*
## 知乎编辑推荐

> 目前抓取推荐专栏的问题和回答。
> 能够翻页抓取，
> 抓取的内容中的段落标签(``<p>``)、图片标签(``<img>``)等均原封不动的抓取过来，没做转义替换处理
> 编辑中有两类文本，一类是知乎作家写的文章，一类是知乎用户回答的问题。这两类均抓取了
> 支持采集最少url数，即可以手动输入"采集上限"，那就是最少采集数
*/

func init() {
	BBS_ZhiHu_Spider.RegisterToSpiderSpecies()
}

var urlList []string

var BBS_ZhiHu_Spider = &FKSpider.Spider{
	Name:        "论坛站_知乎编辑推荐",
	Description: "抓取推荐专栏的问题和回答 [https://www.zhihu.com/explore/recommendations]",
	EnableCookie: false,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.AddQueue(&FKRequest.Request{
				Url: "https://www.zhihu.com/explore/recommendations",
				Rule: "知乎编辑推荐",
			})


		},

		Trunk: map[string]*FKSpider.Rule{
			"知乎编辑推荐": {
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					regular := "#zh-recommend-list-full .zh-general-list .zm-item h2 a";
					query.Find(regular).
						Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Attr("href"); ok {
							url = changeToAbspath(url)
							ctx.AddQueue(&FKRequest.Request{Url: url, Rule: "解析落地页"})
						}})

					limit := ctx.GetLimit()

					if len(query.Find(regular).Nodes) < limit	{
						total := int(math.Ceil(float64(limit) / float64(20)))
						ctx.Aid(map[string]interface{}{
							"loop": [2]int{1, total},
							"Rule": "知乎编辑推荐翻页",
						}, "知乎编辑推荐翻页")
					}
				},
			},

			"知乎编辑推荐翻页": {
				AidFunc: func(ctx *FKSpider.Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						offset := loop[0] * 20
						header := make(http.Header)
						header.Set("Content-Type", "application/x-www-form-urlencoded")
						ctx.AddQueue(&FKRequest.Request{
							Url:  "https://www.zhihu.com/node/ExploreRecommendListV2",
							Rule: aid["Rule"].(string),
							Method: "POST",
							Header: header,
							PostData: url.Values{"method":{"next"}, "params":{`{"limit":20,"offset":` +
								strconv.Itoa(offset) + `}`}}.Encode(),
							Reloadable: true,
						})
					}

					return nil
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					type Items struct {
						R int `json:"r"`
						Msg []interface{} `json:"msg"`
					}

					content, err := ioutil.ReadAll(ctx.GetResponse().Body)
					ctx.GetResponse().Body.Close()
					if err != nil {
						ctx.Log().Error(err.Error());
					}
					e := new(Items)
					err = json.Unmarshal(content, e)
					html := ""
					for _, v := range e.Msg{
						msg, ok := v.(string)
						if ok {
							html = html + "\n" + msg
						}
					}

					ctx = ctx.ResetText(html)
					query := ctx.GetDom()
					query.Find(".zm-item h2 a").Each(func(i int, selection *goquery.Selection){
						if url, ok := selection.Attr("href"); ok {
							url = changeToAbspath(url)
							if filterZhihuAnswerURL(url){
								ctx.AddQueue(&FKRequest.Request{Url: url, Rule: "解析知乎问答落地页"})
							}else{
								ctx.AddQueue(&FKRequest.Request{Url: url, Rule: "解析知乎文章落地页"})
							}
						}
					})

				},
			},

			"解析知乎问答落地页": {
				ItemFields: []string{
					"标题",
					"提问内容",
					"回答内容",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()

					questionHeader := query.Find(".QuestionPage .QuestionHeader .QuestionHeader-content")

					headerMain := questionHeader.Find(".QuestionHeader-main")

					// 获取问题标题
					title := headerMain.Find(".QuestionHeader-title").Text()

					// 获取问题描述
					content := headerMain.Find(".QuestionHeader-detail span").Text()

					answerMain := query.Find(".QuestionPage .Question-main")

					answer, _ := answerMain.Find(".AnswerCard .QuestionAnswer-content .ContentItem .RichContent .RichContent-inner").First().Html()

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: content,
						2: answer,
					})

				},
			},

			"解析知乎文章落地页": {
				ItemFields: []string{
					"标题",
					"内容",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()

					// 获取问题标题
					title,_ := query.Find(".PostIndex-title.av-paddingSide.av-titleFont").Html()

					// 获取问题描述
					content, _ := query.Find(".RichText.PostIndex-content.av-paddingSide.av-card").Html()

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: content,
					})

				},
			},
		},
	},
}

//将相对路径替换为绝对路径
func changeToAbspath(url string)string{
	if strings.HasPrefix(url, "https://"){
		return url
	}
	return "https://www.zhihu.com" + url
}

//判断是用户回答的问题，还是知乎专栏作家书写的文章
func filterZhihuAnswerURL(url string) bool{
	return regexp.MustCompile(`^https:\/\/www\.zhihu\.com\/question\/\d{1,}(\/answer\/\d{1,})?$`).MatchString(url)
}
