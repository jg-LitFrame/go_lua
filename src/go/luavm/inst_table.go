package luavm

import (
	"go/luaapi"
)

const LFIELDS_PER_PLUSH = 50

func newTable(ins Instruction, vm luaapi.LuaVM) {
	a, b, c := ins.ABC()
	vm.CreateTable(Fb2Int(b), Fb2Int(c))
	vm.Replace(a + 1)
}

func getTable(i Instruction, vm luaapi.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

func setTable(i Instruction, vm luaapi.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(a)
}

func setList(i Instruction, vm luaapi.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	bZero := b == 0
	if bZero {
		b = int(vm.ToInteger(-1)) - a - 1
		vm.Pop(1)
	}
	if c > 0 {
		c = c - 1
	} else {
		c = Instruction(vm.Fetch()).Ax()
	}
	idx := int64(c * LFIELDS_PER_PLUSH)
	for j := 1; j <= b; j++ {
		idx++
		vm.PushValue(a + j)
		vm.SetI(a, idx)
	}
	if bZero {
		for j := vm.RegisterCount() + 1; j <= vm.GetTop(); j++ {
			idx++
			vm.PushValue(j)
			vm.SetI(a, idx)
		}
		vm.SetTop(vm.RegisterCount())
	}
}
