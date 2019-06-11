package FKSpider

import (
	"FKBase"
	"FKLog"
	"FKRequest"
	"FKTempDataPool"
	"FKTimer"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type Context struct {
	spider   *Spider                   // 规则
	Request  *FKRequest.Request        // 原始请求
	Response *http.Response            // 响应流，其中URL拷贝自*request.Request
	text     []byte                    // 下载内容Body的字节流格式
	dom      *goquery.Document         // 下载内容Body为html时，可转换为Dom的对象
	items    []FKTempDataPool.DataCell // 存放以文本形式输出的结果数据
	files    []FKTempDataPool.FileCell // 存放欲直接输出的文件("Name": string; "Body": io.ReadCloser)
	err      error                     // 错误标记
	sync.Mutex
}

var (
	GlobalContextPool = &sync.Pool{
		New: func() interface{} {
			return &Context{
				items: []FKTempDataPool.DataCell{},
				files: []FKTempDataPool.FileCell{},
			}
		},
	}
)

func GetContext(sp *Spider, req *FKRequest.Request) *Context {
	ctx := GlobalContextPool.Get().(*Context)
	ctx.spider = sp
	ctx.Request = req
	return ctx
}

func PutContext(ctx *Context) {
	if ctx.Response != nil {
		ctx.Response.Body.Close()
		ctx.Response = nil
	}
	ctx.items = ctx.items[:0]
	ctx.files = ctx.files[:0]
	ctx.spider = nil
	ctx.Request = nil
	ctx.text = nil
	ctx.dom = nil
	ctx.err = nil
	GlobalContextPool.Put(ctx)
}

func (c *Context) SetResponse(resp *http.Response) *Context {
	c.Response = resp
	return c
}

// 标记下载错误
func (c *Context) SetError(err error) {
	c.err = err
}

// 生成并添加请求至队列。
// Request.Url与Request.Rule必须设置。
// Request.Spider无需手动设置(由系统自动设置)。
// Request.EnableCookie在Spider字段中统一设置，规则请求中指定的无效。
// 以下字段有默认值，可不设置:
// Request.Method默认为GET方法;
// Request.DialTimeout默认为常量request.DefaultDialTimeout，小于0时不限制等待响应时长;
// Request.ConnTimeout默认为常量request.DefaultConnTimeout，小于0时不限制下载超时;
// Request.TryTimes默认为常量request.DefaultTryTimes，小于0时不限制失败重载次数;
// Request.RedirectTimes默认不限制重定向次数，小于0时可禁止重定向跳转;
// Request.RetryPause默认为常量request.DefaultRetryPause;
// Request.DownloaderID指定下载器ID，0为默认的Surf高并发下载器，功能完备，1为PhantomJS下载器，特点破防力强，速度慢，低并发。
// 默认自动补填Referer。
func (c *Context) AddQueue(req *FKRequest.Request) *Context {
	// 若已主动终止任务，则崩溃爬虫协程
	c.spider.tryPanic()

	err := req.
		SetSpiderName(c.spider.GetName()).
		SetEnableCookie(c.spider.GetEnableCookie()).
		Prepare()

	if err != nil {
		FKLog.G_Log.Error(err.Error())
		return c
	}

	// 自动设置Referer
	if req.GetReferer() == "" && c.Response != nil {
		req.SetReferer(c.GetUrl())
	}

	c.spider.RequestPush(req)
	return c
}

// 用于动态规则添加请求。
func (c *Context) JsAddQueue(jreq map[string]interface{}) *Context {
	// 若已主动终止任务，则崩溃爬虫协程
	c.spider.tryPanic()

	req := &FKRequest.Request{}
	u, ok := jreq["Url"].(string)
	if !ok {
		return c
	}
	req.Url = u
	req.Rule, _ = jreq["Rule"].(string)
	req.Method, _ = jreq["Method"].(string)
	req.Header = http.Header{}
	if header, ok := jreq["Header"].(map[string]interface{}); ok {
		for k, values := range header {
			if vals, ok := values.([]string); ok {
				for _, v := range vals {
					req.Header.Add(k, v)
				}
			}
		}
	}
	req.PostData, _ = jreq["PostData"].(string)
	req.Reloadable, _ = jreq["Reloadable"].(bool)
	if t, ok := jreq["DialTimeout"].(int64); ok {
		req.DialTimeout = time.Duration(t)
	}
	if t, ok := jreq["ConnTimeout"].(int64); ok {
		req.ConnTimeout = time.Duration(t)
	}
	if t, ok := jreq["RetryPause"].(int64); ok {
		req.RetryPause = time.Duration(t)
	}
	if t, ok := jreq["TryTimes"].(int64); ok {
		req.TryTimes = int(t)
	}
	if t, ok := jreq["RedirectTimes"].(int64); ok {
		req.RedirectTimes = int(t)
	}
	if t, ok := jreq["Priority"].(int64); ok {
		req.Priority = int(t)
	}
	if t, ok := jreq["DownloaderID"].(int64); ok {
		req.DownloaderID = int(t)
	}
	if t, ok := jreq["Temp"].(map[string]interface{}); ok {
		req.Temp = t
	}

	err := req.
		SetSpiderName(c.spider.GetName()).
		SetEnableCookie(c.spider.GetEnableCookie()).
		Prepare()

	if err != nil {
		FKLog.G_Log.Error(err.Error())
		return c
	}

	if req.GetReferer() == "" && c.Response != nil {
		req.SetReferer(c.GetUrl())
	}

	c.spider.RequestPush(req)
	return c
}

// 输出文本结果。
// item类型为map[int]interface{}时，根据ruleName现有的ItemFields字段进行输出，
// item类型为map[string]interface{}时，ruleName不存在的ItemFields字段将被自动添加，
// ruleName为空时默认当前规则。
func (c *Context) Output(item interface{}, ruleName ...string) {
	strRuleName, rule, found := c.getRule(ruleName...)
	if !found {
		FKLog.G_Log.Error("蜘蛛 %s 调用Output()时，指定的规则名不存在！", c.spider.GetName())
		return
	}
	var itemsMap map[string]interface{}
	switch item2 := item.(type) {
	case map[int]interface{}:
		itemsMap = c.CreateItem(item2, strRuleName)
	case FKRequest.RequestTempData:
		for k := range item2 {
			c.spider.UpsertItemField(rule, k)
		}
		itemsMap = item2
	case map[string]interface{}:
		for k := range item2 {
			c.spider.UpsertItemField(rule, k)
		}
		itemsMap = item2
	}
	c.Lock()
	if c.spider.NotDefaultField {
		c.items = append(c.items, FKTempDataPool.GetDataCell(strRuleName, itemsMap, "", "", ""))
	} else {
		c.items = append(c.items, FKTempDataPool.GetDataCell(strRuleName, itemsMap, c.GetUrl(), c.GetReferer(), time.Now().Format("2006-01-02 15:04:05")))
	}
	c.Unlock()
}

// 输出文件。
// nameOrExt指定文件名或仅扩展名，为空时默认保持原文件名（包括扩展名）不变。
func (c *Context) FileOutput(nameOrExt ...string) {
	// 读取完整文件流
	byteArray, err := ioutil.ReadAll(c.Response.Body)
	c.Response.Body.Close()
	if err != nil {
		panic(err.Error())
		return
	}

	// 智能设置完整文件名
	_, s := path.Split(c.GetUrl())
	n := strings.Split(s, "?")[0]

	var baseName, ext string

	if len(nameOrExt) > 0 {
		p, n := path.Split(nameOrExt[0])
		ext = path.Ext(n)
		if baseName2 := strings.TrimSuffix(n, ext); baseName2 != "" {
			baseName = p + baseName2
		}
	}
	if baseName == "" {
		baseName = strings.TrimSuffix(n, path.Ext(n))
	}
	if ext == "" {
		ext = path.Ext(n)
	}
	if ext == "" {
		ext = ".html"
	}

	// 保存到文件临时队列
	c.Lock()
	c.files = append(c.files, FKTempDataPool.GetFileCell(c.GetRuleName(), baseName+ext, byteArray))
	c.Unlock()
}

// 生成文本结果。
// 用ruleName指定匹配的ItemFields字段，为空时默认当前规则。
func (c *Context) CreateItem(item map[int]interface{}, ruleName ...string) map[string]interface{} {
	_, rule, found := c.getRule(ruleName...)
	if !found {
		FKLog.G_Log.Error("蜘蛛 %s 调用CreatItem()时，指定的规则名不存在！", c.spider.GetName())
		return nil
	}

	var item2 = make(map[string]interface{}, len(item))
	for k, v := range item {
		field := c.spider.GetItemField(rule, k)
		item2[field] = v
	}
	return item2
}

// 在请求中保存临时数据。
func (c *Context) SetTemp(key string, value interface{}) *Context {
	c.Request.SetTemp(key, value)
	return c
}

func (c *Context) SetUrl(url string) *Context {
	c.Request.Url = url
	return c
}

func (c *Context) SetReferer(referer string) *Context {
	c.Request.Header.Set("Referer", referer)
	return c
}

// 为指定Rule动态追加结果字段名，并获取索引位置，
// 已存在时获取原来索引位置，
// 若ruleName为空，默认为当前规则。
func (c *Context) UpsertItemField(field string, ruleName ...string) (index int) {
	_, rule, found := c.getRule(ruleName...)
	if !found {
		FKLog.G_Log.Error("蜘蛛 %s 调用UpsertItemField()时，指定的规则名不存在！", c.spider.GetName())
		return
	}
	return c.spider.UpsertItemField(rule, field)
}

// 调用指定Rule下辅助函数AidFunc()。
// 用ruleName指定匹配的AidFunc，为空时默认当前规则。
func (c *Context) Aid(aid map[string]interface{}, ruleName ...string) interface{} {
	// 若已主动终止任务，则崩溃爬虫协程
	c.spider.tryPanic()

	_, rule, found := c.getRule(ruleName...)
	if !found {
		if len(ruleName) > 0 {
			FKLog.G_Log.Error("调用蜘蛛 %s 不存在的规则: %s", c.spider.GetName(), ruleName[0])
		} else {
			FKLog.G_Log.Error("调用蜘蛛 %s 的Aid()时未指定的规则名", c.spider.GetName())
		}
		return nil
	}
	if rule.AidFunc == nil {
		FKLog.G_Log.Error("蜘蛛 %s 的规则 %s 未定义AidFunc", c.spider.GetName(), ruleName[0])
		return nil
	}
	return rule.AidFunc(c, aid)
}

// 解析响应流。
// 用ruleName指定匹配的ParseFunc字段，为空时默认调用Root()。
func (c *Context) Parse(ruleName ...string) *Context {
	// 若已主动终止任务，则崩溃爬虫协程
	c.spider.tryPanic()

	strRuleName, rule, found := c.getRule(ruleName...)
	if c.Response != nil {
		c.Request.SetRuleName(strRuleName)
	}
	if !found {
		c.spider.RuleTree.Root(c)
		return c
	}
	if rule.ParseFunc == nil {
		FKLog.G_Log.Error("蜘蛛 %s 的规则 %s 未定义ParseFunc", c.spider.GetName(), ruleName[0])
		return c
	}
	rule.ParseFunc(c)
	return c
}

// 设置自定义配置。
func (c *Context) SetKeywords(keywords string) *Context {
	c.spider.SetKeywords(keywords)
	return c
}

// 设置采集上限。
func (c *Context) SetLimit(max int) *Context {
	c.spider.SetLimit(int64(max))
	return c
}

// 自定义暂停区间(随机: Pausetime/2 ~ Pausetime*2)，优先级高于外部传参。
// 当且仅当runtime[0]为true时可覆盖现有值。
func (c *Context) SetPausetime(pause int64, runtime ...bool) *Context {
	c.spider.SetPausetime(pause, runtime...)
	return c
}

// 设置定时器，
// @id为定时器唯一标识，
// @bell==nil时为倒计时器，此时@tol为睡眠时长，
// @bell!=nil时为闹铃，此时@tol用于指定醒来时刻（从now起遇到的第tol个bell）。
func (c *Context) SetTimer(id string, tol time.Duration, bell *FKTimer.Bell) bool {
	return c.spider.SetTimer(id, tol, bell)
}

// 启动定时器，并获取定时器是否可以继续使用。
func (c *Context) RunTimer(id string) bool {
	return c.spider.RunTimer(id)
}

// 重置下载的文本内容，
func (c *Context) ResetText(body string) *Context {
	x := (*[2]uintptr)(unsafe.Pointer(&body))
	h := [3]uintptr{x[0], x[1], x[1]}
	c.text = *(*[]byte)(unsafe.Pointer(&h))
	c.dom = nil
	return c
}

// 获取下载错误。
func (c *Context) GetError() error {
	// 若已主动终止任务，则崩溃爬虫协程
	c.spider.tryPanic()
	return c.err
}

// 获取日志接口实例。
func (*Context) Log() FKLog.ILogMgr {
	return FKLog.G_Log
}

// 获取蜘蛛名称。
func (c *Context) GetSpider() *Spider {
	return c.spider
}

// 获取响应流。
func (c *Context) GetResponse() *http.Response {
	return c.Response
}

// 获取响应状态码。
func (c *Context) GetStatusCode() int {
	return c.Response.StatusCode
}

// 获取原始请求。
func (c *Context) GetRequest() *FKRequest.Request {
	return c.Request
}

// 获得一个原始请求的副本。
func (c *Context) CopyRequest() *FKRequest.Request {
	return c.Request.Copy()
}

// 获取结果字段名列表。
func (c *Context) GetItemFields(ruleName ...string) []string {
	_, rule, found := c.getRule(ruleName...)
	if !found {
		FKLog.G_Log.Error("蜘蛛 %s 调用GetItemFields()时，指定的规则名不存在！", c.spider.GetName())
		return nil
	}
	return c.spider.GetItemFields(rule)
}

// 由索引下标获取结果字段名，不存在时获取空字符串，
// 若ruleName为空，默认为当前规则。
func (c *Context) GetItemField(index int, ruleName ...string) (field string) {
	_, rule, found := c.getRule(ruleName...)
	if !found {
		FKLog.G_Log.Error("蜘蛛 %s 调用GetItemField()时，指定的规则名不存在！", c.spider.GetName())
		return
	}
	return c.spider.GetItemField(rule, index)
}

// 由结果字段名获取索引下标，不存在时索引为-1，
// 若ruleName为空，默认为当前规则。
func (c *Context) GetItemFieldIndex(field string, ruleName ...string) (index int) {
	_, rule, found := c.getRule(ruleName...)
	if !found {
		FKLog.G_Log.Error("蜘蛛 %s 调用GetItemField()时，指定的规则名不存在！", c.spider.GetName())
		return
	}
	return c.spider.GetItemFieldIndex(rule, field)
}

func (c *Context) PullItems() (ds []FKTempDataPool.DataCell) {
	c.Lock()
	ds = c.items
	c.items = []FKTempDataPool.DataCell{}
	c.Unlock()
	return
}

func (c *Context) PullFiles() (fs []FKTempDataPool.FileCell) {
	c.Lock()
	fs = c.files
	c.files = []FKTempDataPool.FileCell{}
	c.Unlock()
	return
}

// 获取自定义配置。
func (c *Context) GetKeywords() string {
	return c.spider.GetKeywords()
}

// 获取采集上限。
func (c *Context) GetLimit() int {
	return int(c.spider.GetLimit())
}

// 获取蜘蛛名。
func (c *Context) GetName() string {
	return c.spider.GetName()
}

// 获取规则树。
func (c *Context) GetRules() map[string]*Rule {
	return c.spider.GetRules()
}

// 获取指定规则。
func (c *Context) GetRule(ruleName string) (*Rule, bool) {
	return c.spider.GetRule(ruleName)
}

// 获取当前规则名。
func (c *Context) GetRuleName() string {
	return c.Request.GetRuleName()
}

// 获取请求中临时缓存数据
// defaultValue 不能为 interface{}(nil)
func (c *Context) GetTemp(key string, defaultValue interface{}) interface{} {
	return c.Request.GetTemp(key, defaultValue)
}

// 获取请求中全部缓存数据
func (c *Context) GetTemps() FKRequest.RequestTempData {
	return c.Request.GetTemps()
}

// 获得一个请求的缓存数据副本。
func (c *Context) CopyTemps() FKRequest.RequestTempData {
	temps := make(FKRequest.RequestTempData)
	for k, v := range c.Request.GetTemps() {
		temps[k] = v
	}
	return temps
}

// 从原始请求获取Url，从而保证请求前后的Url完全相等，且中文未被编码。
func (c *Context) GetUrl() string {
	return c.Request.Url
}

func (c *Context) GetMethod() string {
	return c.Request.GetMethod()
}

func (c *Context) GetHost() string {
	return c.Response.Request.URL.Host
}

// 获取响应头信息。
func (c *Context) GetHeader() http.Header {
	return c.Response.Header
}

// 获取请求头信息。
func (c *Context) GetRequestHeader() http.Header {
	return c.Response.Request.Header
}

func (c *Context) GetReferer() string {
	return c.Response.Request.Header.Get("Referer")
}

// 获取响应的Cookie。
func (c *Context) GetCookie() string {
	return c.Response.Header.Get("Set-Cookie")
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (c *Context) GetDom() *goquery.Document {
	if c.dom == nil {
		c.initDom()
	}
	return c.dom
}

// GetBodyStr returns plain string crawled.
func (c *Context) GetText() string {
	if c.text == nil {
		c.initText()
	}
	return FKBase.Bytes2String(c.text)
}

// 获取规则。
func (c *Context) getRule(ruleName ...string) (name string, rule *Rule, found bool) {
	if len(ruleName) == 0 {
		if c.Response == nil {
			return
		}
		name = c.GetRuleName()
	} else {
		name = ruleName[0]
	}
	rule, found = c.spider.GetRule(name)
	return
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (c *Context) initDom() *goquery.Document {
	if c.text == nil {
		c.initText()
	}
	var err error
	c.dom, err = goquery.NewDocumentFromReader(bytes.NewReader(c.text))
	if err != nil {
		panic(err.Error())
	}
	return c.dom
}

// GetBodyStr returns plain string crawled.
func (c *Context) initText() {
	var err error

	// 采用surf内核下载时，尝试自动转码
	if c.Request.DownloaderID == FKRequest.SurfID {
		var contentType, pageEncode string
		// 优先从响应头读取编码类型
		contentType = c.Response.Header.Get("Content-Type")
		if _, params, err := mime.ParseMediaType(contentType); err == nil {
			if cs, ok := params["charset"]; ok {
				pageEncode = strings.ToLower(strings.TrimSpace(cs))
			}
		}
		// 响应头未指定编码类型时，从请求头读取
		if len(pageEncode) == 0 {
			contentType = c.Request.Header.Get("Content-Type")
			if _, params, err := mime.ParseMediaType(contentType); err == nil {
				if cs, ok := params["charset"]; ok {
					pageEncode = strings.ToLower(strings.TrimSpace(cs))
				}
			}
		}

		switch pageEncode {
		// 不做转码处理
		case "utf8", "utf-8", "unicode-1-1-utf-8":
		default:
			// 指定了编码类型，但不是utf8时，自动转码为utf8
			// get converter to utf-8
			// Charset auto determine. Use golang.org/x/net/html/charset. Get response body and change it to utf-8
			var destReader io.Reader

			if len(pageEncode) == 0 {
				destReader, err = charset.NewReader(c.Response.Body, "")
			} else {
				destReader, err = charset.NewReaderLabel(pageEncode, c.Response.Body)
			}

			if err == nil {
				c.text, err = ioutil.ReadAll(destReader)
				if err == nil {
					c.Response.Body.Close()
					return
				} else {
					FKLog.G_Log.Warning(" *     [convert][%v]: %v (ignore transcoding)\n", c.GetUrl(), err)
				}
			} else {
				FKLog.G_Log.Warning(" *     [convert][%v]: %v (ignore transcoding)\n", c.GetUrl(), err)
			}
		}
	}

	// 不做转码处理
	c.text, err = ioutil.ReadAll(c.Response.Body)
	c.Response.Body.Close()
	if err != nil {
		panic(err.Error())
		return
	}

}
