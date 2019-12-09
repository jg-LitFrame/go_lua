package state

func (vm *luaState) AddPC(n int) {
	vm.stack.pc += n
}

func (vm *luaState) PC() int {
	return vm.stack.pc
}

func (vm *luaState) Fetch() uint32 {
	i := vm.stack.closure.proto.Code[vm.stack.pc]
	vm.stack.pc++
	return i
}

func (vm *luaState) GetConst(idx int) {
	c := vm.stack.closure.proto.Constants[idx]
	vm.stack.push(c)
}

func (vm *luaState) GetRK(rk int) {
	if rk > 0xFF {
		vm.GetConst(rk & 0xFF)
	} else {
		vm.PushValue(rk + 1)
	}
}

func (vm *luaState) RegisterCount() int {
	return int(vm.stack.closure.proto.MaxStatckSize)
}

func (vm *luaState) LoadVarargs(n int) {
	if n < 0 {
		n = len(vm.stack.varargs)
	}
	vm.stack.check(n)
	vm.stack.pushN(vm.stack.varargs, n)
}

func (vm *luaState) LoadProto(idx int) {
	subProto := vm.stack.closure.proto.Protos[idx]
	closure := newLuaClosure(subProto)
	vm.stack.push(closure)
	stack := vm.stack

	for i, uvInfo := range subProto.Upvalues {
		uvIdx := int(uvInfo.Idx)
		if uvInfo.Instatck == 1 {
			if stack.openuvs == nil {
				stack.openuvs = map[int]*upvalue{}
			}
			if openuv, found := stack.openuvs[uvIdx]; found {
				closure.upvals[i] = openuv
			} else {
				closure.upvals[i] = &upvalue{val: &stack.slots[uvIdx]}
				stack.openuvs[uvIdx] = closure.upvals[i]
			}
		} else {
			closure.upvals[i] = stack.closure.upvals[uvIdx]
		}
	}
}

func (vm *luaState) CloseUpvalues(n int) {
	for i, openuv := range vm.stack.openuvs {
		if i >= n-1 {
			val := *openuv.val
			openuv.val = &val
			delete(vm.stack.openuvs, i)
		}
	}
}
