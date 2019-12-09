package state

import (
	"fmt"
	"go/number"
	"math"
)

type luaTable struct {
	metaTable *luaTable
	arr       []luaValue
	_map      map[luaValue]luaValue
	keys      map[luaValue]luaValue
}

func newLuaTable(nArr, nRec int) *luaTable {
	t := &luaTable{}
	if nArr > 0 { //TODO 估计有锅
		t.arr = make([]luaValue, 0, nArr)
	}
	if nRec > 0 {
		t._map = make(map[luaValue]luaValue, nRec)
	}
	return t
}

func (tbl *luaTable) get(key luaValue) luaValue {
	key = _floatToInteger(key)
	if idx, ok := key.(int64); ok {
		if idx >= 1 && idx <= int64(len(tbl.arr)) {
			return tbl.arr[idx-1]
		}
	}
	return tbl._map[key]
}

func _floatToInteger(key luaValue) luaValue {
	if f, ok := key.(float64); ok {
		if i, ok := number.FloatToInteger(f); ok {
			return i
		}
	}
	return key
}

func (tbl *luaTable) put(key, val luaValue) {
	if key == nil {
		panic("table index is nill")
	}
	if f, ok := key.(float64); ok && math.IsNaN(f) {
		panic("table index is NaN")
	}
	key = _floatToInteger(key)
	if idx, ok := key.(int64); ok && idx > 0 {

		arrLen := int64(len(tbl.arr))
		if idx <= arrLen {
			tbl.arr[idx-1] = val
			if idx == arrLen && val == nil {
				tbl._shrinkArray()
			}
			return
		}
		if idx == arrLen+1 {
			delete(tbl._map, key)
			if val != nil {
				tbl.arr = append(tbl.arr, val)
				tbl._expendArray()
			}
			return
		}
	}
	if val != nil {
		if tbl._map == nil {
			tbl._map = make(map[luaValue]luaValue, 8)
		}
		tbl._map[key] = val
	} else {
		delete(tbl._map, key)
	}
}

func (tbl *luaTable) _shrinkArray() {
	for i := len(tbl.arr) - 1; i >= 0; i-- {
		//TODO 应该是有bug
		if tbl.arr[i] == nil {
			tbl.arr = tbl.arr[0:i]
		}
	}
}

func (tbl *luaTable) _expendArray() {
	for idx := int64(len(tbl.arr)) + 1; true; idx++ {
		if val, found := tbl._map[idx]; found {
			delete(tbl._map, idx)
			tbl.arr = append(tbl.arr, val)
		} else {
			break
		}
	}
}

func (tbl *luaTable) len() int {
	return len(tbl.arr)
}

func (tbl *luaTable) ParserToString() string {
	s := "{"
	if tbl.arr != nil && len(tbl.arr) > 0 {
		for idx := 0; idx < len(tbl.arr); idx++ {
			s = fmt.Sprintf(s+" [%d]=%s,", idx, LuaValToString(tbl.arr[idx]))
		}
	}
	if tbl._map != nil {
		for key, val := range tbl._map {
			s = fmt.Sprintf(s+" [%s]=%s,", LuaValToString(key), LuaValToString(val))
		}
	}

	s += " }"
	return s
}

func (tbl *luaTable) nextKey(key luaValue) luaValue {
	if tbl.keys == nil || key == nil {
		tbl.initKeys()
	}
	return tbl.keys[key]
}

func (tbl *luaTable) initKeys() {
	tbl.keys = make(map[luaValue]luaValue)
	var key luaValue = nil
	for i, v := range tbl.arr {
		if v != nil {
			tbl.keys[key] = int64(i + 1)
			key = int64(i + 1)
		}
	}
	for k, v := range tbl._map {
		if v != nil {
			tbl.keys[key] = k
			key = k
		}
	}
}
