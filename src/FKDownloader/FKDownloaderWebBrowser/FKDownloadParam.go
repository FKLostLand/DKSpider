package FKDownloaderWebBrowser

import (
	"FKBase"
	"FKUserAgent"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DownloadParam struct {
	method        string
	url           *url.URL
	proxy         *url.URL
	body          io.Reader
	header        http.Header
	enableCookie  bool
	dialTimeout   time.Duration
	connTimeout   time.Duration
	tryTimes      int
	retryPause    time.Duration
	redirectTimes int
	client        *http.Client
}

func CreateDownloadParam(req DownloadRequest) (param *DownloadParam, err error) {
	param = new(DownloadParam)
	param.url, err = FKBase.UrlEncode(req.GetUrl())
	if err != nil {
		return nil, err
	}

	if req.GetProxy() != "" {
		if param.proxy, err = url.Parse(req.GetProxy()); err != nil {
			return nil, err
		}
	}

	param.header = req.GetHeader()
	if param.header == nil {
		param.header = make(http.Header)
	}

	switch method := strings.ToUpper(req.GetMethod()); method {
	case "GET", "HEAD":
		param.method = method
	case "POST":
		param.method = method
		param.header.Add("Content-Type", "application/x-www-form-urlencoded")
		param.body = strings.NewReader(req.GetPostData())
	case "POST-M":
		param.method = "POST"
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		values, _ := url.ParseQuery(req.GetPostData())
		for k, vs := range values {
			for _, v := range vs {
				writer.WriteField(k, v)
			}
		}
		err := writer.Close()
		if err != nil {
			return nil, err
		}
		param.header.Add("Content-Type", writer.FormDataContentType())
		param.body = body

	default:
		param.method = "GET"
	}

	param.enableCookie = req.GetEnableCookie()

	if len(param.header.Get("User-Agent")) == 0 {
		if param.enableCookie {
			param.header.Add("User-Agent", FKUserAgent.GlobalUserAgent.CreateStaticWebBrowserUA())
		} else {
			param.header.Add("User-Agent", FKUserAgent.GlobalUserAgent.CreateRandomWebBrowserUA())
		}
	}

	param.dialTimeout = req.GetDialTimeout()
	if param.dialTimeout < 0 {
		param.dialTimeout = 0
	}

	param.connTimeout = req.GetConnTimeout()
	param.tryTimes = req.GetTryTimes()
	param.retryPause = req.GetRetryPause()
	param.redirectTimes = req.GetRedirectTimes()
	return
}

func (p *DownloadParam) GetBody() io.Reader {
	return p.body
}

func (p *DownloadParam) IsEnableCookie() bool {
	return p.enableCookie
}

func (p *DownloadParam) GetClient() *http.Client {
	return p.client
}

func (p *DownloadParam) SetClient(t *http.Client) {
	p.client = t
}

func (p *DownloadParam) GetProxy() *url.URL {
	return p.proxy
}

func (p *DownloadParam) GetUrl() *url.URL {
	return p.url
}

func (p *DownloadParam) GetTryTimes() int {
	return p.tryTimes
}

func (p *DownloadParam) GetDialTimeout() time.Duration {
	return p.dialTimeout
}

func (p *DownloadParam) GetConnTimeout() time.Duration {
	return p.connTimeout
}

func (p *DownloadParam) GetRetryPause() time.Duration {
	return p.retryPause
}

func (p *DownloadParam) GetMethod() string {
	return p.method
}

func (p *DownloadParam) GetHeader() http.Header {
	return p.header
}

// 回写Request内容
func (p *DownloadParam) Writeback(resp *http.Response) *http.Response {
	if resp == nil {
		resp = new(http.Response)
		resp.Request = new(http.Request)
	} else if resp.Request == nil {
		resp.Request = new(http.Request)
	}

	if resp.Header == nil {
		resp.Header = make(http.Header)
	}

	resp.Request.Method = p.method
	resp.Request.Header = p.header
	resp.Request.Host = p.url.Host

	return resp
}

// checkRedirect is used as the value to http.Client.CheckRedirect
// when redirectTimes equal 0, redirect times is ∞
// when redirectTimes less than 0, not allow redirects
func (p *DownloadParam) CheckRedirect(req *http.Request, via []*http.Request) error {
	if p.redirectTimes == 0 {
		return nil
	}
	if len(via) >= p.redirectTimes {
		if p.redirectTimes < 0 {
			return fmt.Errorf("redirect is not allowed")
		}
		return fmt.Errorf("Stopped after %v  times redirect", p.redirectTimes)
	}
	return nil
}
