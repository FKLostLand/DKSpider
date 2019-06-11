package FKScript

import (
	"FKSpider"
	"FKBase"
	"FKRequest"
	"strings"
	"encoding/json"
	"FKLog"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strconv"
)

/*
keywords:
首页：youlike
新时代：good_safe2toera
辟谣：refuteRumour
汽车: car
房产：estate
科技：science
财经：economy
体育：sport
旅游：travel
美食：food
游戏：game
教育：education
股票：stock
育儿：child
养生：healthcare
娱乐：fun
国际：international
国内：domestic
 */

func init(){
	News_KuaiZiXun_Spider.RegisterToSpiderSpecies()
}

type News_KuaiZiXunList_Article struct{
	Title string `json:"title"`
	Content string `json:"content"`
}

type News_KuaiZiXunList_Res struct{
	Abst string `json:"abst"`
	DetailApi string `json:"u"`
}

type News_KuaiZiXunList_Data struct{
	Res []News_KuaiZiXunList_Res `json:"res"`
}

type News_KuaiZiXunList_Response struct{
	Data News_KuaiZiXunList_Data `json:"data"`
}

var News_KuaiZiXun_Spider = &FKSpider.Spider{
	Name: "新闻站_快资讯",
	Description: "【开启Keywords】360Qihoo新闻站，抓取标题和内容 [sh.qihoo.com]",
	EnableCookie: false,
	Keywords: FKBase.KEYWORDS,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.Aid(map[string]interface{}{
				"loop": [2]int{0, 1000},
				"Rule": "生成请求",
			}, "生成请求")
		},
		Trunk: map[string]* FKSpider.Rule{
			"生成请求":{
				AidFunc: func(ctx *FKSpider.Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&FKRequest.Request{
							Method: "GET",
							Url: "http://papi.look.360.cn/mlist?c=" + ctx.GetKeywords() + "&u=e90bc5f760e96470de94cc112f72fb07" +
								"&uid=e90bc5f760e96470de94cc112f72fb07&sign=360dh&version=2.0&sqid=&device=2&market=pc_def&net=4" +
									"&tj_cmode=pclook&where=list&pid=index&src=&v=1&sv=4&n=10&action=1&f=jsonp&stype=portal" +
										"&newest_showtime=&oldest_showtime=&ufrom=2&scene=2" + "&num=" + strconv.Itoa(loop[0]),
							Rule: aid["Rule"].(string),
						})
					}
					return nil
				},

				ParseFunc: func(ctx *FKSpider.Context) {
					str := ctx.GetText()
					str = strings.TrimPrefix(str, "undefined_callback(")
					str = strings.TrimSuffix(str, ");")
					var news_KuaiZiXunList_Response News_KuaiZiXunList_Response
					err := json.Unmarshal([]byte(str), &news_KuaiZiXunList_Response)
					if err != nil {
						FKLog.G_Log.Informational("解析快资讯列表错误： %v\n", err)
						return
					}

					newsLength := len(news_KuaiZiXunList_Response.Data.Res)
					for i := 0; i < newsLength; i++ {
						ctx.AddQueue(&FKRequest.Request{
							Url:  news_KuaiZiXunList_Response.Data.Res[i].DetailApi,
							Rule: "输出结果",
							Priority: 1,
						})
					}
				},
			},

			"输出结果":{
				ItemFields: []string{
					"标题",
					"内容",
				},

				ParseFunc: func(ctx *FKSpider.Context){
					var title, content string
					title = ""
					content = ""

					html := ctx.GetDom()

					var article News_KuaiZiXunList_Article
					v := html.Find("script").FilterFunction(
						func(i int, s *goquery.Selection) bool {
							v := s.Text()
							return strings.Contains(v, "data_new =")
					}).Text()

					v = strings.TrimSpace(v)
					v = strings.TrimPrefix(v,"var data_new = ")
					v = strings.Split(v, ";")[0]

					err := json.Unmarshal([]byte(v), &article)
					if err != nil {
						log.Printf("解析错误： %v\n", err)
						return
					}

					title = strings.TrimSpace(article.Title)
					content = strings.TrimSpace(article.Content)

					if title != "" || content != "" {
						ctx.Output(map[int]interface{}{
							0: title,
							1: content,
						})
					}
				},
			},
		},
	},
}