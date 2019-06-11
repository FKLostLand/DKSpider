package FKDownloaderWebBrowser

import (
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func TestDownloader(t *testing.T) {
	var values = "username=123456@qq.com&password=123456&login_btn=login_btn&submit=login_btn"

	// 默认使用surf内核下载
	t.Log("********************************************* surf内核GET下载测试开始 *********************************************\n")
	resp, err := Download(&DefaultDownloadRequest{
		Url: "http://www.baidu.com/",
	})
	if err != nil {
		t.Error(err)
	}
	t.Logf("baidu resp.Status: %s\nresp.Header: %#v\n", resp.Status, resp.Header)

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	t.Logf("baidu resp.Body: %s\nerr: %v", b, err)

	// 默认使用surf内核下载
	t.Log("********************************************* surf内核POST下载测试开始 *********************************************\n")
	resp, err = Download(&DefaultDownloadRequest{
		Url:      "http://accounts.lewaos.com/",
		Method:   "POST",
		PostData: values,
	})
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("lewaos resp.Status: %s\nresp.Header: %#v\n", resp.Status, resp.Header)

	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	t.Logf("lewaos resp.Body: %s\nerr: %v", b, err)

	t.Log("********************************************* phantomjs内核GET下载测试开始 *********************************************\n")

	// 指定使用phantomjs内核下载
	resp, err = Download(&DefaultDownloadRequest{
		Url:          "http://www.baidu.com/",
		DownloaderID: 1,
	})
	if err != nil {
		log.Fatal(err)
	}

	t.Logf("baidu resp.Status: %s\nresp.Header: %#v\n", resp.Status, resp.Header)

	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	t.Logf("baidu resp.Body: %s\nerr: %v", b, err)

	t.Log("********************************************* phantomjs内核POST下载测试开始 *********************************************\n")

	// 指定使用phantomjs内核下载
	resp, err = Download(&DefaultDownloadRequest{
		DownloaderID: 1,
		Url:          "http://accounts.lewaos.com/",
		Method:       "POST",
		PostData:     values,
	})
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("lewaos resp.Status: %s\nresp.Header: %#v\n", resp.Status, resp.Header)

	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	t.Logf("lewaos resp.Body: %s\nerr: %v", b, err)

	DestroyPhantomJsFiles()
	time.Sleep(10e9)
}
