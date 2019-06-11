package FKBase

import (
	"encoding/json"
	"fmt"
	"github.com/axgle/mahonia"
	"hash/crc32"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
)

// Bytes2String直接转换底层指针，两者指向的相同的内存，改一个另外一个也会变。
// 效率是string([]byte{})的百倍以上，且转换量越大效率优势越明显。
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2Bytes直接转换底层指针，两者指向的相同的内存，改一个另外一个也会变。
// 效率是string([]byte{})的百倍以上，且转换量越大效率优势越明显。
// 转换之后若没做其他操作直接改变里面的字符，则程序会崩溃。
// 如 b:=String2bytes("xxx"); b[1]='d'; 程序将panic。
func String2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// 将文件名特殊字符替换为中文字符
func ReplaceSignToChineseSign(fileName string) string {
	var q = 1
	r := []rune(fileName)
	size := len(r)
	for i := 0; i < size; i++ {
		switch r[i] {
		case '"':
			if q%2 == 1 {
				r[i] = '“'
			} else {
				r[i] = '”'
			}
			q++
		case ':':
			r[i] = '：'
		case '*':
			r[i] = '×'
		case '<':
			r[i] = '＜'
		case '>':
			r[i] = '＞'
		case '?':
			r[i] = '？'
		case '/':
			r[i] = '／'
		case '|':
			r[i] = '∣'
		case '\\':
			r[i] = '╲'
		}
	}
	return strings.Replace(string(r), KEYWORDS, ``, -1)
}

// 字符串Hash
func String2Hash(s string) string {
	const IEEE = 0xedb88320
	var IEEETable = crc32.MakeTable(IEEE)
	hash := fmt.Sprintf("%x", crc32.Checksum([]byte(s), IEEETable))
	return hash
}

// 切分用户输入的自定义信息
func KeywordsParse(keywords string) []string {
	keywords = strings.TrimSpace(keywords)
	if keywords == "" {
		return []string{}
	}
	for _, v := range regexp.MustCompile(">[ \t\n\v\f\r]+<").FindAllString(keywords, -1) {
		keywords = strings.Replace(keywords, v, "><", -1)
	}
	m := map[string]bool{}
	for _, v := range strings.Split(keywords, "><") {
		v = strings.TrimPrefix(v, "<")
		v = strings.TrimSuffix(v, ">")
		if v == "" {
			continue
		}
		m[v] = true
	}
	s := make([]string, len(m))
	i := 0
	for k := range m {
		s[i] = k
		i++
	}
	return s
}

// 将对象转为json字符串
func JsonString(obj interface{}) string {
	b, _ := json.Marshal(obj)
	s := fmt.Sprintf("%+v", Bytes2String(b))
	r := strings.Replace(s, `\u003c`, "<", -1)
	r = strings.Replace(r, `\u003e`, ">", -1)
	return r
}

func Atoa(str interface{}) string {
	if str == nil {
		return ""
	}
	return strings.Trim(str.(string), " ")
}

func Atoi(str interface{}) int {
	if str == nil {
		return 0
	}
	i, _ := strconv.Atoi(strings.Trim(str.(string), " "))
	return i
}

func DecodeString(src, charset string) string {
	return mahonia.NewDecoder(charset).ConvertString(src)
}

func EncodeString(src, charset string) string {
	return mahonia.NewEncoder(charset).ConvertString(src)
}

func IsEmptyString(src string) bool{
	if len(src) <= 0{
		return true
	}
	// 去除空格
	src = strings.Replace(src, " ", "", -1)
	// 去除换行符
	src = strings.Replace(src, "\n", "", -1)
	if len(src) <= 0{
		return true
	}
	return false
}