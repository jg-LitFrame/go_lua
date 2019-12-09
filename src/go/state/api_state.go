package state

import (
	"fmt"
	"go/luaapi"
)

func (state *luaState) GetTop() int {
	return state.stack.top
}

func (state *luaState) AbsIndex(idx int) int {
	return state.stack.absIndex(idx)
}

func (state *luaState) CheckStack(n int) bool {
	state.stack.check(n)
	return true
}

func (state *luaState) Pop(n int) {
	for i := 0; i < n; i++ {
		state.stack.pop()
	}
}

func (state *luaState) Copy(from, to int) {
	val := state.stack.get(from)
	state.stack.set(to, val)
}

func (state *luaState) PushValue(idx int) {
	val := state.stack.get(idx)
	state.stack.push(val)
}

func (state *luaState) Replace(idx int) {
	val := state.stack.pop()
	state.stack.set(idx, val)
}

func (state *luaState) Insert(idx int) {
	state.Rotate(idx, 1)
}

func (state *luaState) Remove(idx int) {
	state.Rotate(idx, -1)
	state.Pop(1)
}

func (state *luaState) Rotate(idx, n int) {
	t := state.stack.top - 1
	p := state.stack.absIndex(idx) - 1
	var m int
	if n >= 0 {
		m = t - n
	} else {
		m = p - n - 1
	}
	state.stack.reverse(p, m)
	state.stack.reverse(m+1, t)
	state.stack.reverse(p, t)
}

func (state *luaState) SetTop(idx int) {
	newTop := state.stack.absIndex(idx)
	if newTop < 0 {
		panic("stack underflow")
	}
	n := state.stack.top - newTop
	if n > 0 {
		for i := 0; i < n; i++ {
			state.stack.pop()
		}
	} else if n < 0 {
		for i := 0; i > n; i-- {
			state.stack.push(nil)
		}
	}
}

func (state *luaState) PushNil() {
	state.stack.push(nil)
}

func (state *luaState) PushBoolean(b bool) {
	state.stack.push(b)
}

func (state *luaState) PushInteger(n int64) {
	state.stack.push(n)
}

func (state *luaState) PushNumber(n float64) {
	state.stack.push(n)
}

func (state *luaState) PushString(s string) {
	state.stack.push(s)
}

func (state *luaState) TypeName(tp luaapi.LuaType) string {
	switch tp {
	case luaapi.LUA_TNONE:
		return "no value"
	case luaapi.LUA_TNIL:
		return "nil"
	case luaapi.LUA_TBOOLEAN:
		return "boolean"
	case luaapi.LUA_TNUMBER:
		return "number"
	case luaapi.LUA_TSTRING:
		return "string"
	case luaapi.LUA_TTABLE:
		return "table"
	case luaapi.LUA_TFUNCTION:
		return "function"
	case luaapi.LUA_TTHREAD:
		return "thread"
	default:
		return "userdata"
	}
}

func (state *luaState) Type(idx int) luaapi.LuaType {
	if state.stack.isValid(idx) {
		val := state.stack.get(idx)
		return typeOf(val)
	}
	return luaapi.LUA_TNONE
}

func (state *luaState) IsNone(idx int) bool {
	return state.Type(idx) == luaapi.LUA_TNONE
}

func (state *luaState) IsNil(idx int) bool {
	return state.Type(idx) == luaapi.LUA_TNIL
}

func (state *luaState) IsNoneOrNil(idx int) bool {
	return state.Type(idx) == luaapi.LUA_TNIL
}

func (state *luaState) IsBoolean(idx int) bool {
	return state.Type(idx) == luaapi.LUA_TBOOLEAN
}

func (state *luaState) IsString(idx int) bool {
	t := state.Type(idx)
	return t == luaapi.LUA_TSTRING || t == luaapi.LUA_TNUMBER
}

func (state *luaState) IsNumber(idx int) bool {
	_, ok := state.ToNumberX(idx)
	return ok
}

func (state *luaState) IsInteger(idx int) bool {
	val := state.stack.get(idx)
	_, ok := val.(int64)
	return ok
}

func (state *luaState) IsThread(idx int) bool {
	return state.Type(idx) == luaapi.LUA_TTHREAD
}

func (state *luaState) IsTable(idx int) bool {
	return state.Type(idx) == luaapi.LUA_TTABLE
}

func (state *luaState) IsFunction(idx int) bool {
	return state.Type(idx) == luaapi.LUA_TFUNCTION
}

func (state *luaState) ToBoolean(idx int) bool {
	val := state.stack.get(idx)
	return convertToBoolean(val)
}

func convertToBoolean(val luaValue) bool {
	switch x := val.(type) {
	case nil:
		return false
	case bool:
		return x
	default:
		return true
	}
}
func (state *luaState) ToNumber(idx int) float64 {
	val, _ := state.ToNumberX(idx)
	return val
}

func (state *luaState) ToNumberX(idx int) (float64, bool) {
	val := state.stack.get(idx)
	return convertToFloat(val)
}

func (state *luaState) ToInteger(idx int) int64 {
	i, _ := state.ToIntegerX(idx)
	return i
}

func (state *luaState) ToIntegerX(idx int) (int64, bool) {
	val := state.stack.get(idx)
	return convertToInteger(val)
}

func (state *luaState) ToStringX(idx int) (string, bool) {
	val := state.stack.get(idx)
	switch x := val.(type) {
	case string:
		return x, true
	case int64, float64:
		s := fmt.Sprintf("%v", x)
		return s, true
	default:
		return "", true
	}
}

func (state *luaState) ToString(idx int) string {
	s, _ := state.ToStringX(idx)
	return s
}

func (state *luaState) PushGoFunction(f luaapi.GoFunction, n int) {
	closure := newGoClosure(f, n)
	for i := n; i > 0; i-- {
		val := state.stack.pop()
		closure.upvals[n-1] = &upvalue{&val}
	}
	state.stack.push(newGoClosure(f, 0))
}

func (state *luaState) IsGoFunction(idx int) bool {
	val := state.stack.get(idx)
	if c, ok := val.(*luaClosure); ok {
		return c.goFunc != nil
	}
	return false
}

func (state *luaState) ToGoFunction(idx int) luaapi.GoFunction {
	val := state.stack.get(idx)
	if c, ok := val.(*luaClosure); ok {
		return c.goFunc
	}
	return nil
}

func (state *luaState) PushGlobalTable() {
	global := state.registry.get(luaapi.LUA_RIDX_GLOBALS)
	state.stack.push(global)
}

func (state *luaState) GetGlobal(name string) luaapi.LuaType {
	t := state.registry.get(luaapi.LUA_RIDX_GLOBALS)
	return state.getTable(t, name, false)
}

func (state *luaState) SetGlobal(name string) {
	t := state.registry.get(luaapi.LUA_RIDX_GLOBALS)
	v := state.stack.pop()
	state.setTable(t, name, v, false)
}

func (state *luaState) Register(name string, f luaapi.GoFunction) {
	state.PushGoFunction(f, 0)
	state.SetGlobal(name)
}

func (state *luaState) GetMetatable(idx int) bool {
	val := state.stack.get(idx)

	if mt := getMetatable(val, state); mt != nil {
		state.stack.push(mt)
		return true
	} else {
		return false
	}
}

func (state *luaState) SetMetatable(idx int) {
	val := state.stack.get(idx)
	mtVal := state.stack.pop()

	if mtVal == nil {
		setMetatable(val, nil, state)
	} else if mt, ok := mtVal.(*luaTable); ok {
		setMetatable(val, mt, state)
	} else {
		panic("table expected!")
	}
}
