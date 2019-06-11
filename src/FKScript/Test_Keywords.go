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
规则参考：https://github.com/iMeiji/Toutiao/wiki/%E4%BB%8A%E6%97%A5%E5%A4%B4%E6%9D%A1Api%E5%88%86%E6%9E%90
KeyWord值请设置为：<news_car><news_game>
 */

func init(){
	Test_Keywords_Spider.RegisterToSpiderSpecies()
}

var Test_Keywords_Spider = &FKSpider.Spider{
	Name:        "测试站_自定义配置测试",
	Description: "【开启Keywords】测试今日头条新闻分配下载，KeyWord值请设置为：<news_car><news_game>",
	EnableCookie: false,
	Keywords: FKBase.KEYWORDS,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context){
			// https://tool.chinaz.com/Tools/unixtime.aspx
			// 1370344418 2013-06-04
			// 1559026838 2019-05-28
			timestamp := time.Now().Unix()
			ctx.Aid(map[string]interface{}{
				"loop": [2]int64{1559026838, timestamp},
				"Rule": "生成请求",
			}, "生成请求")
		},

		Trunk: map[string]* FKSpider.Rule{
			"生成请求":{
				AidFunc: func(ctx *FKSpider.Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int64); loop[0] < loop[1]; loop[0] += 86400 {
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
					var new_TouTiaoList News_TouTiaoList
					err := json.Unmarshal([]byte(str), &new_TouTiaoList)
					if err != nil {
						FKLog.G_Log.Informational("解析头条新闻列表错误： %v\n", err)
						return
					}

					newsLength := len(new_TouTiaoList.Data)
					//FKLog.G_Log.Informational("解析新闻列表 Len = %d", newsLength)
					for i := 0; i < newsLength; i++ {
						//FKLog.G_Log.Informational("GroupID= %s", new_TouTiaoList.Data[i].GroupID)
						context.AddQueue(&FKRequest.Request{
							Url:  "http://m.toutiao.com/i" + new_TouTiaoList.Data[i].GroupID + "/info/",
							Rule: "新闻内容",
							Temp: map[string]interface{}{
								"title":    strings.TrimSpace(new_TouTiaoList.Data[i].Title),
								"date":     strings.TrimSpace(new_TouTiaoList.Data[i].Datetime),
								"keywords": strings.TrimSpace(new_TouTiaoList.Data[i].Keywords),
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