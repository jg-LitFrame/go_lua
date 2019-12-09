package state

import (
	"go/luaapi"
)

func _eq(a, b luaValue, ls *luaState, bRaw bool) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case string:
		y, ok := b.(string)
		return ok && x == y
	case int64:
		switch y := b.(type) {
		case int64:
			return x == y
		case float64:
			return float64(x) == y
		default:
			return false
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x == y
		case int64:
			return x == float64(y)
		default:
			return false
		}
	case *luaTable:
		if !bRaw {
			if y, ok := b.(*luaTable); ok && x != y && ls != nil {
				if result, ok := callMetamethod(x, y, "__eq", ls); ok {
					return convertToBoolean(result)
				}
			}
		}
		return a == b
	default:
		return a == b
	}
}

func _lt(a, b luaValue, ls *luaState, bRaw bool) bool {
	switch x := a.(type) {
	case string:
		if y, ok := b.(string); ok {
			return x < y
		}
	case int64:
		switch y := b.(type) {
		case int64:
			return x < y
		case float64:
			return float64(x) < y
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x < y
		case int64:
			return x < float64(y)
		}
	}
	if !bRaw {
		if result, ok := callMetamethod(a, b, "__lt", ls); ok {
			return convertToBoolean(result)
		}
	}
	panic("comparsion error")
}

func _le(a, b luaValue, ls *luaState, bRaw bool) bool {
	switch x := a.(type) {
	case string:
		if y, ok := b.(string); ok {
			return x <= y
		}
	case int64:
		switch y := b.(type) {
		case int64:
			return x <= y
		case float64:
			return float64(x) <= y
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x <= y
		case int64:
			return x <= float64(y)
		}
	}
	if !bRaw {
		if result, ok := callMetamethod(a, b, "__le", ls); ok {
			return convertToBoolean(result)
		}
	}
	panic("comparsion error")
}

func (state *luaState) Compare(idx1, idx2 int, op luaapi.CompareOp) bool {
	a := state.stack.get(idx1)
	b := state.stack.get(idx2)
	switch op {
	case luaapi.LUA_OPEQ:
		return _eq(a, b, state, false)
	case luaapi.LUA_OPLT:
		return _lt(a, b, state, false)
	case luaapi.LUA_OPLE:
		return _le(a, b, state, false)
	case luaapi.LUA_RAW_OPEQ:
		return _eq(a, b, state, true)
	default:
		panic("invalid compare op!")
	}

}

func (state *luaState) RawEqual(idx1, idx2 int) bool {
	return state.Compare(idx1, idx2, luaapi.LUA_RAW_OPEQ)
}
