package FKScript


/*

// 使用init初始化来进行爬虫规则注册。
func init() {
	FileTest_Spider.RegisterToSpiderSpecies()
}

var FileTest_Spider = &FKSpider.Spider{
	Name:        "爬虫下载测试",  				// 爬虫规则名称，必须设置，必须全局唯一。
	Description: "这是描述文字",					// 爬虫规则描述，用来在规则表中给用户显示
	MedianPauseTime: 100,						// 两次采集的间隔时间。通常在UI中统一设置，在规则中无需额外设置
	Keywords: FKBase.KEYWORDS,					// 规则关键字，若设置了，则表示启动关键字。若未设置，则表示不使用关键字
	RequestLimit: math.MaxInt64,				// 采集次数最大限制。通常在UI中统一设置，在规则中无需额外设置
	EnableCookie: false,						// 该值为ture表示支持登录功能。该值为false表示会随机轮换UA
	RuleTree: &FKSpider.RuleTree{				// 规则的核心部分，包含全部规则解析
		Root: func(ctx *FKSpider.Context) {		// Root: 规则树 树根。即采集规则的入口函数
			ctx.AddQueue(&FKRequest.Request{	// 下载请求对象
				Url:          "http://xx.png",	// 下载请求对象的路径
				Rule:         "百度图片",		// 下载响应的解析Rule名称
				ConnTimeout:  -1,				// 下载超时时间(秒)，小于0表示不限制
				DownloaderID: 0, 				// 下载器方式：图片等多媒体文件必须使用0（surf go原生下载器），其他可以用1.
			})
		},
		Trunk: 									// Trunk: 规则树 树干。可包含大量规则Rule
			map[string]* FKSpider.Rule{			// Rule: 单条解析规则
			"百度图片": {						// Rule's name: 当其他请求进行规则调用时需要
				ItemFields: []string{			// Rule's ItemFields: 这个是输出字段，在非数据库输出方式（Excel/Csv）时作为标题行
					"标题",
					"价格",
				},
				ParseFunc: func(ctx * FKSpider.Context) {		// Rule's ParseFunc： 下载内容的解析函数，即下载到的数据该如何处理
					ctx.FileOutput("baidu")
				},
				AidFunc: func(ctx *FKSpider.Context, aid map[string]interface{}) interface{} {  // Rules's AidFunc: 规则的辅助函数，你可以自由定制各种处理方法，最常见的是生成批量请求
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&FKRequest.Request{
							Url:  "http://www.baidu.com/user=" + strconv.Itoa(loop[0]) + ".html",
							Rule: aid["Rule"].(string),
						})
					}
					return nil
				},
			},
		},
	},
}
*/

/*
// 供参考的数据结构。

type (
	//采集规则树
	RuleTree struct {
		Root  func(*Context)   // 根节点(执行入口)
		Trunk map[string]*Rule // 节点散列表(执行采集过程)
	}
	// 采集规则节点
	Rule struct {
		ItemFields []string                                           // 结果字段列表(选填，写上可保证字段顺序)
		ParseFunc  func(*Context)                                     // 内容解析函数
		AidFunc    func(*Context, map[string]interface{}) interface{} // 通用辅助函数
	}
	// 蜘蛛对象
	Spider struct {
		Name            string                                                  // 蜘蛛名称
		Description     string                                                  // 蜘蛛描述文字
		MedianPauseTime int64                                                   // 暂停时间中位数。优先使用配置参数，其次使用界面参数。
		RequestLimit    int64                                                   // 请求数限制。0为不限。优先使用配置参数，其次使用界面参数。
		Keywords        string                                                  // 自定义关键字，使用前必须设置规则初始值为  KEYWORDS
		EnableCookie    bool                                                    // 是否使用cookie记录
		NotDefaultField bool                                                    // 是否向输出结果中添加默认字段（例如url, downloadTime等）
		Namespace       func(s *Spider) string                                  // 命名空间，用来输出文件，路径命名
		SubNamespace    func(s *Spider, dataCell map[string]interface{}) string // 次级命名空间，用来输出文件，路径命名
		RuleTree        *RuleTree                                               // 采集规则树

		// 以下字段系统自动赋值
		id        int                 // 自动分配的SpiderQueue中的索引
		subName   string              // 由Keywords转换为的二级标识名
		reqMatrix *FKScheduler.Matrix // 请求矩阵
		timer     *FKTimer.Timer      // 定时器
		status    int                 // 执行状态
		lock      sync.RWMutex
		once      sync.Once
	}
)
type Request struct {
	Spider        string          // 规则名，系统自动设置，禁止人为填写
	Url           string          // 目标URL，必须设置
	Rule          string          // 用于解析响应的规则节点名，必须设置
	Method        string          // GET POST POST-M HEAD， 默认GET
	Header        http.Header     // 请求头信息
	EnableCookie  bool            // 是否使用cookies，在Spider的EnableCookie设置
	PostData      string          // POST values
	DialTimeout   time.Duration   // 创建连接，请求服务器超时，小于0表示不限制
	ConnTimeout   time.Duration   // 连接状态，页面下载超时，小于0表示不限制
	TryTimes      int             // 尝试下载的最大次数，小于0表示无限重试
	RetryPause    time.Duration   // 下载失败后，下次尝试下载的等待时间
	RedirectTimes int             // 重定向的最大次数，为0时不限，小于0时禁止重定向
	Temp          RequestTempData // 临时数据
	TempIsJson    map[string]bool // 将Temp中以JSON存储的字段标记为true，自动设置，禁止人为填写
	Priority      int             // 指定调度优先级，默认为0（最小优先级为0）
	Reloadable    bool            // 是否允许重复该链接下载
	DownloaderID  int             // 下载器内核ID: SurfID(高并发下载器，各种控制功能齐全) or PhomtomJsID(破防力强，速度慢，低并发)

	proxy  string // 当用户界面设置可使用代理IP时，自动设置代理
	unique string // 唯一ID
	lock   sync.RWMutex
}
 */