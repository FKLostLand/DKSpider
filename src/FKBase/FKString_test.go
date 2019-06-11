package FKBase

import (
	"testing"
)

func TestBytes2String(t *testing.T) {
	var b []byte
	b = append(b, 228)
	b = append(b, 189)
	b = append(b, 160)
	b = append(b, 229)
	b = append(b, 165)
	b = append(b, 189)
	t.Log(Bytes2String(b))
}

func BenchmarkBytes2String(b *testing.B) {
	var bytes []byte
	bytes = append(bytes, 228)
	bytes = append(bytes, 189)
	bytes = append(bytes, 160)
	bytes = append(bytes, 229)
	bytes = append(bytes, 165)
	bytes = append(bytes, 189)

	var s string
	for i := 1; i <= b.N; i++ {
		s = string(bytes)
	}
	b.Log(s)
}

func BenchmarkBytes2String2(b *testing.B) {
	var bytes []byte
	bytes = append(bytes, 228)
	bytes = append(bytes, 189)
	bytes = append(bytes, 160)
	bytes = append(bytes, 229)
	bytes = append(bytes, 165)
	bytes = append(bytes, 189)

	var s string
	for i := 1; i <= b.N; i++ {
		s = Bytes2String(bytes)
	}
	b.Log(s)
}

func TestString2Bytes(t *testing.T) {
	b := String2Bytes("你好")
	t.Log(b)
}

func TestReplaceSignToChineseSign(t *testing.T) {
	b := "1111\"111\"<>? 2*3/|d\\:墨镜安排"
	t.Log(ReplaceSignToChineseSign(b))
}

func TestString2Hash(t *testing.T) {
	t.Log(String2Hash("HelloWorld"))
	t.Log(String2Hash("HelloWorld"))
	t.Log(String2Hash("Helloworld"))
}

func BenchmarkString2Hash(b *testing.B) {
	var s string
	for i := 1; i <= b.N; i++ {
		s = String2Hash("HelloWorld")
	}
	b.Log(s)
}

func TestJsonString(t *testing.T) {
	pUrl, err := UrlEncode("www.baidu.com")
	if err != nil {
		t.Error(err)
	}
	t.Log(JsonString(pUrl))
}
