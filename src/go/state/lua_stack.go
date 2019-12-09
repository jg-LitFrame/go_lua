package state

import (
	"go/binchunk"
	"go/luaapi"
)

type luaStack struct {
	slots   []luaValue
	top     int
	pre     *luaStack
	closure *luaClosure
	varargs []luaValue
	pc      int
	state   *luaState
	openuvs map[int]*upvalue
}

type luaClosure struct {
	proto  *binchunk.Prototype
	goFunc luaapi.GoFunction
	upvals []*upvalue
}

type upvalue struct {
	val *luaValue
}

func newLuaClosure(proto *binchunk.Prototype) *luaClosure {
	c := &luaClosure{proto: proto}
	if nUpvals := len(proto.Upvalues); nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}

func newGoClosure(f luaapi.GoFunction, nUpvals int) *luaClosure {
	c := &luaClosure{goFunc: f}
	if nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}

func newLuaStack(size int, state *luaState) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size),
		top:   0,
		state: state,
	}
}

func (stack *luaStack) check(n int) {
	free := len(stack.slots) - stack.top
	if n >= free {
		return
	}
	for i := free; i < n; i++ {
		stack.slots = append(stack.slots, nil)
	}
}

func (stack *luaStack) push(val luaValue) {
	if stack.top == len(stack.slots) {
		panic("stack overflow")
	}
	stack.slots[stack.top] = val
	stack.top++
}

func (stack *luaStack) pop() luaValue {
	if stack.top < 1 {
		panic("stack underflow")
	}
	stack.top--
	val := stack.slots[stack.top]
	stack.slots[stack.top] = nil
	return val
}

func (stack *luaStack) absIndex(idx int) int {
	if idx <= luaapi.LUA_REGISTRYINDEX {
		return idx
	}
	if idx >= 0 {
		return idx
	}
	return idx + stack.top + 1
}

func (stack *luaStack) isValid(idx int) bool {
	if idx == luaapi.LUA_REGISTRYINDEX {
		return true
	}
	if idx < luaapi.LUA_REGISTRYINDEX {
		uvIdx := luaapi.LUA_REGISTRYINDEX - idx - 1
		c := stack.closure
		return c != nil && uvIdx < len(c.upvals)
	}
	absIdx := stack.absIndex(idx)
	return absIdx > 0 && absIdx <= stack.top
}

func (stack *luaStack) get(idx int) luaValue {
	if idx == luaapi.LUA_REGISTRYINDEX {
		return stack.state.registry
	}
	if idx < luaapi.LUA_REGISTRYINDEX {
		uvIdx := luaapi.LUA_REGISTRYINDEX - idx - 1
		c := stack.closure
		if c == nil || uvIdx >= len(c.upvals) {
			return nil
		}
		return *(c.upvals[uvIdx].val)
	}
	absIdx := stack.absIndex(idx)
	if absIdx > 0 && absIdx <= stack.top {
		return stack.slots[absIdx-1]
	}
	return nil
}

func (stack *luaStack) set(idx int, val luaValue) {
	if idx == luaapi.LUA_REGISTRYINDEX {
		stack.state.registry = val.(*luaTable)
		return
	}
	if idx < luaapi.LUA_REGISTRYINDEX {
		uvIdx := luaapi.LUA_REGISTRYINDEX - idx - 1
		c := stack.closure
		if c != nil && uvIdx < len(c.upvals) {
			*(c.upvals[uvIdx].val) = val
		}
		return
	}
	absIdx := stack.absIndex(idx)
	if absIdx > 0 && absIdx <= stack.top {
		stack.slots[absIdx-1] = val
		return
	}
	panic("invalid index!!")
}

func (stack *luaStack) reverse(from, to int) {
	slots := stack.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}

func (stack *luaStack) popN(n int) []luaValue {
	vals := make([]luaValue, n)
	for i := n - 1; i >= 0; i-- {
		vals[i] = stack.pop()
	}
	return vals
}

func (stack *luaStack) pushN(vals []luaValue, n int) {
	nVals := len(vals)
	if n < 0 {
		n = nVals
	}
	for i := 0; i < n; i++ {
		if i < nVals {
			stack.push(vals[i])
		} else {
			stack.push(nil)
		}
	}
}
