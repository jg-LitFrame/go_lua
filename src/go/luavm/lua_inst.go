package luavm

import (
	"go/luaapi"
)

func move(i Instruction, vm luaapi.LuaVM) {
	a, b, _ := i.ABC()
	a++
	b++
	vm.Copy(b, a)
}

func jmp(i Instruction, vm luaapi.LuaVM) {
	a, sBx := i.AsBx()
	vm.AddPC(sBx)
	if a != 0 {
		vm.CloseUpvalues(a)
	}
}

func loadNil(i Instruction, vm luaapi.LuaVM) {
	a, b, _ := i.ABC()
	a++
	vm.PushNil()
	for i := a; i <= a+b; i++ {
		vm.Copy(-1, i)
	}
	vm.Pop(1)
}

func loadBool(i Instruction, vm luaapi.LuaVM) {
	a, b, c := i.ABC()
	a++
	vm.PushBoolean(b != 0)
	vm.Replace(a)
	if c != 0 {
		vm.AddPC(1)
	}
}

func loadK(i Instruction, vm luaapi.LuaVM) {
	a, bx := i.ABx()
	a++
	vm.GetConst(bx)
	vm.Replace(a)
}

func loadKx(i Instruction, vm luaapi.LuaVM) {
	a, _ := i.ABx()
	a++
	ax := Instruction(vm.Fetch()).Ax()
	vm.GetConst(ax)
	vm.Replace(a)
}

func _binaryArith(i Instruction, vm luaapi.LuaVM, op luaapi.ArithOp) {
	a, b, c := i.ABC()
	a++
	vm.GetRK(b)
	vm.GetRK(c)
	vm.Arith(op)
	vm.Replace(a)
}

func _unaryArith(i Instruction, vm luaapi.LuaVM, op luaapi.ArithOp) {
	a, b, _ := i.ABC()
	a++
	b++
	vm.PushValue(b)
	vm.Arith(op)
	vm.Replace(a)
}

func add(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPADD) }
func sub(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPSUB) }
func mul(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPMUL) }
func mod(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPMOD) }
func pow(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPPOW) }
func div(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPDIV) }
func idiv(i Instruction, vm luaapi.LuaVM) { _binaryArith(i, vm, luaapi.LUA_OPIDIV) }
func band(i Instruction, vm luaapi.LuaVM) { _binaryArith(i, vm, luaapi.LUA_OPBAND) }
func bor(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPBOR) }
func bxor(i Instruction, vm luaapi.LuaVM) { _binaryArith(i, vm, luaapi.LUA_OPBXOR) }
func shl(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPSHL) }
func shr(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPSHR) }
func unm(i Instruction, vm luaapi.LuaVM)  { _binaryArith(i, vm, luaapi.LUA_OPUNM) }
func bnot(i Instruction, vm luaapi.LuaVM) { _binaryArith(i, vm, luaapi.LUA_OPBNOT) }

func _len(i Instruction, vm luaapi.LuaVM) {
	a, b, _ := i.ABC()
	a++
	b++
	vm.Len(b)
	vm.Replace(a)
}

func concat(i Instruction, vm luaapi.LuaVM) {
	a, b, c := i.ABC()
	a++
	b++
	c++
	n := c - b + 1
	vm.CheckStack(n)
	for i := b; i <= c; i++ {
		vm.PushValue(i)
	}
	vm.Concat(n)
	vm.Replace(a)
}

func _compare(i Instruction, vm luaapi.LuaVM, op luaapi.CompareOp) {
	a, b, c := i.ABC()
	vm.GetRK(b)
	vm.GetRK(c)
	if vm.Compare(-2, -1, op) != (a != 0) {
		vm.AddPC(1)
	}
	vm.Pop(2)
}

func eq(i Instruction, vm luaapi.LuaVM) { _compare(i, vm, luaapi.LUA_OPEQ) }
func lt(i Instruction, vm luaapi.LuaVM) { _compare(i, vm, luaapi.LUA_OPLT) }
func le(i Instruction, vm luaapi.LuaVM) { _compare(i, vm, luaapi.LUA_OPLE) }

func not(i Instruction, vm luaapi.LuaVM) {
	a, b, _ := i.ABC()
	a++
	b++
	vm.PushBoolean(!vm.ToBoolean(b))
	vm.Replace(a)
}

func testSet(i Instruction, vm luaapi.LuaVM) {
	a, b, c := i.ABC()
	a++
	b++
	if vm.ToBoolean(b) == (c != 0) {
		vm.Copy(b, a)
	} else {
		vm.AddPC(1)
	}
}

func test(i Instruction, vm luaapi.LuaVM) {
	a, _, c := i.ABC()
	a++
	if vm.ToBoolean(a) != (c != 0) {
		vm.AddPC(1)
	}
}

///{{{ for 数值循环
func forPrep(i Instruction, vm luaapi.LuaVM) {
	a, sBx := i.AsBx()
	a++
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(luaapi.LUA_OPSUB)
	vm.Replace(a)
	vm.AddPC(sBx)
}

func forLoop(i Instruction, vm luaapi.LuaVM) {
	a, sBx := i.AsBx()
	a++
	vm.PushValue(a + 2)
	vm.PushValue(a)
	vm.Arith(luaapi.LUA_OPADD)
	vm.Replace(a)

	isPositiveStep := vm.ToNumber(a+2) >= 0
	if isPositiveStep && vm.Compare(a, a+1, luaapi.LUA_OPLE) ||
		!isPositiveStep && vm.Compare(a+1, a, luaapi.LUA_OPLE) {
		vm.AddPC(sBx)
		vm.Copy(a, a+3)
	}
}

//}}}

/*
func getTableUp(inst Instruction, vm luaapi.LuaVM) {
	a, _, c := inst.ABC()
	a++

	vm.PushGlobalTable()
	vm.GetRK(c)
	vm.GetTable(-2)
	vm.Replace(a)
	vm.Pop(1)
}
*/

func getUpVal(i Instruction, vm luaapi.LuaVM) {
	a, b, _ := i.ABC()
	a++
	b++
	vm.Copy(luaapi.LuaUpvaluesIndex(b), a)
}

func setUpVal(i Instruction, vm luaapi.LuaVM) {
	a, b, _ := i.ABC()
	a++
	b++
	vm.Copy(a, luaapi.LuaUpvaluesIndex(b))
}

func getTabUp(i Instruction, vm luaapi.LuaVM) {
	a, b, c := i.ABC()
	a++
	b++
	vm.GetRK(c)
	vm.GetTable(luaapi.LuaUpvaluesIndex(b))
	vm.Replace(a)
}

func setTabUp(i Instruction, vm luaapi.LuaVM) {
	a, b, c := i.ABC()
	a++
	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(luaapi.LuaUpvaluesIndex(a))
}
