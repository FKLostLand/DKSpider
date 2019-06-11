package FKScript

import (
	"FKSpider"
	"FKRequest"
	"encoding/json"
	"log"
	"strings"
	"strconv"
	"time"
	"FKLog"
	"net/http"
	"FKBase"
)

/*
* 参考：https://github.com/iMeiji/Toutiao/wiki/%E4%BB%8A%E6%97%A5%E5%A4%B4%E6%9D%A1Api%E5%88%86%E6%9E%90

keywords:
'推荐': '__all__',
'热点': 'news_hot',
'社会': 'news_society',
'娱乐': 'news_entertainment',
'科技': 'news_tech',
'军事': 'news_military',
'体育': 'news_sports'
'汽车': 'news_car',
'财经': 'news_finance',
'国际': 'news_world',
'时尚': 'news_fashion',
'旅游': 'news_travel',
'探索': 'news_discovery',
'育儿': 'news_baby',
'养生': 'news_regimen',
'故事': 'news_story',
'美文': 'news_essay',
'游戏': 'news_game',
'历史': 'news_history',
'美食': 'news_food',

 */

func init(){
	News_TouTiao_Spider.RegisterToSpiderSpecies()
}

type Item_TouTiao struct{
	Title string `json:"title"`
	Datetime string `json:"datetime"`
	Keywords string `json:"keywords"`
	GroupID string `json:"group_id"`
}

type News_TouTiaoList struct{
	Data[] Item_TouTiao `json:"data"`
	HasMore bool `json:"has_more"`
	ReturnCount int `json:"return_count"`
	Html string `json:"html"`
	PageId string `json:"page_id"`
}

type News_TouTiaoArticle_CK struct{}
type News_TouTiaoArticle_Data struct{
	DetailSource string `json:"detail_source"`
	PulishTime int64 `json:"pulish_time"`
	Title string `json:"title"`
	Content string `json:"content"`
}
type News_TouTiaoArticle struct {
	Success bool `json:"success"`
	CK News_TouTiaoArticle_CK `json:"_ck"`
	Data News_TouTiaoArticle_Data `json:"data"`
}

var News_TouTiao_Spider = &FKSpider.Spider{
	Name:        "新闻站_今日头条移动版",
	Description: "【开启Keywords】今日头条移动版新闻，抓取标题和内容 [m.toutiao.com]",
	EnableCookie: false,
	Keywords: FKBase.KEYWORDS,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context){
			// https://tool.chinaz.com/Tools/unixtime.aspx
			// 1370344418 2013-06-04
			// 1559026838 2019-05-28
			timestamp := time.Now().Unix()
			ctx.Aid(map[string]interface{}{
				"loop": [2]int64{1370344418, timestamp},
				}, "生成请求")
		},

		Trunk: map[string]* FKSpider.Rule{
			"生成请求":{
				AidFunc: func(ctx *FKSpider.Context, aid map[string]interface{}) interface{} {
					header := make(http.Header)
					header.Set("Host", "m.toutiao.com")
					header.Set("Cache-Control", "max-age=0")
					header.Set("Connection", "keep-alive")
					header.Set("Upgrade-Insecure-Requests", "1")
					header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
					header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
					header.Set("Accept-Encoding", "gzip, deflate")
					header.Set("Accept-Language", "zh-CN,zh;q=0.9")
					header.Set("Cookie", "tt_webid=6698541961139226119; UM_distinctid=16b20f1d48d703-0aac32d32c27ea-e353165-1dcb80-16b20f1d48e675; uuid=\"w:26bb242e8e4245199cf2016cf2b4dde7\"; RT=\"z=1&dm=toutiao.com&si=3vzxls89zcl&ss=jwhcyln8&sl=f&tt=0&obo=f&ld=2j1re&r=732c1fe0f3a07d695fc0790da6431fe8&ul=2j1ri&hd=2j1rn\"")
					for loop := aid["loop"].([2]int64); loop[0] < loop[1]; loop[0] += 86400 {
						ctx.AddQueue(&FKRequest.Request{
							Method: "GET",
							Url: "http://m.toutiao.com/list/?tag=" + ctx.GetKeywords() + "&ac=wap&count=20&format=json_raw&as=" +
								"A17538D54D106FF&cp=585DF0A65F0F1E1&min_behot_time=" + strconv.FormatInt(loop[0], 10),
							Rule:   aid["Rule"].(string),
							Header: header,
						})
					}
					return nil
				},

				ParseFunc: func(context *FKSpider.Context) {
					str := context.GetText()
					//FKLog.G_Log.Informational("收到 %s", str)
					var news_TouTiaoList News_TouTiaoList
					err := json.Unmarshal([]byte(str), &news_TouTiaoList)
					if err != nil {
						FKLog.G_Log.Informational("解析头条新闻列表错误： %v\n", err)
						return
					}

					newsLength := len(news_TouTiaoList.Data)
					//FKLog.G_Log.Informational("解析新闻列表 Len = %d", newsLength)
					for i := 0; i < newsLength; i++ {
						//FKLog.G_Log.Informational("GroupID= %s", new_TouTiaoList.Data[i].GroupID)
						context.AddQueue(&FKRequest.Request{
							Url:  "http://m.toutiao.com/i" + news_TouTiaoList.Data[i].GroupID + "/info/",
							Rule: "新闻内容",
							Temp: map[string]interface{}{
								"title":    strings.TrimSpace(news_TouTiaoList.Data[i].Title),
								"date":     strings.TrimSpace(news_TouTiaoList.Data[i].Datetime),
								"keywords": strings.TrimSpace(news_TouTiaoList.Data[i].Keywords),
							},
						})
					}
				},
			},

			"新闻内容": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"标题",
					"内容",
					"关键字",
					"时间",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					str := ctx.GetText()

					// 获取内容
					var article News_TouTiaoArticle
					err := json.Unmarshal([]byte(str), &article)
					if err != nil {
						log.Printf("解析错误： %v\n", err)
						return
					}
					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: ctx.GetTemp("title", ""),
						1: strings.TrimSpace(article.Data.Content),
						2: ctx.GetTemp("keywords", ""),
						3: ctx.GetTemp("date", ""),
					})
				},
			},
		},
	},
}