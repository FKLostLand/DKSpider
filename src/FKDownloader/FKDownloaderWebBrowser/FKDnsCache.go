package FKDownloaderWebBrowser

import "FKBase"

// 本地DNS缓存
var globalDnsCache = &DnsCache{ipPortLib: FKBase.CreateSyncMap()}

type DnsCache struct {
	ipPortLib FKBase.SyncMap
}

// 添加一个DNS缓存
func (d *DnsCache) Reg(addr, ipPort string) {
	d.ipPortLib.Store(addr, ipPort)
}

// 删除一个DNS缓存
func (d *DnsCache) Del(addr string) {
	d.ipPortLib.Delete(addr)
}

// 查询一个域名的DNS解析
func (d *DnsCache) Query(addr string) (string, bool) {
	ipPort, ok := d.ipPortLib.Load(addr)
	if !ok {
		return "", false
	}
	return ipPort.(string), true
}
