package FKUserAgent

import (
	"reflect"
	"strings"
	"testing"
)

func TestCreateUA(t *testing.T) {
	// 输出全局Key
	keys := reflect.ValueOf(globalUserAgents).MapKeys()
	strkeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}
	t.Log("Keys:  " + strings.Join(strkeys, ","))
	t.Log(globalUserAgents["common"][0])

	// 查看真实UA
	t.Log(GlobalUserAgent.CreateRealUA())
	// 创建一些伪UA
	t.Log(GlobalUserAgent.CreateUAByBrowserType("baidubot"))
	t.Log(GlobalUserAgent.CreateUAByBrowserTypeAndVersion("msie", "8.0"))
	// 随机测试
	t.Log(GlobalUserAgent.CreateRandomUA())
	t.Log(GlobalUserAgent.CreateRandomWebBrowserUA())
	// 非法测试
	t.Log(GlobalUserAgent.CreateUAByBrowserType("unexist"))
}

func BenchmarkCreateRandomUA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GlobalUserAgent.CreateRandomWebBrowserUA()
	}
	b.Log(GlobalUserAgent.CreateRandomWebBrowserUA())
	for i := 0; i < b.N; i++ {
		GlobalUserAgent.CreateStaticWebBrowserUA()
	}
	b.Log(GlobalUserAgent.CreateStaticWebBrowserUA())
}
