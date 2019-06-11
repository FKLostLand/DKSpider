package FKBase

type IFKBytes interface {
	// 将字节大小转换为string
	// 例如：31323 字节转换为 30.59KB
	FormatUintBytesToString(b int64) string
	// 将字符串字节转换为数字字节
	// 例如：6GB 会被转换为 6442450944
	ParseStringToUintBytes(value string) (i int64, err error)
}

var (
	// 对外接口
	GlobalBytes = createBytes()
)
