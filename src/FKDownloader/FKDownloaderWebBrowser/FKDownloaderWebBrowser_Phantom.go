package FKDownloaderWebBrowser

import (
	"FKBase"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type (
	// Phantom 基于PhantomJs的下载器实现，作为surfer的补充
	// 效率较surfer会慢很多，但是因为模拟浏览器，破防性更好
	// 支持UserAgent/TryTimes/RetryPause/自定义js
	Phantom struct {
		phantomJsFile string            // PhantomJs完整文件名
		tempJsDir     string            // 临时js存放目录
		jsFileMap     map[string]string // 已存在的js文件
		cookieJar     *cookiejar.Jar
	}

	// Response 用于解析PhantomJs的响应内容
	Response struct {
		Cookies []string
		Body    string
		Error   string
		Header  []struct {
			Name  string
			Value string
		}
	}

	//给PhantomJs传输cookie用
	Cookie struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Domain string `json:"domain"`
		Path   string `json:"path"`
	}
)

func CreatePhantomWebBrowser(phantomjsFile, tempJsDir string, jar ...*cookiejar.Jar) DownloaderWebBrowser {
	phantom := &Phantom{
		phantomJsFile: phantomjsFile,
		tempJsDir:     tempJsDir,
		jsFileMap:     make(map[string]string),
	}
	if len(jar) != 0 {
		phantom.cookieJar = jar[0]
	} else {
		phantom.cookieJar, _ = cookiejar.New(nil)
	}
	if !filepath.IsAbs(phantom.phantomJsFile) {
		phantom.phantomJsFile, _ = filepath.Abs(phantom.phantomJsFile)
	}
	if !filepath.IsAbs(phantom.tempJsDir) {
		phantom.tempJsDir, _ = filepath.Abs(phantom.tempJsDir)
	}
	// 创建/打开目录
	err := os.MkdirAll(phantom.tempJsDir, 0777)
	if err != nil {
		log.Printf("[E] Surfer: %v\n", err)
		return phantom
	}
	phantom.createJsFile("js", globalPhantomJS)
	return phantom
}

// 实现surfer下载器接口
func (p *Phantom) Download(req DownloadRequest) (resp *http.Response, err error) {
	var encoding = "utf-8"
	if _, params, err := mime.ParseMediaType(req.GetHeader().Get("Content-Type")); err == nil {
		if cs, ok := params["charset"]; ok {
			encoding = strings.ToLower(strings.TrimSpace(cs))
		}
	}

	req.GetHeader().Del("Content-Type")

	param, err := CreateDownloadParam(req)
	if err != nil {
		return nil, err
	}

	cookie := ""
	if req.GetEnableCookie() {
		httpCookies := p.cookieJar.Cookies(param.GetUrl())
		if len(httpCookies) > 0 {
			surferCookies := make([]*Cookie, len(httpCookies))

			for n, c := range httpCookies {
				surferCookie := &Cookie{Name: c.Name, Value: c.Value, Domain: param.GetUrl().Host, Path: "/"}
				surferCookies[n] = surferCookie
			}

			c, err := json.Marshal(surferCookies)
			if err != nil {
				log.Printf("cookie marshal error:%v", err)
			}
			cookie = string(c)
		}
	}

	resp = param.Writeback(resp)
	resp.Request.URL = param.GetUrl()

	var args = []string{
		p.jsFileMap["js"],
		req.GetUrl(),
		cookie,
		encoding,
		param.GetHeader().Get("User-Agent"),
		req.GetPostData(),
		strings.ToLower(param.GetMethod()),
		fmt.Sprint(int(req.GetDialTimeout() / time.Millisecond)),
	}
	if req.GetProxy() != "" {
		args = append([]string{"--proxy=" + req.GetProxy()}, args...)
	}

	for i := 0; i < param.GetTryTimes(); i++ {
		if i != 0 {
			time.Sleep(param.GetRetryPause())
		}

		cmd := exec.Command(p.phantomJsFile, args...)
		if resp.Body, err = cmd.StdoutPipe(); err != nil {
			continue
		}
		err = cmd.Start()
		if err != nil || resp.Body == nil {
			continue
		}
		var b []byte
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		retResp := Response{}
		err = json.Unmarshal(b, &retResp)
		if err != nil {
			continue
		}

		if retResp.Error != "" {
			log.Printf("phantomjs response error:%s", retResp.Error)
			continue
		}

		//设置header
		for _, h := range retResp.Header {
			resp.Header.Add(h.Name, h.Value)
		}

		//设置cookie
		for _, c := range retResp.Cookies {
			resp.Header.Add("Set-Cookie", c)
		}
		if req.GetEnableCookie() {
			if rc := resp.Cookies(); len(rc) > 0 {
				p.cookieJar.SetCookies(param.GetUrl(), rc)
			}
		}
		resp.Body = ioutil.NopCloser(strings.NewReader(retResp.Body))
		break
	}

	if err == nil {
		resp.StatusCode = http.StatusOK
		resp.Status = http.StatusText(http.StatusOK)
	} else {
		resp.StatusCode = http.StatusBadGateway
		resp.Status = err.Error()
	}
	return
}

//销毁js临时文件
func (p *Phantom) DestroyJsFiles() {
	JSDir, _ := filepath.Split(p.tempJsDir)
	if JSDir == "" {
		return
	}
	for _, filename := range p.jsFileMap {
		os.Remove(filename)
	}
	if len(FKBase.WalkDir(JSDir)) == 1 {
		os.Remove(JSDir)
	}
}

func (p *Phantom) createJsFile(fileName, jsCode string) {
	fullFileName := filepath.Join(p.tempJsDir, fileName)
	// 创建并写入文件
	f, _ := os.Create(fullFileName)
	f.Write([]byte(jsCode))
	f.Close()
	p.jsFileMap[fileName] = fullFileName
}
