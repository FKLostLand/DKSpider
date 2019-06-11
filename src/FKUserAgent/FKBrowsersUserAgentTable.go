package FKUserAgent

import "FKConfig"

// 操作系统类型枚举
const (
	enumOsWindows = iota
	enumOsLinux
	enumOsMac
)

type (
	// 模板数据
	TemplateData struct {
		Name string // 浏览器名称
		Ver  string // 浏览器版本
		OSN  string // 操作系统名
		OSV  string // 操作系统版本
		Coms string // UserAgent自定义额外数据
	}
	// 系统属性数据
	OSAttributes struct {
		OSName    string   // 操作系统名
		OSVersion string   // 操作系统版本
		Comments  []string // 添加到UserAgent的自定义额外数据
	}
	// UserAgent 字符串表
	// Key是浏览器版本，Value是浏览器信息
	Formats map[string]string

	// 一种浏览器的UserAgent信息
	UAData struct {
		TopVersion string
		DefaultOS  int
		Formats    Formats
	}

	// 浏览器UserAgent数据表
	// Key是浏览器类型名称，Value是该浏览器对应的UserAgent 字符串表
	UATable map[string]UAData
)

// 默认操作系统属性
var DefaultOSAttributes = map[int]OSAttributes{
	enumOsWindows: {"Windows NT", "6.3", []string{"x64"}},
	enumOsLinux:   {"Linux", "3.16.1", []string{"x64"}},
	enumOsMac:     {"Intel Mac OS X", "10_6_8", []string{}},
}

