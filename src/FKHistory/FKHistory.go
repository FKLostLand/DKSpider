package FKHistory

import (
	"FKLog"
	"FKMySQL"
	"FKRequest"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type (
	History struct {
		*Success
		*Failure
		provider string
		sync.RWMutex
	}
)

// 读取成功记录
func (h *History) ReadSuccess(provider string, inherit bool) {
	h.RWMutex.Lock()
	h.provider = provider
	h.RWMutex.Unlock()

	if !inherit {
		// 不继承历史记录时
		h.Success.old = make(map[string]bool)
		h.Success.new = make(map[string]bool)
		h.Success.inheritable = false
		return

	} else if h.Success.inheritable {
		// 本次与上次均继承历史记录时
		return

	} else {
		// 上次没有继承历史记录，但本次继承时
		h.Success.old = make(map[string]bool)
		h.Success.new = make(map[string]bool)
		h.Success.inheritable = true
	}

	switch provider {
	case "mysql":
		_, err := FKMySQL.DB()
		if err != nil {
			FKLog.G_Log.Error(" *     Fail  [读取成功记录][mysql]: %v\n", err)
			return
		}
		table, ok := getReadMysqlTable(h.Success.tabName)
		if !ok {
			table = FKMySQL.CreateSQLTable().SetTableName(h.Success.tabName)
			setReadMysqlTable(h.Success.tabName, table)
		}
		rows, err := table.SelectAll()
		if err != nil {
			return
		}

		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			h.Success.old[id] = true
		}

	default:
		f, err := os.Open(h.Success.fileName)
		if err != nil {
			return
		}
		defer f.Close()
		b, _ := ioutil.ReadAll(f)
		if len(b) == 0 {
			return
		}
		b[0] = '{'
		json.Unmarshal(append(b, '}'), &h.Success.old)
	}
	FKLog.G_Log.Informational(" *     [读取成功记录]: %v 条\n", len(h.Success.old))
}

// 取出失败记录
func (h *History) ReadFailure(provider string, inherit bool) {
	h.RWMutex.Lock()
	h.provider = provider
	h.RWMutex.Unlock()

	if !inherit {
		// 不继承历史记录时
		h.Failure.list = make(map[string]*FKRequest.Request)
		h.Failure.inheritable = false
		return

	} else if h.Failure.inheritable {
		// 本次与上次均继承历史记录时
		return

	} else {
		// 上次没有继承历史记录，但本次继承时
		h.Failure.list = make(map[string]*FKRequest.Request)
		h.Failure.inheritable = true
	}
	var fLen int
	switch provider {
	case "mysql":
		_, err := FKMySQL.DB()
		if err != nil {
			FKLog.G_Log.Error(" *     Fail  [取出失败记录][mysql]: %v\n", err)
			return
		}
		table, ok := getReadMysqlTable(h.Failure.tabName)
		if !ok {
			table = FKMySQL.CreateSQLTable().SetTableName(h.Failure.tabName)
			setReadMysqlTable(h.Failure.tabName, table)
		}
		rows, err := table.SelectAll()
		if err != nil {
			return
		}

		for rows.Next() {
			var key, failure string
			err = rows.Scan(&key, &failure)
			req, err := FKRequest.UnSerialize(failure)
			if err != nil {
				continue
			}
			h.Failure.list[key] = req
			fLen++
		}

	default:
		f, err := os.Open(h.Failure.fileName)
		if err != nil {
			return
		}
		b, _ := ioutil.ReadAll(f)
		f.Close()

		if len(b) == 0 {
			return
		}

		docs := map[string]string{}
		json.Unmarshal(b, &docs)

		fLen = len(docs)

		for key, s := range docs {
			req, err := FKRequest.UnSerialize(s)
			if err != nil {
				continue
			}
			h.Failure.list[key] = req
		}
	}

	FKLog.G_Log.Informational(" *     [取出失败记录]: %v 条\n", fLen)
}

// 清空缓存，但不输出
func (h *History) Empty() {
	h.RWMutex.Lock()
	h.Success.new = make(map[string]bool)
	h.Success.old = make(map[string]bool)
	h.Failure.list = make(map[string]*FKRequest.Request)
	h.RWMutex.Unlock()
}

// I/O输出成功记录，但不清缓存
func (h *History) FlushSuccess(provider string) {
	h.RWMutex.Lock()
	h.provider = provider
	h.RWMutex.Unlock()
	sucLen, err := h.Success.flush(provider)
	if sucLen <= 0 {
		return
	}
	if err != nil {
		FKLog.G_Log.Error("%v", err)
	} else {
		FKLog.G_Log.Informational(" *     [添加成功记录]: %v 条\n", sucLen)
	}
}

// I/O输出失败记录，但不清缓存
func (h *History) FlushFailure(provider string) {
	h.RWMutex.Lock()
	h.provider = provider
	h.RWMutex.Unlock()
	failLen, err := h.Failure.flush(provider)
	if failLen <= 0 {
		return
	}
	if err != nil {
		FKLog.G_Log.Error("%v", err)
	} else {
		FKLog.G_Log.Informational(" *     [添加失败记录]: %v 条\n", failLen)
	}
}

var (
	readMysqlTable     = map[string]*FKMySQL.FKSQLTable{}
	readMysqlTableLock sync.RWMutex
)

func getReadMysqlTable(name string) (*FKMySQL.FKSQLTable, bool) {
	readMysqlTableLock.RLock()
	tab, ok := readMysqlTable[name]
	readMysqlTableLock.RUnlock()
	if ok {
		return tab.Clone(), true
	}
	return nil, false
}

func setReadMysqlTable(name string, tab *FKMySQL.FKSQLTable) {
	readMysqlTableLock.Lock()
	readMysqlTable[name] = tab
	readMysqlTableLock.Unlock()
}

var (
	writeMysqlTable     = map[string]*FKMySQL.FKSQLTable{}
	writeMysqlTableLock sync.RWMutex
)

func getWriteMysqlTable(name string) (*FKMySQL.FKSQLTable, bool) {
	writeMysqlTableLock.RLock()
	tab, ok := writeMysqlTable[name]
	writeMysqlTableLock.RUnlock()
	if ok {
		return tab.Clone(), true
	}
	return nil, false
}

func setWriteMysqlTable(name string, tab *FKMySQL.FKSQLTable) {
	writeMysqlTableLock.Lock()
	writeMysqlTable[name] = tab
	writeMysqlTableLock.Unlock()
}
