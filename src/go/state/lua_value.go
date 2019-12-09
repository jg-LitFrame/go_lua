package state

import (
	"fmt"
	"go/luaapi"
	"go/number"
)

type luaValue interface{}

func typeOf(val luaValue) luaapi.LuaType {
	switch val.(type) {
	case nil:
		return luaapi.LUA_TNIL
	case bool:
		return luaapi.LUA_TBOOLEAN
	case int64:
		return luaapi.LUA_TNUMBER
	case float64:
		return luaapi.LUA_TNUMBER
	case string:
		return luaapi.LUA_TSTRING
	case *luaTable:
		return luaapi.LUA_TTABLE
	case *luaClosure:
		return luaapi.LUA_TFUNCTION
	default:
		panic("not impl val type!!!")
	}
}

func convertToFloat(val luaValue) (float64, bool) {
	switch x := val.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	case string:
		return number.ParseFloat(x)
	default:
		return 0, false
	}
}

func convertToInteger(val luaValue) (int64, bool) {
	switch x := val.(type) {
	case int64:
		return x, true
	case float64:
		return number.FloatToInteger(x)
	case string:
		return _stringToInteger(x)
	default:
		return 0, false
	}
}

func _stringToInteger(s string) (int64, bool) {
	if i, ok := number.ParseInteger(s); ok {
		return i, true
	}
	if f, ok := number.ParseFloat(s); ok {
		return number.FloatToInteger(f)
	}
	return 0, false
}

func LuaValToString(val luaValue) string {
	switch x := val.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprintf("%t", x)
	case int64:
		return fmt.Sprintf("%d", x)
	case float64:
		return fmt.Sprintf("%g", x)
	case string:
		return fmt.Sprintf("%s", x)
	case *luaTable:
		return x.ParserToString()
	case *luaClosure:
		return "luaClosure"
	default:
		panic("not impl val type!!!")
	}
}

func PrintLuaVal(val luaValue) {
	fmt.Print(LuaValToString(val))
}

func setMetatable(val luaValue, mt *luaTable, ls *luaState) {
	if t, ok := val.(*luaTable); ok {
		t.metaTable = mt
		return
	}
	key := fmt.Sprintf("_MT%d", typeOf(val))
	ls.registry.put(key, mt)
}

func getMetatable(val luaValue, ls *luaState) *luaTable {
	if t, ok := val.(*luaTable); ok {
		return t.metaTable
	}
	key := fmt.Sprintf("_MT%d", typeOf(val))
	if mt := ls.registry.get(key); mt != nil {
		return mt.(*luaTable)
	}
	return nil
}