// 全浏览器UA数据表
var globalBrowsersUserAgentTable = UATable{
	"chrome": {
		"37.0.2049.0",
		enumOsWindows,
		Formats{
			"37": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"36": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"35": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"34": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"33": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"32": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"31": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"30": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
		},
	},
	"firefox": {
		"31.0",
		enumOsWindows,
		Formats{
			"31": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:31.0) Gecko/20100101 Firefox/{{.Ver}}",
			"30": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:30.0) Gecko/20120101 Firefox/{{.Ver}}",
			"29": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:29.0) Gecko/20120101 Firefox/{{.Ver}}",
			"28": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:28.0) Gecko/20100101 Firefox/{{.Ver}}",
			"27": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:27.0) Gecko/20130101 Firefox/{{.Ver}}",
			"26": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:26.0) Gecko/20121011 Firefox/{{.Ver}}",
			"25": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:25.0) Gecko/20100101 Firefox/{{.Ver}}",
		},
	},
	"msie": {
		"10.0",
		enumOsWindows,
		Formats{
			"10": "Mozilla/5.0 (compatible; MSIE 10.0; {{.OSN}} {{.OSV}}{{if .Coms}}{{.Coms}}; {{end}}Trident/5.0; .NET CLR 3.5.30729)",
			"9":  "Mozilla/5.0 (compatible; MSIE 9.0; {{.OSN}} {{.OSV}}{{if .Coms}}{{.Coms}}; {{end}}Trident/5.0; .NET CLR 3.0.30729)",
			"8":  "Mozilla/5.0 (compatible; MSIE 8.0; {{.OSN}} {{.OSV}}{{if .Coms}}{{.Coms}}; {{end}}Trident/4.0; .NET CLR 3.0.04320)",
			"7":  "Mozilla/4.0 (compatible; MSIE 7.0; {{.OSN}} {{.OSV}}{{if .Coms}}{{.Coms}}; {{end}}.NET CLR 2.0.50727)",
		},
	},
	"opera": {
		"12.14",
		enumOsWindows,
		Formats{
			"12": "Opera/9.80 ({{.OSN}} {{.OSV}}; U{{.Coms}}) Presto/2.9.181 Version/{{.Ver}}",
			"11": "Opera/9.80 ({{.OSN}} {{.OSV}}; U{{.Coms}}) Presto/2.7.62 Version/{{.Ver}}",
			"10": "Opera/9.80 ({{.OSN}} {{.OSV}}; U{{.Coms}}) Presto/2.2.15 Version/{{.Ver}}",
			"9":  "Opera/9.00 ({{.OSN}} {{.OSV}}; U{{.Coms}})",
		},
	},
	"safari": {
		"6.0",
		enumOsMac,
		Formats{
			"6": "Mozilla/5.0 (Macintosh; {{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/536.26 (KHTML, like Gecko) Version/{{.Ver}} Safari/8536.25",
			"5": "Mozilla/5.0 (Macintosh; {{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/531.2+ (KHTML, like Gecko) Version/{{.Ver}} Safari/531.2+",
			"4": "Mozilla/5.0 (Macintosh; {{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/528.16 (KHTML, like Gecko) Version/{{.Ver}} Safari/528.16",
		},
	},
	"itunes": {
		"9.1.1",
		enumOsMac,
		Formats{
			"9": "iTunes/{{.Ver}}",
			"8": "iTunes/{{.Ver}}",
			"7": "iTunes/{{.Ver}} (Macintosh; U; PPC Mac OS X 10.4.7{{.Coms}})",
			"6": "iTunes/{{.Ver}} (Macintosh; U; PPC Mac OS X 10.4.5{{.Coms}})",
		},
	},
	"aol": {
		"9.7",
		enumOsWindows,
		Formats{
			"9": "Mozilla/5.0 (compatible; MSIE 9.0; AOL {{.Ver}}; AOLBuild 4343.19; {{.OSN}} {{.OSV}}; WOW64; Trident/5.0; FunWebProducts{{.Coms}})",
			"8": "Mozilla/4.0 (compatible; MSIE 7.0; AOL {{.Ver}}; {{.OSN}} {{.OSV}}; GTB5; .NET CLR 1.1.4322; .NET CLR 2.0.50727{{.Coms}})",
			"7": "Mozilla/4.0 (compatible; MSIE 7.0; AOL {{.Ver}}; {{.OSN}} {{.OSV}}; FunWebProducts{{.Coms}})",
			"6": "Mozilla/4.0 (compatible; MSIE 6.0; AOL {{.Ver}}; {{.OSN}} {{.OSV}}{{.Coms}})",
		},
	},
	"konqueror": {
		"4.9",
		enumOsLinux,
		Formats{
			"4": "Mozilla/5.0 (compatible; Konqueror/4.0; {{.OSN}}{{.Coms}}) KHTML/4.0.3 (like Gecko)",
			"3": "Mozilla/5.0 (compatible; Konqueror/3.0-rc6; i686 {{.OSN}}; 20021127{{.Coms}})",
			"2": "Mozilla/5.0 (compatible; Konqueror/2.1.1; {{.OSN}}{{.Coms}})",
		},
	},
	"netscape": {
		"9.1.0285",
		enumOsWindows,
		Formats{
			"9": "Mozilla/5.0 ({{.OSN}}; U; {{.OSN}} {{.OSV}}; rv:1.9.2.4{{.Coms}}) Gecko/20070321 Netscape/{{.Ver}}",
			"8": "Mozilla/5.0 ({{.OSN}}; U; {{.OSN}} {{.OSV}}; rv:1.7.5{{.Coms}}) Gecko/20050519 Netscape/{{.Ver}}",
			"7": "Mozilla/5.0 ({{.OSN}}; U; {{.OSN}} {{.OSV}}; rv:1.0.1{{.Coms}}) Gecko/20020921 Netscape/{{.Ver}}",
		},
	},
	"lynx": {
		"2.8.8dev.3",
		enumOsLinux,
		Formats{
			"2": "Lynx/{{.Ver}} libwww-FM/2.14 SSL-MM/1.4.1",
			"1": "Lynx (textmode)",
		},
	},
	"baidubot": {
		"2.0",
		enumOsLinux,
		Formats{
			"2": "Mozilla/5.0 (compatible; Baiduspider/{{.Ver}}; +http://www.baidu.com/search/spider.html{{.Coms}})",
			"1": "Baiduspider+{{.Ver}}(+http://www.baidu.com/search/spider_jp.html{{.Coms}})",
		},
	},
	"googlebot": {
		"2.1",
		enumOsLinux,
		Formats{
			"2": "Mozilla/5.0 (compatible; Googlebot/{{.Ver}}; +http://www.google.com/bot.html{{.Coms}})",
			"1": "Googlebot/{{.Ver}} (+http://www.google.com/bot.html{{.Coms}})",
		},
	},
	"bingbot": {
		"2.0",
		enumOsWindows,
		Formats{
			"2": "Mozilla/5.0 (compatible; bingbot/{{.Ver}}; +http://www.bing.com/bingbot.htm{{.Coms}})",
		},
	},
	"yahoobot": {
		"2.0",
		enumOsLinux,
		Formats{
			"2": "Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp{{.Coms}})",
		},
	},
	"default": {
		FKConfig.APP_VERSION,
		enumOsLinux,
		Formats{
			"1": "{{.Name}}/{{.Ver}} ({{.OSN}} {{.OSV}}{{.Coms}})",
		},
	},
}
