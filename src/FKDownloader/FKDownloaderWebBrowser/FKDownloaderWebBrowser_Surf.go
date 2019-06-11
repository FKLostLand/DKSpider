package FKDownloaderWebBrowser

import (
	"FKUserAgent"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

type Surf struct {
	CookieJar *cookiejar.Jar
}

// 创建Surf下载器
func CreateSurfWebBrowser(jar ...*cookiejar.Jar) DownloaderWebBrowser {
	s := new(Surf)
	if len(jar) != 0 {
		s.CookieJar = jar[0]
	} else {
		s.CookieJar, _ = cookiejar.New(nil)
	}
	return s
}

// Download 实现surfer下载器接口
func (s *Surf) Download(req DownloadRequest) (resp *http.Response, err error) {
	param, err := CreateDownloadParam(req)
	if err != nil {
		return nil, err
	}
	param.GetHeader().Set("Connection", "close")
	param.SetClient(s.buildClient(param))
	resp, err = s.httpRequest(param)

	if err == nil {
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			var gzipReader *gzip.Reader
			gzipReader, err = gzip.NewReader(resp.Body)
			if err == nil {
				resp.Body = gzipReader
			}

		case "deflate":
			resp.Body = flate.NewReader(resp.Body)

		case "zlib":
			var readCloser io.ReadCloser
			readCloser, err = zlib.NewReader(resp.Body)
			if err == nil {
				resp.Body = readCloser
			}
		}
	}

	resp = param.Writeback(resp)

	return
}

// buildClient creates, configures, and returns a *http.Client type.
func (s *Surf) buildClient(param *DownloadParam) *http.Client {
	client := &http.Client{
		CheckRedirect: param.CheckRedirect,
	}

	if param.IsEnableCookie() {
		client.Jar = s.CookieJar
	}

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			var (
				c          net.Conn
				err        error
				ipPort, ok = globalDnsCache.Query(addr)
			)
			if !ok {
				ipPort = addr
				defer func() {
					if err == nil {
						globalDnsCache.Reg(addr, c.RemoteAddr().String())
					}
				}()
			} else {
				defer func() {
					if err != nil {
						globalDnsCache.Del(addr)
					}
				}()
			}
			c, err = net.DialTimeout(network, ipPort, param.GetDialTimeout())
			if err != nil {
				return nil, err
			}
			if param.GetConnTimeout() > 0 {
				c.SetDeadline(time.Now().Add(param.GetConnTimeout()))
			}
			return c, nil
		},
	}

	if param.GetProxy() != nil {
		transport.Proxy = http.ProxyURL(param.GetProxy())
	}

	if strings.ToLower(param.GetUrl().Scheme) == "https" {
		transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
		transport.DisableCompression = true
	}
	client.Transport = transport
	return client
}

// send uses the given *http.Request to make an HTTP request.
func (s *Surf) httpRequest(param *DownloadParam) (resp *http.Response, err error) {
	req, err := http.NewRequest(param.GetMethod(), param.GetUrl().String(), param.GetBody())
	if err != nil {
		return nil, err
	}

	req.Header = param.GetHeader()

	if param.GetTryTimes() <= 0 {
		for {
			resp, err = param.GetClient().Do(req)
			if err != nil {
				if !param.IsEnableCookie() {
					req.Header.Set("User-Agent", FKUserAgent.GlobalUserAgent.CreateRandomWebBrowserUA())
				}
				time.Sleep(param.GetRetryPause())
				continue
			}
			break
		}
	} else {
		for i := 0; i < param.GetTryTimes(); i++ {
			resp, err = param.GetClient().Do(req)
			if err != nil {
				if !param.IsEnableCookie() {
					req.Header.Set("User-Agent", FKUserAgent.GlobalUserAgent.CreateRandomWebBrowserUA())
				}
				time.Sleep(param.GetRetryPause())
				continue
			}
			break
		}
	}

	return resp, err
}
