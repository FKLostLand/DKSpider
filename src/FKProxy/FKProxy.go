package FKProxy

import (
	"FKConfig"
	"FKDownloader/FKDownloaderWebBrowser"
	"FKLog"
	"FKPing"
	"FKRequest"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Proxy struct {
	ipRegexp           *regexp.Regexp
	proxyIPTypeRegexp  *regexp.Regexp
	proxyUrlTypeRegexp *regexp.Regexp
	allIps             map[string]string
	all                map[string]bool
	online             int32
	usable             map[string]*ProxyHost
	ticker             *time.Ticker
	tickMinute         int64
	threadPool         chan bool
	surf               FKDownloaderWebBrowser.DownloaderWebBrowser
	sync.Mutex
}

const (
	CONN_TIMEOUT = 4 //4s
	DAIL_TIMEOUT = 4 //4s
	TRY_TIMES    = 3
	// IP测速的最大并发量
	MAX_THREAD_NUM = 1000
)

func CreateProxy() *Proxy {
	p := &Proxy{
		ipRegexp:           regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`),
		proxyIPTypeRegexp:  regexp.MustCompile(`https?://([\w]*:[\w]*@)?[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+`),
		proxyUrlTypeRegexp: regexp.MustCompile(`((https?|ftp):\/\/)?(([^:\n\r]+):([^@\n\r]+)@)?((www\.)?([^/\n\r:]+)):?([0-9]{1,5})?\/?([^?\n\r]+)?\??([^#\n\r]*)?#?([^\n\r]*)`),
		allIps:             map[string]string{},
		all:                map[string]bool{},
		usable:             make(map[string]*ProxyHost),
		threadPool:         make(chan bool, MAX_THREAD_NUM),
		surf:               FKDownloaderWebBrowser.CreateSurfWebBrowser(),
	}
	go p.Update()
	return p
}

// 代理IP数量
func (p *Proxy) Count() int32 {
	return p.online
}

// 更新代理IP列表
func (p *Proxy) Update() *Proxy {
	f, err := os.Open(FKConfig.PROXY_LIB_FILE_PATH)
	if err != nil {
		return p
	}
	b, _ := ioutil.ReadAll(f)
	f.Close()

	proxysIPType := p.proxyIPTypeRegexp.FindAllString(string(b), -1)
	for _, proxy := range proxysIPType {
		p.allIps[proxy] = p.ipRegexp.FindString(proxy)
		p.all[proxy] = false
	}

	proxysUrlType := p.proxyUrlTypeRegexp.FindAllString(string(b), -1)
	for _, proxy := range proxysUrlType {
		gvalue := p.proxyUrlTypeRegexp.FindStringSubmatch(proxy)
		p.allIps[proxy] = gvalue[6]
		p.all[proxy] = false
	}

	log.Printf(" *     读取代理IP: %v 条\n", len(p.all))

	p.findOnline()

	return p
}

// 筛选在线的代理IP
func (p *Proxy) findOnline() *Proxy {
	log.Printf(" *     正在筛选在线的代理IP……")
	p.online = 0
	for proxy := range p.all {
		p.threadPool <- true
		go func(proxy string) {
			alive, _, _ := FKPing.Ping(p.allIps[proxy], CONN_TIMEOUT)
			p.Lock()
			p.all[proxy] = alive
			p.Unlock()
			if alive {
				atomic.AddInt32(&p.online, 1)
			}
			<-p.threadPool
		}(proxy)
	}
	for len(p.threadPool) > 0 {
		time.Sleep(0.2e9)
	}
	p.online = atomic.LoadInt32(&p.online)
	log.Printf(" *     在线代理IP筛选完成，共计：%v 个\n", p.online)

	return p
}

// 更新继时器
func (p *Proxy) UpdateTicker(tickMinute int64) {
	p.tickMinute = tickMinute
	p.ticker = time.NewTicker(time.Duration(p.tickMinute) * time.Minute)
	for _, proxyForHost := range p.usable {
		proxyForHost.curIndex++
		proxyForHost.isEcho = true
	}
}

// 获取本次循环中未使用的代理IP及其响应时长
func (p *Proxy) GetOne(u string) (curProxy string) {
	if p.online == 0 {
		return
	}
	u2, _ := url.Parse(u)
	if u2.Host == "" {
		FKLog.G_Log.Informational(" *     [%v]设置代理IP失败，目标url不正确\n", u)
		return
	}
	var key = u2.Host
	if strings.Count(key, ".") > 1 {
		key = key[strings.Index(key, ".")+1:]
	}

	p.Lock()
	defer p.Unlock()

	var ok = true
	var proxyForHost = p.usable[key]

	select {
	case <-p.ticker.C:
		proxyForHost.curIndex++
		if proxyForHost.curIndex >= proxyForHost.Len() {
			_, ok = p.testAndSort(key, u2.Scheme+"://"+u2.Host)
		}
		proxyForHost.isEcho = true

	default:
		if proxyForHost == nil {
			p.usable[key] = &ProxyHost{
				proxys:    []string{},
				timedelay: []time.Duration{},
				isEcho:    true,
			}
			proxyForHost, ok = p.testAndSort(key, u2.Scheme+"://"+u2.Host)
		} else if l := proxyForHost.Len(); l == 0 {
			ok = false
		} else if proxyForHost.curIndex >= l {
			_, ok = p.testAndSort(key, u2.Scheme+"://"+u2.Host)
			proxyForHost.isEcho = true
		}
	}
	if !ok {
		FKLog.G_Log.Informational(" *     [%v]设置代理IP失败，没有可用的代理IP\n", key)
		return
	}
	curProxy = proxyForHost.proxys[proxyForHost.curIndex]
	if proxyForHost.isEcho {
		FKLog.G_Log.Informational(" *     设置代理IP为 [%v](%v)\n",
			curProxy,
			proxyForHost.timedelay[proxyForHost.curIndex],
		)
		proxyForHost.isEcho = false
	}
	return
}

// 测试并排序
func (p *Proxy) testAndSort(key string, testHost string) (*ProxyHost, bool) {
	FKLog.G_Log.Informational(" *     [%v]正在测试与排序代理IP……", key)
	proxyForHost := p.usable[key]
	proxyForHost.proxys = []string{}
	proxyForHost.timedelay = []time.Duration{}
	proxyForHost.curIndex = 0
	for proxy, online := range p.all {
		if !online {
			continue
		}
		p.threadPool <- true
		go func(proxy string) {
			alive, timedelay := p.findUsable(proxy, testHost)
			if alive {
				proxyForHost.Mutex.Lock()
				proxyForHost.proxys = append(proxyForHost.proxys, proxy)
				proxyForHost.timedelay = append(proxyForHost.timedelay, timedelay)
				proxyForHost.Mutex.Unlock()
			}
			<-p.threadPool
		}(proxy)
	}
	for len(p.threadPool) > 0 {
		time.Sleep(0.2e9)
	}
	if proxyForHost.Len() > 0 {
		sort.Sort(proxyForHost)
		FKLog.G_Log.Informational(" *     [%v]测试与排序代理IP完成，可用：%v 个\n", key, proxyForHost.Len())
		return proxyForHost, true
	}
	FKLog.G_Log.Informational(" *     [%v]测试与排序代理IP完成，没有可用的代理IP\n", key)
	return proxyForHost, false
}

// 测试代理ip可用性
func (p *Proxy) findUsable(proxy string, testHost string) (alive bool, timedelay time.Duration) {
	t0 := time.Now()
	req := &FKRequest.Request{
		Url:         testHost,
		Method:      "HEAD",
		Header:      make(http.Header),
		DialTimeout: time.Second * time.Duration(DAIL_TIMEOUT),
		ConnTimeout: time.Second * time.Duration(CONN_TIMEOUT),
		TryTimes:    TRY_TIMES,
	}
	req.SetProxy(proxy)
	resp, err := p.surf.Download(req)

	if resp.StatusCode != http.StatusOK {
		return false, 0
	}

	return err == nil, time.Since(t0)
}
