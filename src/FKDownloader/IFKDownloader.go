package FKDownloader

import (
	"FKConfig"
	"FKDownloader/FKDownloaderWebBrowser"
	"FKRequest"
	"FKSpider"
	"net/http/cookiejar"
)

// The Downloader interface.
// You can implement the interface by implement function Download.
// Function Download need to return Page instance pointer that has request result downloaded from Request.
type Downloader interface {
	Download(*FKSpider.Spider, *FKRequest.Request) *FKSpider.Context
}

var (
	globalCookieJar, _  = cookiejar.New(nil)
	GlobalMixDownloader = &MixDownloader{
		surf:    FKDownloaderWebBrowser.CreateSurfWebBrowser(globalCookieJar),
		phantom: FKDownloaderWebBrowser.CreatePhantomWebBrowser(FKConfig.CONFIG_PHANTOM_JS_PATH, FKConfig.PHANTOM_JS_CACHE_PATH, globalCookieJar),
	}
)
