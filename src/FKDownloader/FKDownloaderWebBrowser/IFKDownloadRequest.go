package FKDownloaderWebBrowser

import (
	"net/http"
	"time"
)

type (
	DownloadRequest interface {
		// url
		GetUrl() string
		// GET POST POST-M HEAD
		GetMethod() string
		// POST values
		GetPostData() string
		// http header
		GetHeader() http.Header
		// enable http cookies
		GetEnableCookie() bool
		// dial tcp: i/o timeout
		GetDialTimeout() time.Duration
		// WSARecv tcp: i/o timeout
		GetConnTimeout() time.Duration
		// the max times of download
		GetTryTimes() int
		// the pause time of retry
		GetRetryPause() time.Duration
		// the download ProxyHost
		GetProxy() string
		// max redirect times
		GetRedirectTimes() int
		// select Surf ro PhomtomJS
		GetDownloaderID() int
	}
)
