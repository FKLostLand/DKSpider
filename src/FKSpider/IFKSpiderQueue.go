package FKSpider

// 采集引擎中规则队列
type (
	SpiderQueue interface {
		Reset() //重置清空队列
		Add(*Spider)
		AddAll([]*Spider)
		AddKeywords(string) //为队列成员遍历添加Keywords属性，但前提必须是队列成员未被添加过Keywords
		GetByIndex(int) *Spider
		GetByName(string) *Spider
		GetAll() []*Spider
		Len() int // 返回队列长度
	}
)
