package FKSpider

import (
	"FKBase"
	"FKLog"
)

type (
	defaultSpiderQueue struct {
		list []*Spider
	}
)

func CreateSpiderQueue() SpiderQueue {
	return &defaultSpiderQueue{
		list: []*Spider{},
	}
}

func (q *defaultSpiderQueue) Reset() {
	q.list = []*Spider{}
}

func (q *defaultSpiderQueue) Add(sp *Spider) {
	sp.SetId(q.Len())
	q.list = append(q.list, sp)
}

func (q *defaultSpiderQueue) AddAll(list []*Spider) {
	for _, v := range list {
		q.Add(v)
	}
}

// 添加keywords，遍历蜘蛛队列得到新的队列（已被显式赋值过的spider将不再重新分配keywords）
func (q *defaultSpiderQueue) AddKeywords(keywords string) {
	keywordsSlice := FKBase.KeywordsParse(keywords)
	if len(keywordsSlice) == 0 {
		return
	}

	var unit1 []*Spider // 不可被添加自定义配置的蜘蛛
	var unit2 []*Spider // 可被添加自定义配置的蜘蛛
	for _, v := range q.GetAll() {
		if v.GetKeywords() == FKBase.KEYWORDS {
			unit2 = append(unit2, v)
			continue
		}
		unit1 = append(unit1, v)
	}

	if len(unit2) == 0 {
		FKLog.G_Log.Warning("本批任务无需填写自定义配置！\n")
		return
	}

	q.Reset()

	for _, keyword := range keywordsSlice {
		for _, v := range unit2 {
			v.Keywords = keyword
			nv := *v
			q.Add((&nv).Copy())
		}
	}
	if q.Len() == 0 {
		q.AddAll(append(unit1, unit2...))
	}

	q.AddAll(unit1)
}

func (q *defaultSpiderQueue) GetByIndex(idx int) *Spider {
	return q.list[idx]
}

func (q *defaultSpiderQueue) GetByName(n string) *Spider {
	for _, sp := range q.list {
		if sp.GetName() == n {
			return sp
		}
	}
	return nil
}

func (q *defaultSpiderQueue) GetAll() []*Spider {
	return q.list
}

func (q *defaultSpiderQueue) Len() int {
	return len(q.list)
}
