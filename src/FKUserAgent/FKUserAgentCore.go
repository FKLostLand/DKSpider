package FKUserAgent

import (
	"FKConfig"
	"FKLog"
	"bytes"
	"strings"
	"text/template"
)

// 获取指定浏览器版本
func getBrowserTopVersion(browserName string) string {
	browserName = strings.ToLower(browserName)
	data, ok := globalBrowsersUserAgentTable[browserName]
	if ok {
		return data.TopVersion
	}
	return globalBrowsersUserAgentTable["default"].TopVersion
}

func format(browserName, browserVersion string) string {
	browserName = strings.ToLower(browserName)
	majVer := strings.Split(browserVersion, ".")[0]
	data, ok := globalBrowsersUserAgentTable[browserName]
	if ok {
		format, ok := data.Formats[majVer]
		if ok {
			return format
		} else {
			top := getBrowserTopVersion(browserName)
			majVer = strings.Split(top, ".")[0]
			return data.Formats[majVer]
		}
	}

	if browserName != FKConfig.APP_NAME {
		FKLog.G_Log.Notice("Create an unknown webBrowser'UA : %s", browserName)
	}
	return globalBrowsersUserAgentTable["default"].Formats["1"]
}

// 根据条件生成UA
func createFromDetails(bname, bver, osname, osver string, c []string) string {
	// 若不指定版本，则使用自定义版本
	if bver == "" {
		bver = getBrowserTopVersion(bname)
	}
	// 填充自定义数据
	comments := strings.Join(c, "; ")
	if comments != "" {
		comments = "; " + comments
	}

	data := TemplateData{bname, bver, osname, osver, comments}
	buff := &bytes.Buffer{}
	t := template.New("formatter")
	t.Parse(format(bname, bver))
	t.Execute(buff, data)

	return buff.String()
}
