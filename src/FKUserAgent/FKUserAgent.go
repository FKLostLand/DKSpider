package FKUserAgent

import (
	"FKConfig"
	"math/rand"
	"runtime"
	"strings"
	"time"
)

type FKUserAgent struct {
	UserAgents map[string][]string // 全部UserAgent
}

var globalUserAgents = map[string][]string{}

// 创建接口实例
func createFKUserAgent() IFKUserAgent {
	return &FKUserAgent{
		UserAgents: globalUserAgents,
	}
}

// 创建本进程真实的Agent头
// comment:（部分情况用作特殊测试）
func (p *FKUserAgent) CreateRealUA() string {
	return createFromDetails(FKConfig.APP_NAME, FKConfig.APP_VERSION, osName(), osVersion(), []string{runtime.Version()})
}

// 创建一个指定浏览器的Agent头
// browser类型: 基本类型 "chrome", "firefox", "msie", "opera", "safari", "aol"， "konqueror"， "netscape"
// 				特殊类型  "itunes", "lynx"， "baidubot"， "googlebot"，"bingbot"， "yahoobot"
func (p *FKUserAgent) CreateUAByBrowserType(browser string) string {
	bn := strings.ToLower(browser)
	data := globalBrowsersUserAgentTable[bn]
	os := data.DefaultOS
	osAttribs := DefaultOSAttributes[os]

	return createFromDetails(
		browser,
		data.TopVersion,
		osAttribs.OSName,
		osAttribs.OSVersion,
		osAttribs.Comments)
}

//  创建一个指定浏览器和版本的Agent头
// browser类型: 基本类型 "chrome", "firefox", "msie", "opera", "safari", "aol"， "konqueror"， "netscape"
// 				特殊类型  "itunes", "lynx"， "baidubot"， "googlebot"，"bingbot"， "yahoobot"
// version: 字符串，允许各种格式，如"1.0", "2.0.1", "3.1.2dev.4", "37.0.2049.1"均可
func (p *FKUserAgent) CreateUAByBrowserTypeAndVersion(browser, version string) string {
	bn := strings.ToLower(browser)
	data := globalBrowsersUserAgentTable[bn]
	os := data.DefaultOS
	osAttribs := DefaultOSAttributes[os]

	return createFromDetails(
		browser,
		version,
		osAttribs.OSName,
		osAttribs.OSVersion,
		osAttribs.Comments)
}

// 创建一个随机UA
func (p *FKUserAgent) CreateRandomUA() string {
	lAll := len(p.UserAgents["all"])
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idxAll := r.Intn(lAll)

	return p.UserAgents["all"][idxAll]
}

// 生成一个随机浏览器类型UserAgent
// comment:(不包含爬虫类UserAgent)
func (p *FKUserAgent) CreateRandomWebBrowserUA() string {
	l := len(p.UserAgents["common"])
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Intn(l)

	return p.UserAgents["common"][idx]
}

// 生成一个固定的浏览器类型UserAgent
// comment:(不包含爬虫类UserAgent)
func (p *FKUserAgent) CreateStaticWebBrowserUA() string {
	l := len(p.UserAgents["common"])
	if l > 0 {
		return p.UserAgents["common"][0]
	}
	return p.CreateRealUA()
}

// 初始化创建 globalUserAgents
func init() {
	for browser, userAgentData := range globalBrowsersUserAgentTable {
		// 默认版本浏览器不做添加
		if browser == "default" {
			continue
		}

		// 填充组装 globalUserAgents
		os := userAgentData.DefaultOS
		osAttribs := DefaultOSAttributes[os]
		for version := range userAgentData.Formats {
			ua := createFromDetails(
				browser,
				version,
				osAttribs.OSName,
				osAttribs.OSVersion,
				osAttribs.Comments)
			globalUserAgents["all"] = append(globalUserAgents["all"], ua)

			if browser != "itunes" && browser != "lynx" && browser != "googlebot" &&
				browser != "bingbot" && browser != "yahoobot" && browser != "baidubot" {
				globalUserAgents["common"] = append(globalUserAgents["common"], ua)
			}
		}
	}

	lCommon := len(globalUserAgents["common"])
	lAll := len(globalUserAgents["all"])
	//FKLog.G_Log.App("Init globalUserAgents: Common UA'len = %d, All UA'len = %d", lCommon, lAll)

	// 随机交换UserAgent做混淆
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idxCommon := r.Intn(lCommon)
	idxAll := r.Intn(lAll)

	globalUserAgents["all"][0], globalUserAgents["all"][idxAll] = globalUserAgents["all"][idxAll], globalUserAgents["all"][0]
	globalUserAgents["common"][0], globalUserAgents["common"][idxCommon] = globalUserAgents["common"][idxCommon], globalUserAgents["common"][0]
}
