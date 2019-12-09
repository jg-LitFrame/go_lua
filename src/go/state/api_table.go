package state

import (
	"go/luaapi"
)

func (state *luaState) CreateTable(nArr, nRec int) {
	t := newLuaTable(nArr, nRec)
	state.stack.push(t)
}

func (state *luaState) NewTable() {
	state.CreateTable(0, 0)
}

func (state *luaState) GetTable(idx int) luaapi.LuaType {
	t := state.stack.get(idx)
	k := state.stack.pop()
	return state.getTable(t, k, false)
}

func (state *luaState) getTable(t, k luaValue, bRaw bool) luaapi.LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		if bRaw || v != nil || !tbl.hasMetafield("__index") {
			state.stack.push(v)
			return typeOf(v)
		}
	}
	if !bRaw {
		if mf := getMetafield(t, "__index", state); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				return state.getTable(x, k, false)
			case *luaClosure:
				state.stack.push(mf)
				state.stack.push(t)
				state.stack.push(k)
				state.Call(2, 1)
				v := state.stack.get(-1)
				return typeOf(v)
			}
		}
	}
	panic("not a table")
}

func (state *luaState) GetField(idx int, k string) luaapi.LuaType {
	t := state.stack.get(idx)
	return state.getTable(t, k, false)
}

func (state *luaState) GetI(idx int, i int64) luaapi.LuaType {
	t := state.stack.get(idx)
	return state.getTable(t, i, false)
}

func (state *luaState) SetTable(idx int) {
	t := state.stack.get(idx)
	v := state.stack.pop()
	k := state.stack.pop()
	state.setTable(t, k, v, false)
}

func (state *luaState) setTable(t, k, v luaValue, bRaw bool) {
	if tbl, ok := t.(*luaTable); ok {
		if bRaw || tbl.get(k) != nil || !tbl.hasMetafield("__newindex") {
			tbl.put(k, v)
			return
		}
	}
	if !bRaw {
		if mf := getMetafield(t, "__newindex", state); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				state.setTable(x, k, v, false)
				return
			case *luaClosure:
				state.stack.push(mf)
				state.stack.push(t)
				state.stack.push(k)
				state.stack.push(v)
				state.Call(3, 0)
				return
			}
		}
	}

	panic("not a table")
}

func (state *luaState) SetField(idx int, k string) {
	t := state.stack.get(idx)
	v := state.stack.pop()
	state.setTable(t, k, v, false)
}

func (state *luaState) SetI(idx int, i int64) {
	t := state.stack.get(idx)
	v := state.stack.pop()
	state.setTable(t, i, v, false)
}

func (tbl *luaTable) hasMetafield(fieldName string) bool {
	return tbl.metaTable != nil && tbl.metaTable.get(fieldName) != nil
}

func (state *luaState) RawGet(idx int) luaapi.LuaType {
	t := state.stack.get(idx)
	k := state.stack.pop()
	return state.getTable(t, k, true)
}

func (state *luaState) RawSet(idx int) {
	t := state.stack.get(idx)
	v := state.stack.pop()
	k := state.stack.pop()
	state.setTable(t, k, v, true)
}

func (state *luaState) RawGetI(idx int, i int64) luaapi.LuaType {
	t := state.stack.get(idx)
	return state.getTable(t, i, true)
}

func (state *luaState) RawSetI(idx int, i int64) {
	t := state.stack.get(idx)
	v := state.stack.pop()
	state.setTable(t, i, v, true)
}
