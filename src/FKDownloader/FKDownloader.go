package FKDownloader

import (
	"FKDownloader/FKDownloaderWebBrowser"
	"FKRequest"
	"FKSpider"
	"errors"
	"net/http"
)

type MixDownloader struct {
	surf    FKDownloaderWebBrowser.DownloaderWebBrowser
	phantom FKDownloaderWebBrowser.DownloaderWebBrowser
}

func (md *MixDownloader) Download(sp *FKSpider.Spider, cReq *FKRequest.Request) *FKSpider.Context {
	ctx := FKSpider.GetContext(sp, cReq)

	var resp *http.Response
	var err error

	switch cReq.GetDownloaderID() {
	case FKRequest.SurfID:
		resp, err = md.surf.Download(cReq)

	case FKRequest.PhomtomJsID:
		resp, err = md.phantom.Download(cReq)
	}

	if resp.StatusCode >= 400 {
		err = errors.New("响应状态 " + resp.Status)
	}

	ctx.SetResponse(resp).SetError(err)

	return ctx
}
