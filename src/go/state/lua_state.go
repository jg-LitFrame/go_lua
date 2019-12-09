package state

import (
	"go/luaapi"
)

type luaState struct {
	stack    *luaStack
	registry *luaTable
}

func New() *luaState {
	registry := newLuaTable(0, 0)
	registry.put(luaapi.LUA_RIDX_GLOBALS, newLuaTable(0, 0))

	ls := &luaState{
		registry: registry,
	}
	ls.pushLuaStack(newLuaStack(luaapi.LUA_MINSTACK, ls))
	return ls
}

func (state *luaState) pushLuaStack(stack *luaStack) {
	stack.pre = state.stack
	state.stack = stack
}

func (state *luaState) popLuaStatck() {
	statck := state.stack
	state.stack = statck.pre
	statck.pre = nil
}
