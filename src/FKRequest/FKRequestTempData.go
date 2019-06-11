package FKRequest

import (
	"FKBase"
	"FKLog"
	"encoding/json"
	"reflect"
)

type RequestTempData map[string]interface{}

// 返回临时缓存数据
func (rt RequestTempData) get(key string, defaultValue interface{}) interface{} {
	defer func() {
		if p := recover(); p != nil {
			FKLog.G_Log.Error(" *     Request.Temp.Get(%v): %v", key, p)
		}
	}()

	var (
		err error
		b   = FKBase.String2Bytes(rt[key].(string))
	)

	if reflect.TypeOf(defaultValue).Kind() == reflect.Ptr {
		err = json.Unmarshal(b, defaultValue)
	} else {
		err = json.Unmarshal(b, &defaultValue)
	}
	if err != nil {
		FKLog.G_Log.Error(" *     Request.Temp.Get(%v): %v", key, err)
	}
	return defaultValue
}

func (rt RequestTempData) set(key string, value interface{}) RequestTempData {
	b, err := json.Marshal(value)
	if err != nil {
		FKLog.G_Log.Error(" *     Request.Temp.Set(%v): %v", key, err)
	}
	rt[key] = FKBase.Bytes2String(b)
	return rt
}
