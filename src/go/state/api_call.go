package state

import (
	"go/binchunk"
	"go/luaapi"
	"go/luavm"
)

func (state *luaState) Load(chunk []byte, chunkName, mode string) int {
	proto := binchunk.Undump(chunk)
	c := newLuaClosure(proto)
	state.stack.push(c)

	if len(proto.Upvalues) > 0 {
		env := state.registry.get(luaapi.LUA_RIDX_GLOBALS)
		c.upvals[0] = &upvalue{val: &env}
	}
	return 0
}

func (state *luaState) Call(nArgs, nResults int) {
	val := state.stack.get(-(nArgs + 1))
	c, ok := val.(*luaClosure)
	if !ok {
		if mf := getMetafield(val, "__call", state); mf != nil {
			if c, ok = mf.(*luaClosure); ok {
				state.stack.push(val)
				state.Insert(-(nArgs + 2))
				nArgs++
			}
		}
	}
	if ok {
		//	fmt.Printf("Call %s<%d, %d>\n", c.proto.Source, c.proto.LineDefine, c.proto.LastLineDefined)
		if c.proto != nil {
			state.callLuaClosure(nArgs, nResults, c)
		} else {
			state.callGoClosure(nArgs, nResults, c)
		}
	} else {
		panic("not function!")
	}
}

func (state *luaState) callLuaClosure(nArgs, nResults int, c *luaClosure) {
	nRegs := int(c.proto.MaxStatckSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVarargs == 1

	newStack := newLuaStack(nRegs+20, state)
	newStack.closure = c

	funcAndArgs := state.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams)
	newStack.top = nRegs
	if nArgs > nParams && isVararg {
		newStack.varargs = funcAndArgs[nParams+1:]
	}

	state.pushLuaStack(newStack)
	state.runLuaClosure()
	state.popLuaStatck()

	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		state.stack.check(len(results))
		state.stack.pushN(results, nResults)
	}
}

func (state *luaState) runLuaClosure() {
	for {
		inst := luavm.Instruction(state.Fetch())
		inst.Execute(state)
		if inst.OpCode() == luavm.OP_RETURN {
			break
		}
	}
}

func (state *luaState) callGoClosure(nArgs, nResults int, c *luaClosure) {
	newStatck := newLuaStack(nArgs+20, state)
	newStatck.closure = c
	if nArgs > 0 {
		args := state.stack.popN(nArgs)
		newStatck.pushN(args, nArgs)
	}
	state.stack.pop()

	state.pushLuaStack(newStatck)
	r := c.goFunc(state)
	state.popLuaStatck()

	if nResults != 0 {
		results := newStatck.popN(r)
		state.stack.check(len(results))
		state.stack.pushN(results, nResults)
	}
}

func (state *luaState) PCall(nArgs, nResults, msgh int) (status int) {
	caller := state.stack
	status = luaapi.LUA_ERRRUN

	defer func() {
		if err := recover(); err != nil {
			if msgh != 0 {
				panic(err)
			}
			for state.stack != caller {
				state.popLuaStatck()
			}
			state.stack.push(err)
		}
	}()

	state.Call(nArgs, nResults)
	status = luaapi.LUA_OK
	return
}
