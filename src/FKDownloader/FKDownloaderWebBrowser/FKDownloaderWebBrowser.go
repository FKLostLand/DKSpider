package FKDownloaderWebBrowser

import (
	"net/http"
	"net/http/cookiejar"
	"sync"
)

var (
	surf                  DownloaderWebBrowser
	phantom               DownloaderWebBrowser
	syncOnceCreateSurf    sync.Once
	syncOnceCreatePhantom sync.Once
	tempJsDir             = "./tmp"
	phantomJsFile         = `./phantomjs`
	cookieJar, _          = cookiejar.New(nil)
)

func Download(req DownloadRequest) (resp *http.Response, err error) {
	switch req.GetDownloaderID() {
	case SurfID:
		syncOnceCreateSurf.Do(func() { surf = CreateSurfWebBrowser(cookieJar) })
		resp, err = surf.Download(req)
	case PhantomJsID:
		syncOnceCreatePhantom.Do(func() { phantom = CreatePhantomWebBrowser(phantomJsFile, tempJsDir, cookieJar) })
		resp, err = phantom.Download(req)
	}
	return
}

// 销毁Phantomjs的js临时文件
func DestroyPhantomJsFiles() {
	if pt, ok := phantom.(*Phantom); ok {
		pt.DestroyJsFiles()
	}
}
