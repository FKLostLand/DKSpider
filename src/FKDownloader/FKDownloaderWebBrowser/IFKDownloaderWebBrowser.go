package FKDownloaderWebBrowser

import (
	"net/http"
)

// Downloader represents an core of HTTP web browser for crawler.
type DownloaderWebBrowser interface {
	// GET @param url string, header http.Header, cookies []*http.Cookie
	// HEAD @param url string, header http.Header, cookies []*http.Cookie
	// POST PostForm @param url, referer string, values url.Values, header http.Header, cookies []*http.Cookie
	// POST-M PostMultipart @param url, referer string, values url.Values, header http.Header, cookies []*http.Cookie
	Download(DownloadRequest) (resp *http.Response, err error)
}
