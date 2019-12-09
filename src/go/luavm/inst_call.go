package luavm

import (
	"go/luaapi"
)

func closure(inst Instruction, vm luaapi.LuaVM) {
	a, bx := inst.ABx()
	a++
	vm.LoadProto(bx)
	vm.Replace(a)
}

func call(inst Instruction, vm luaapi.LuaVM) {
	a, b, c := inst.ABC()
	a++

	nArgs := _pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	_popResults(a, c, vm)
}

func _pushFuncAndArgs(a, b int, vm luaapi.LuaVM) int {
	if b >= 1 {
		vm.CheckStack(b)
		for i := a; i < a+b; i++ {
			vm.PushValue(i)
		}
		return b - 1
	} else {
		_fixStack(a, vm)
		top := vm.GetTop()
		count := vm.RegisterCount()
		return top - count - 1
	}
}

func _fixStack(a int, vm luaapi.LuaVM) {
	x := int(vm.ToInteger(-1))
	vm.Pop(1)

	vm.CheckStack(x - a)
	for i := a; i < x; i++ {
		vm.PushValue(i)
	}
	vm.Rotate(vm.RegisterCount()+1, x-a)
}

func _popResults(a, c int, vm luaapi.LuaVM) {
	if c == 1 {

	} else if c > 1 {
		for i := a + c - 2; i >= a; i-- {
			vm.Replace(i)
		}
	} else {
		vm.CheckStack(1)
		vm.PushInteger(int64(a))
	}
}

func _return(inst Instruction, vm luaapi.LuaVM) {
	a, b, _ := inst.ABC()
	a++

	if b == 1 {

	} else if b > 1 {
		vm.CheckStack(b - 1)
		for i := a; i <= a+b-2; i++ {
			vm.PushValue(i)
		}
	} else {
		_fixStack(a, vm)
	}
}

func vararg(inst Instruction, vm luaapi.LuaVM) {
	a, b, _ := inst.ABC()
	a++
	if b != 1 {
		vm.LoadVarargs(b - 1)
		_popResults(a, b, vm)
	}
}

func tailCall(inst Instruction, vm luaapi.LuaVM) {
	a, b, _ := inst.ABC()
	a++
	c := 0
	nArgs := _pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	_popResults(a, c, vm)
}

func self(inst Instruction, vm luaapi.LuaVM) {
	a, b, c := inst.ABC()
	a++
	b++
	vm.Copy(b, a+1)
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

func tForCall(inst Instruction, vm luaapi.LuaVM) {
	a, _, c := inst.ABC()
	a++
	_pushFuncAndArgs(a, 3, vm)
	vm.Call(2, c)
	_popResults(a+3, c+1, vm)
}

func tForLoop(inst Instruction, vm luaapi.LuaVM) {
	a, sBx := inst.AsBx()
	a++
	if !vm.IsNil(a + 1) {
		vm.Copy(a+1, a)
		vm.AddPC(sBx)
	}
}
