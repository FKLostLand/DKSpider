package FKHistory

import (
	"FKMySQL"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Success struct {
	tabName     string
	fileName    string
	new         map[string]bool // [FKRequest.Unique()]true
	old         map[string]bool // [FKRequest.Unique()]true
	inheritable bool
	sync.RWMutex
}

// 更新或加入成功记录，
// 对比是否已存在，不存在就记录，
// 返回值表示是否有插入操作。
func (s *Success) UpsertSuccess(reqUnique string) bool {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	if s.old[reqUnique] {
		return false
	}
	if s.new[reqUnique] {
		return false
	}
	s.new[reqUnique] = true
	return true
}

func (s *Success) HasSuccess(reqUnique string) bool {
	s.RWMutex.Lock()
	has := s.old[reqUnique] || s.new[reqUnique]
	s.RWMutex.Unlock()
	return has
}

// 删除成功记录
func (s *Success) DeleteSuccess(reqUnique string) {
	s.RWMutex.Lock()
	delete(s.new, reqUnique)
	s.RWMutex.Unlock()
}

func (s *Success) flush(provider string) (sLen int, err error) {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	sLen = len(s.new)
	if sLen == 0 {
		return
	}

	switch provider {
	case "mysql":
		_, err := FKMySQL.DB()
		if err != nil {
			return sLen, fmt.Errorf(" *     Fail  [添加成功记录][mysql]: %v 条 [ERROR]  %v\n", sLen, err)
		}
		table, ok := getWriteMysqlTable(s.tabName)
		if !ok {
			table = FKMySQL.CreateSQLTable()
			table.SetTableName(s.tabName).CustomPrimaryKey(`id VARCHAR(255) NOT NULL PRIMARY KEY`)
			err = table.Create()
			if err != nil {
				return sLen, fmt.Errorf(" *     Fail  [添加成功记录][mysql]: %v 条 [ERROR]  %v\n", sLen, err)
			}
			setWriteMysqlTable(s.tabName, table)
		}
		for key := range s.new {
			table.AutoInsert([]string{key})
			s.old[key] = true
		}
		err = table.FlushInsert()
		if err != nil {
			return sLen, fmt.Errorf(" *     Fail  [添加成功记录][mysql]: %v 条 [ERROR]  %v\n", sLen, err)
		}

	default:
		f, _ := os.OpenFile(s.fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)

		b, _ := json.Marshal(s.new)
		b[0] = ','
		f.Write(b[:len(b)-1])
		f.Close()

		for key := range s.new {
			s.old[key] = true
		}
	}
	s.new = make(map[string]bool)
	return
}
