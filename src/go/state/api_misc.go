package state

func (state *luaState) Len(idx int) {
	val := state.stack.get(idx)
	if a, ok := val.(string); ok {
		state.stack.push(int64(len(a)))
	} else if result, ok := callMetamethod(val, val, "__len", state); ok {
		state.stack.push(result)
	} else if t, ok := val.(*luaTable); ok {
		state.stack.push(int64(t.len()))
	} else {
		panic("length error")
	}
}

func (state *luaState) Concat(n int) {
	if n == 0 {
		state.stack.push("")
	} else if n >= 2 {
		for i := 1; i < n; i++ {
			if state.IsString(-1) && state.IsString(-2) {
				s2 := state.ToString(-1)
				s1 := state.ToString(-2)
				state.stack.pop()
				state.stack.pop()
				state.stack.push(s1 + s2)
				continue
			}
			b := state.stack.pop()
			a := state.stack.pop()
			if result, ok := callMetamethod(a, b, "__concat", state); ok {
				state.stack.push(result)
				continue
			}
			panic("concatenation error")
		}
	}
}

func (state *luaState) Error() int {
	err := state.stack.pop()
	panic(err)
	//println(err)
	//return 0
}

func (state *luaState) RawLen(idx int) {
	val := state.stack.get(idx)
	if a, ok := val.(string); ok {
		state.stack.push(int64(len(a)))
	} else if t, ok := val.(*luaTable); ok {
		state.stack.push(int64(t.len()))
	} else {
		panic("length error")
	}
}

func (state *luaState) Next(idx int) bool {
	val := state.stack.get(idx)
	if t, ok := val.(*luaTable); ok {
		key := state.stack.pop()
		if nextKey := t.nextKey(key); nextKey != nil {
			state.stack.push(nextKey)
			state.stack.push(t.get(nextKey))
			return true
		}
		return false
	}
	panic("table expected")
}
