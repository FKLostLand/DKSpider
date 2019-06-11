package FKHistory

import (
	"FKMySQL"
	"FKRequest"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Failure struct {
	tabName     string
	fileName    string
	list        map[string]*FKRequest.Request // key使用请求url
	inheritable bool
	sync.RWMutex
}

func (f *Failure) PullFailureList() map[string]*FKRequest.Request {
	list := f.list
	f.list = make(map[string]*FKRequest.Request)
	return list
}

// 更新或加入失败记录，
// 对比是否已存在，不存在就记录，
// 返回值表示是否有插入操作。
func (f *Failure) UpsertFailure(req *FKRequest.Request) bool {
	f.RWMutex.Lock()
	defer f.RWMutex.Unlock()

	if f.list[req.Unique()] != nil {
		return false
	}
	f.list[req.Unique()] = req
	return true
}

// 删除失败记录
func (f *Failure) DeleteFailure(req *FKRequest.Request) {
	f.RWMutex.Lock()
	delete(f.list, req.Unique())
	f.RWMutex.Unlock()
}

// 先清空历史失败记录再更新
func (f *Failure) flush(provider string) (fLen int, err error) {
	f.RWMutex.Lock()
	defer f.RWMutex.Unlock()
	fLen = len(f.list)

	switch provider {
	case "mysql":
		_, err := FKMySQL.DB()
		if err != nil {
			return fLen, fmt.Errorf(" *     Fail  [添加失败记录][mysql]: %v 条 [PING]  %v\n", fLen, err)
		}
		table, ok := getWriteMysqlTable(f.tabName)
		if !ok {
			table = FKMySQL.CreateSQLTable()
			table.SetTableName(f.tabName).CustomPrimaryKey(`id VARCHAR(255) NOT NULL PRIMARY KEY`).AddColumn(`failure MEDIUMTEXT`)
			setWriteMysqlTable(f.tabName, table)
			// 创建失败记录表
			err = table.Create()
			if err != nil {
				return fLen, fmt.Errorf(" *     Fail  [添加失败记录][mysql]: %v 条 [CREATE]  %v\n", fLen, err)
			}
		} else {
			// 清空失败记录表
			err = table.Truncate()
			if err != nil {
				return fLen, fmt.Errorf(" *     Fail  [添加失败记录][mysql]: %v 条 [TRUNCATE]  %v\n", fLen, err)
			}
		}

		// 添加失败记录
		for key, req := range f.list {
			table.AutoInsert([]string{key, req.Serialize()})
			err = table.FlushInsert()
			if err != nil {
				fLen--
			}
		}

	default:
		// 删除失败记录文件
		os.Remove(f.fileName)
		if fLen == 0 {
			return
		}

		file, _ := os.OpenFile(f.fileName, os.O_CREATE|os.O_WRONLY, 0777)

		docs := make(map[string]string, len(f.list))
		for key, req := range f.list {
			docs[key] = req.Serialize()
		}
		b, _ := json.Marshal(docs)
		b = bytes.Replace(b, []byte(`\u0026`), []byte(`&`), -1)
		file.Write(b)
		file.Close()
	}
	return
}
