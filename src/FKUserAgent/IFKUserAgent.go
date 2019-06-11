package FKUserAgent

/*
	负责生成HTTP UserAgent
*/
type IFKUserAgent interface {
	// 生成一个随机UserAgent
	CreateRandomUA() string
	// 生成一个随机浏览器类型UserAgent
	// comment:(不包含爬虫类UserAgent)
	CreateRandomWebBrowserUA() string
	// 生成一个固定的浏览器类型UserAgent
	// comment:(不包含爬虫类UserAgent)
	CreateStaticWebBrowserUA() string
	// 创建本进程真实的Agent头
	// comment:（部分情况用作特殊测试）
	CreateRealUA() string
	// 创建一个指定浏览器的Agent头
	// browser类型: 基本类型 "chrome", "firefox", "msie", "opera", "safari", "aol"， "konqueror"， "netscape"
	// 				特殊类型  "itunes", "lynx"， "baidubot"， "googlebot"，"bingbot"， "yahoobot"
	CreateUAByBrowserType(browser string) string
	//  创建一个指定浏览器和版本的Agent头
	// browser类型: 基本类型 "chrome", "firefox", "msie", "opera", "safari", "aol"， "konqueror"， "netscape"
	// 				特殊类型  "itunes", "lynx"， "baidubot"， "googlebot"，"bingbot"， "yahoobot"
	// version: 字符串，允许各种格式，如"1.0", "2.0.1", "3.1.2dev.4", "37.0.2049.1"均可
	CreateUAByBrowserTypeAndVersion(browser, version string) string
}

var (
	// 对外接口
	GlobalUserAgent = createFKUserAgent()
)
