package FKBase

import (
	"math"
	"testing"
)

func TestBytes(t *testing.T) {
	t.Log(GlobalBytes.FormatUintBytesToString(11))
	t.Log(GlobalBytes.FormatUintBytesToString(1111))
	t.Log(GlobalBytes.FormatUintBytesToString(111122223))
	t.Log(GlobalBytes.FormatUintBytesToString(11112222333))
	t.Log(GlobalBytes.FormatUintBytesToString(1111222233334444))
	t.Log(GlobalBytes.FormatUintBytesToString(math.MaxInt64))

	v, _ := GlobalBytes.ParseStringToUintBytes("123B")
	t.Log(v)
	v, _ = GlobalBytes.ParseStringToUintBytes("123.45KB")
	t.Log(v)
	v, _ = GlobalBytes.ParseStringToUintBytes("12.34MB")
	t.Log(v)
	v, _ = GlobalBytes.ParseStringToUintBytes("1.23GB")
	t.Log(v)
	v, _ = GlobalBytes.ParseStringToUintBytes("123.45TB")
	t.Log(v)
}

func TestBytesParseErrors(t *testing.T) {
	_, err := GlobalBytes.ParseStringToUintBytes("B999")
	if err != nil {
		t.Log(err)
	}
}
