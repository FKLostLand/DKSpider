package FKTempDataPool

import "sync"

/*
	这是临时存储对象，当GC时会被清除
*/
var (
	globalDataCellPool = &sync.Pool{
		New: func() interface{} {
			return DataCell{}
		},
	}
	globalFileCellPool = &sync.Pool{
		New: func() interface{} {
			return FileCell{}
		},
	}
)

type (
	// 数据存储单元
	DataCell map[string]interface{}
	// 文件存储单元
	FileCell map[string]interface{} // FileCell存储的完整文件名为： file/"Dir"/"RuleName"/"time"/"Name"
)

func GetDataCell(ruleName string, data map[string]interface{}, url string, parentUrl string, downloadTime string) DataCell {
	cell := globalDataCellPool.Get().(DataCell)
	cell["RuleName"] = ruleName   // 规定Data中的key
	cell["Data"] = data           // 数据存储,key须与Rule的Fields保持一致
	cell["Url"] = url             // 用于索引
	cell["ParentUrl"] = parentUrl // DataCell的上级url
	cell["DownloadTime"] = downloadTime
	return cell
}

func GetFileCell(ruleName, name string, bytes []byte) FileCell {
	cell := globalFileCellPool.Get().(FileCell)
	cell["RuleName"] = ruleName // 存储路径中的一部分
	cell["Name"] = name         // 规定文件名
	cell["Bytes"] = bytes       // 文件内容
	return cell
}

func PutDataCell(cell DataCell) {
	cell["RuleName"] = nil
	cell["Data"] = nil
	cell["Url"] = nil
	cell["ParentUrl"] = nil
	cell["DownloadTime"] = nil
	globalDataCellPool.Put(cell)
}

func PutFileCell(cell FileCell) {
	cell["RuleName"] = nil
	cell["Name"] = nil
	cell["Bytes"] = nil
	globalFileCellPool.Put(cell)
}
