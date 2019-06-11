package FKRequest

import (
	"FKBase"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	SurfID             = 0               // Surf下载器标识符(默认Go原生的surf下载内核，此值不可改动)
	PhomtomJsID        = 1               // PhomtomJs下载器标识符（备用的phantomjs下载内核，一般不使用，效率差，头信息支持不完善）
	DefaultDialTimeout = 2 * time.Minute // 默认请求服务器超时
	DefaultConnTimeout = 2 * time.Minute // 默认下载超时
	DefaultTryTimes    = 3               // 默认最大下载次数
	DefaultRetryPause  = 2 * time.Second // 默认重新下载前停顿时长
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

// 发送请求之前的准备工作，修正检查一系列参数值
func (r *Request) Prepare() error {
	URL, err := url.Parse(r.Url)
	if err != nil {
		return err
	}
	r.Url = URL.String()

	if r.Method == "" {
		r.Method = "GET"
	} else {
		r.Method = strings.ToUpper(r.Method)
	}

	if r.Header == nil {
		r.Header = make(http.Header)
	}

	if r.DialTimeout < 0 {
		r.DialTimeout = 0
	} else if r.DialTimeout == 0 {
		r.DialTimeout = DefaultDialTimeout
	}

	if r.ConnTimeout < 0 {
		r.ConnTimeout = 0
	} else if r.ConnTimeout == 0 {
		r.ConnTimeout = DefaultConnTimeout
	}

	if r.TryTimes == 0 {
		r.TryTimes = DefaultTryTimes
	}

	if r.RetryPause <= 0 {
		r.RetryPause = DefaultRetryPause
	}

	if r.Priority < 0 {
		r.Priority = 0
	}

	if r.DownloaderID < SurfID || r.DownloaderID > PhomtomJsID {
		r.DownloaderID = SurfID
	}

	if r.TempIsJson == nil {
		r.TempIsJson = make(map[string]bool)
	}

	if r.Temp == nil {
		r.Temp = make(RequestTempData)
	}
	return nil
}

// 反序列化
func UnSerialize(s string) (*Request, error) {
	req := new(Request)
	return req, json.Unmarshal([]byte(s), req)
}

// 序列化
func (r *Request) Serialize() string {
	for k, v := range r.Temp {
		r.Temp.set(k, v)
		r.TempIsJson[k] = true
	}
	b, _ := json.Marshal(r)
	return strings.Replace(FKBase.Bytes2String(b), `\u0026`, `&`, -1)
}

// 请求的唯一识别码
func (r *Request) Unique() string {
	if r.unique == "" {
		block := md5.Sum([]byte(r.Spider + r.Rule + r.Url + r.Method))
		r.unique = hex.EncodeToString(block[:])
	}
	return r.unique
}

// 获取副本
func (r *Request) Copy() *Request {
	reqcopy := new(Request)
	b, _ := json.Marshal(r)
	json.Unmarshal(b, reqcopy)
	return reqcopy
}

// 获取Url
func (r *Request) GetUrl() string {
	return r.Url
}

// 获取Http请求的方法名称
func (r *Request) GetMethod() string {
	return r.Method
}

// 设定Http请求方法的类型
func (r *Request) SetMethod(method string) *Request {
	r.Method = strings.ToUpper(method)
	return r
}

func (r *Request) SetUrl(url string) *Request {
	r.Url = url
	return r
}

func (r *Request) GetReferer() string {
	return r.Header.Get("Referer")
}

func (r *Request) SetReferer(referer string) *Request {
	r.Header.Set("Referer", referer)
	return r
}

func (r *Request) GetPostData() string {
	return r.PostData
}

func (r *Request) GetHeader() http.Header {
	return r.Header
}

func (r *Request) SetHeader(key, value string) *Request {
	r.Header.Set(key, value)
	return r
}

func (r *Request) AddHeader(key, value string) *Request {
	r.Header.Add(key, value)
	return r
}

func (r *Request) GetEnableCookie() bool {
	return r.EnableCookie
}

func (r *Request) SetEnableCookie(enableCookie bool) *Request {
	r.EnableCookie = enableCookie
	return r
}

func (r *Request) GetCookies() string {
	return r.Header.Get("Cookie")
}

func (r *Request) SetCookies(cookie string) *Request {
	r.Header.Set("Cookie", cookie)
	return r
}

func (r *Request) GetDialTimeout() time.Duration {
	return r.DialTimeout
}

func (r *Request) GetConnTimeout() time.Duration {
	return r.ConnTimeout
}

func (r *Request) GetTryTimes() int {
	return r.TryTimes
}

func (r *Request) GetRetryPause() time.Duration {
	return r.RetryPause
}

func (r *Request) GetProxy() string {
	return r.proxy
}

func (r *Request) SetProxy(proxy string) *Request {
	r.proxy = proxy
	return r
}

func (r *Request) GetRedirectTimes() int {
	return r.RedirectTimes
}

func (r *Request) GetRuleName() string {
	return r.Rule
}

func (r *Request) SetRuleName(ruleName string) *Request {
	r.Rule = ruleName
	return r
}

func (r *Request) GetSpiderName() string {
	return r.Spider
}

func (r *Request) SetSpiderName(spiderName string) *Request {
	r.Spider = spiderName
	return r
}

func (r *Request) IsReloadable() bool {
	return r.Reloadable
}

func (r *Request) SetReloadable(can bool) *Request {
	r.Reloadable = can
	return r
}

// 获取临时缓存数据
// defaultValue 不能为 nil
func (r *Request) GetTemp(key string, defaultValue interface{}) interface{} {
	if defaultValue == nil {
		panic("In FKRequest.GetTemp(), param defaultValue shouldn't be nil，key = " + key)
	}
	r.lock.RLock()
	defer r.lock.RUnlock()

	if r.Temp[key] == nil {
		return defaultValue
	}

	if r.TempIsJson[key] {
		return r.Temp.get(key, defaultValue)
	}

	return r.Temp[key]
}

func (r *Request) SetTemp(key string, value interface{}) *Request {
	r.lock.Lock()
	r.Temp[key] = value
	delete(r.TempIsJson, key)
	r.lock.Unlock()
	return r
}

func (r *Request) GetTemps() RequestTempData {
	return r.Temp
}

func (r *Request) SetTemps(temp map[string]interface{}) *Request {
	r.lock.Lock()

	r.Temp = temp
	r.TempIsJson = make(map[string]bool)

	r.lock.Unlock()
	return r
}

func (r *Request) GetPriority() int {
	return r.Priority
}

func (r *Request) SetPriority(priority int) *Request {
	r.Priority = priority
	return r
}

func (r *Request) GetDownloaderID() int {
	return r.DownloaderID
}

func (r *Request) SetDownloaderID(id int) *Request {
	r.DownloaderID = id
	return r
}

func (r *Request) MarshalJSON() ([]byte, error) {
	for k, v := range r.Temp {
		if r.TempIsJson[k] {
			continue
		}
		r.Temp.set(k, v)
		r.TempIsJson[k] = true
	}
	b, err := json.Marshal(*r)
	return b, err
}
