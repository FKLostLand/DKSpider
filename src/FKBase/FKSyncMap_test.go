package FKBase

import "testing"

func TestSyncMap(t *testing.T) {
	testMap := CreateSyncMap()
	testMap.Store("1", 222)
	testMap.Store("2", "2")
	t.Log("testMap's length = ", testMap.Len())

	value1, ok := testMap.Load("1")
	if !ok {
		t.Error("can't find key 1")
	}
	t.Log(value1.(int))

	value2, ok := testMap.Load("2")
	if !ok {
		t.Error("can't find key 2")
	}
	t.Log(value2.(string))

	testMap.Delete("1")
	t.Log("testMap's length = ", testMap.Len())

	value1, ok = testMap.Load("1")
	if !ok {
		t.Log("Yes, can't find key 1")
	} else {
		t.Error("they find key 1!")
	}

	testMap.Clear()
	value2, ok = testMap.Load("2")
	if !ok {
		t.Log("Yes, can't find key 2")
	} else {
		t.Error("they find key 2!")
	}
}
