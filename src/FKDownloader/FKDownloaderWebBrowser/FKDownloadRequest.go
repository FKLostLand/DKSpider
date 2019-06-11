package FKDownloaderWebBrowser

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type (
	// 默认实现的Request
	DefaultDownloadRequest struct {
		// url (必须填写)
		Url string
		// GET POST POST-M HEAD (默认为GET)
		Method string
		// http header
		Header http.Header
		// 是否使用cookies，在Spider的EnableCookie设置
		EnableCookie bool
		// POST values
		PostData string
		// dial tcp: i/o timeout
		DialTimeout time.Duration
		// WSARecv tcp: i/o timeout
		ConnTimeout time.Duration
		// the max times of download
		TryTimes int
		// how long pause when retry
		RetryPause time.Duration
		// max redirect times
		// when RedirectTimes equal 0, redirect times is ∞
		// when RedirectTimes less than 0, redirect times is 0
		RedirectTimes int
		// the download ProxyHost
		Proxy string

		// 指定下载器ID
		// 0为Surf高并发下载器，各种控制功能齐全
		// 1为PhantomJS下载器，特点破防力强，速度慢，低并发
		DownloaderID int

		// 保证prepare只调用一次
		syncOncePrepare sync.Once
	}
)

const (
	SurfID             = 0               // Surf下载器标识符
	PhantomJsID        = 1               // PhantomJs下载器标识符
	DefaultMethod      = "GET"           // 默认请求方法
	DefaultDialTimeout = 2 * time.Minute // 默认请求服务器超时
	DefaultConnTimeout = 2 * time.Minute // 默认下载超时
	DefaultTryTimes    = 3               // 默认最大下载次数
	DefaultRetryPause  = 2 * time.Second // 默认重新下载前停顿时长
)

// 发送请求之前的准备工作，修正检查一系列参数值
func (r *DefaultDownloadRequest) Prepare() {
	if r.Method == "" {
		r.Method = DefaultMethod
	}
	r.Method = strings.ToUpper(r.Method)

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

	if r.DownloaderID != PhantomJsID {
		r.DownloaderID = SurfID
	}
}

// url
func (r *DefaultDownloadRequest) GetUrl() string {
	r.syncOncePrepare.Do(r.Prepare)
	return r.Url
}

// GET POST POST-M HEAD
func (r *DefaultDownloadRequest) GetMethod() string {
	r.syncOncePrepare.Do(r.Prepare)
	return r.Method
}

// POST values
func (r *DefaultDownloadRequest) GetPostData() string {
	r.syncOncePrepare.Do(r.Prepare)
	return r.PostData
}

// http header
func (r *DefaultDownloadRequest) GetHeader() http.Header {
	r.syncOncePrepare.Do(r.Prepare)
	return r.Header
}

// enable http cookies
func (r *DefaultDownloadRequest) GetEnableCookie() bool {
	r.syncOncePrepare.Do(r.Prepare)
	return r.EnableCookie
}

// dial tcp: i/o timeout
func (r *DefaultDownloadRequest) GetDialTimeout() time.Duration {
	r.syncOncePrepare.Do(r.Prepare)
	return r.DialTimeout
}

// WSARecv tcp: i/o timeout
func (r *DefaultDownloadRequest) GetConnTimeout() time.Duration {
	r.syncOncePrepare.Do(r.Prepare)
	return r.ConnTimeout
}

// the max times of download
func (r *DefaultDownloadRequest) GetTryTimes() int {
	r.syncOncePrepare.Do(r.Prepare)
	return r.TryTimes
}

// the pause time of retry
func (r *DefaultDownloadRequest) GetRetryPause() time.Duration {
	r.syncOncePrepare.Do(r.Prepare)
	return r.RetryPause
}

// the download ProxyHost
func (r *DefaultDownloadRequest) GetProxy() string {
	r.syncOncePrepare.Do(r.Prepare)
	return r.Proxy
}

// max redirect times
func (r *DefaultDownloadRequest) GetRedirectTimes() int {
	r.syncOncePrepare.Do(r.Prepare)
	return r.RedirectTimes
}

// select Surf ro PhomtomJS
func (r *DefaultDownloadRequest) GetDownloaderID() int {
	r.syncOncePrepare.Do(r.Prepare)
	return r.DownloaderID
}
