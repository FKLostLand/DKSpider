package FKSpider

import (
	"FKPinyin"
	"fmt"
)

// 全局蜘蛛种类实例
var GlobalSpiderSpecies = &SpiderSpecies{
	list: []*Spider{},
	hash: map[string]*Spider{},
}

// 蜘蛛种类列表
type SpiderSpecies struct {
	list   []*Spider
	hash   map[string]*Spider
	sorted bool
}

// 向蜘蛛种类清单添加新种类
func (ss *SpiderSpecies) Add(sp *Spider) *Spider {
	name := sp.Name
	for i := 2; true; i++ {
		if _, ok := ss.hash[name]; !ok {
			sp.Name = name
			ss.hash[sp.Name] = sp
			break
		}
		name = fmt.Sprintf("%s(%d)", sp.Name, i)
	}
	sp.Name = name
	ss.list = append(ss.list, sp)
	return sp
}

// 获取全部蜘蛛种类
func (ss *SpiderSpecies) Get() []*Spider {
	if !ss.sorted {
		l := len(ss.list)
		initials := make([]string, l)
		newlist := map[string]*Spider{}
		for i := 0; i < l; i++ {
			initials[i] = ss.list[i].GetName()
			newlist[initials[i]] = ss.list[i]
		}
		FKPinyin.SortInitials(initials)
		for i := 0; i < l; i++ {
			ss.list[i] = newlist[initials[i]]
		}
		ss.sorted = true
	}
	return ss.list
}

func (ss *SpiderSpecies) GetByName(name string) *Spider {
	return ss.hash[name]
}
