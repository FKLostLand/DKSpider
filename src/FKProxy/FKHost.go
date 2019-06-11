package FKProxy

import (
	"sync"
	"time"
)

type ProxyHost struct {
	curIndex  int // 当前代理ip索引
	proxys    []string
	timedelay []time.Duration
	isEcho    bool // 是否打印换ip信息
	sync.Mutex
}

// 实现排序接口
func (h *ProxyHost) Len() int {
	return len(h.proxys)
}

func (h *ProxyHost) Less(i, j int) bool {
	return h.timedelay[i] < h.timedelay[j]
}

func (h *ProxyHost) Swap(i, j int) {
	h.proxys[i], h.proxys[j] = h.proxys[j], h.proxys[i]
	h.timedelay[i], h.timedelay[j] = h.timedelay[j], h.timedelay[i]
}
