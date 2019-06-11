package FKTempDataPool

import (
	"testing"
)

func TestGetDataCell(t *testing.T) {

	t.Log(globalDataCellPool)

	m := make(map[string]interface{})
	m["name"] = "simon"
	m["age"] = 12
	d := GetDataCell("TestRule", m, "/test/1.html", "www.baidu.com", "2s")
	t.Log(d)

	t.Log(globalDataCellPool)

	PutDataCell(d)

	t.Log(globalDataCellPool)

	//t.Log(len(globalDataCellPool))
}
