package FKBase

import "testing"

func TestUrlEncode(t *testing.T) {
	pUrl, err := UrlEncode("www.baidu.com")
	if err != nil {
		t.Error(err)
	}

	t.Log(pUrl.Path)
	t.Log(pUrl.Host)
	t.Log(pUrl.Port())

	pUrl, err = UrlEncode("http://127.0.0.1:8080")
	if err != nil {
		t.Error(err)
	}

	t.Log(pUrl.Path)
	t.Log(pUrl.Host)
	t.Log(pUrl.Port())
}
