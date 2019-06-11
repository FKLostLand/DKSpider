package FKBase

import "net/url"

// 返回编码后的url.URL指针、及解析错误
func UrlEncode(urlStr string) (*url.URL, error) {
	urlObj, err := url.Parse(urlStr)
	urlObj.RawQuery = urlObj.Query().Encode()
	return urlObj, err
}
