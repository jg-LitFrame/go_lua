package luavm

import (
	"go/luaapi"
)

const MAXARG_Bx = 1<<18 - 1
const MAXARG_sBx = MAXARG_Bx >> 1

type Instruction uint32

func (ins Instruction) OpCode() int {
	return int(ins & 0x3F)
}

func (ins Instruction) ABC() (a, b, c int) {
	a = int(ins >> 6 & 0xFF)
	c = int(ins >> 14 & 0x1FF)
	b = int(ins >> 23 & 0x1FF)
	return a, b, c
}

func (ins Instruction) ABx() (a, bx int) {
	a = int(ins >> 6 & 0xFF)
	bx = int(ins >> 14)
	return
}

func (ins Instruction) AsBx() (a, bx int) {
	a = int(ins >> 6 & 0xFF)
	bx = int(ins >> 14)
	return a, bx - MAXARG_sBx
}

func (ins Instruction) Ax() int {
	return int(ins >> 6)
}

func (ins Instruction) OpName() string {
	return opcodes[ins.OpCode()].name
}

func (ins Instruction) OpMode() byte {
	return opcodes[ins.OpCode()].opMode
}

func (ins Instruction) BMode() byte {
	return opcodes[ins.OpCode()].argBMode
}

func (ins Instruction) CMode() byte {
	return opcodes[ins.OpCode()].argCMode
}

func (ins Instruction) Execute(vm luaapi.LuaVM) {
	action := opcodes[ins.OpCode()].action
	if action != nil {
		action(ins, vm)
	} else {
		panic(ins.OpName())
	}
}
