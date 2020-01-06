package codegen

import (
	"go/compiler/ast"
	"go/lexer"
	"go/luavm"
)

func cgExp(f *funcInfo, node ast.Exp, a, n int) {
	switch exp := node.(type) {
	case *ast.VarargExp:
		cgVarargExp(f, node, a, n)
	}
}

func cgVarargExp(f *funcInfo, node *ast.VarargExp, a, n int) {
	if !f.isVararg {
		panic("can not use '...' outside a vararg function")
	}
	f.emitVararg(a, n)
}

func cgFuncDefExp(f *funcInfo, node *ast.FuncDefExp, a int) {
	subFI := newFuncInfo(f, node)
	f.subFuncs = append(f.subFuncs, subFI)

	for _, param := range node.ParList {
		subFI.addLocalVar(param)
	}

	cgBlock(subFI, node.Block)
	subFI.exitScope()
	subFI.emitReturn(0, 0)

	bx := len(f.subFuncs) - 1
	f.emitClosure(a, bx)
}

func cgTableConstructorExp(f *funcInfo, node *ast.TableConstructorExp, a int) {
	nArr := 0
	for _, keyExp := range node.KeyExps {
		if keyExp == nil {
			nArr ++
		}
	}

	nExps := len(node.KeyExps)
	mulRet := nExps > 0 && isVarargOrFuncCall(node.ValExps[nExps - 1])

	f.emitNewTable(a, nArr, nExps - nArr)

	arrIdx := 0

	for i, keyExp := range node.KeyExps {
		valExp := node.ValExps[i]
		if keyExp == nil {
			arrIdx ++
			tmp := f.allocReg()
			if i == nExps - 1 && mulRet {
				cgExp(f, valExp, tmp, -1)
			}else{
				cgExp(f, valExp, tmp, 1)
			}
			if arrIdx % 50 == 0 || arrIdx == nArr {
				n := arrIdx % 50
				if n == 0 {
					n = 50
				}
				c := (arrIdx - 1) / 50 + 1
				f.freeRegs(n)
				if i == nExps - 1 && mulRet {
					f.emitSetList(a, 0, c)
				}else{
					f.emitSetList(a, n, c)
				}
			}
			continue
		}
		b := f.allocReg()
		cgExp(f, keyExp, b, 1)
		c := f.allocReg()
		cgExp(f, valExp, c, 1)
		f.freeRegs(2)
		f.emitSetTable(a, b, c)
	}
}

func cgUnopExp(f *funcInfo, node *ast.UnopExp, a int){ 
	b := f.allocReg()
	cgExp(f, node.Exp, b, 1)
	f.emitUnaryOp(node.Op, a, b)
	f.freeReg()
}

func cgConcatExp(f *funcInfo, node *ast.ConcatExp, a int) {
	for _, exp := range node.Exps {
		a := f.allocReg()
		cgExp(f, exp, a, 1)
	}

	c := f.usedRegs - 1
	b := c - len(node.Exps) + 1
	f.freeRegs(c - b + 1)
	f.emitABC(luavm.OP_CONCAT, a, b, c)
}

func cgBinopExp(f *funcInfo, node *ast.BinopExp, a int) {
	switch node.Op {
	case lexer.TOKEN_OP_AND, lexer.TOKEN_OP_OR:
		b := f.allocReg()
		cgExp(f, node.Exp1, b, 1)
		f.freeReg()
		if node.Op == lexer.TOKEN_OP_ADD {
			f.emitTestSet(a, b, 0)
		}else{
			f.emitTestSet(a, b, 1)
		}
		pcOfJmp := f.emitJmp(0, 0)

		b := f.allocReg()
		cgExp(f, node.Exp2, b, 1)
		f.freeReg()

		f.emitMove(a, b)
		f.fixSbx(pcOfJmp, f.pc() - pcOfJmp)
	default:
		b := f.allocReg()
		cgExp(f, node.Exp1, b, 1)
		c := f.allocReg()
		cgExp(f, node.Exp2, c, 1)
		f.emitBinaryOp(node.Op, a, b, c)
		f.freeRegs(2)
	}
}

func cgNameExp(f *funcInfo, node *ast.NameExp, a int) {
	if r := f.slotOfLocVar(node.Name); r >= 0 {
		f.emitMove(a, r)
	}else if idx := f.indexOfUpval(node.Name); idx >= 0 {
		f.emitGetUpval(a, idx)
	}else{
		taExp := &TableAccessExp {
			PrefixExp : &NameExp{0, "_Env"},
			KeyExp: &StringExp{0, node.Name},
		}
		cgTableAccessExp(f, taExp, a)
	}
}

func cgTableAccessExp(f *funcInfo, node *ast.TableAccessExp, a int) {
	b := f.allocReg()
	cgExp(f, node.PrefixExp, b, 1)
	c := f.allocReg()
	cgExp(f, node.KeyExp, c, 1)
	f.emitGetTable(a, b, c)
	f.freeRegs(2)
}

func cgFuncCallExp(f *funcInfo, node *ast.FuncCallExp, a, n int){
	nArgs := prepFuncCall(f, node, a)
	f.emitCall(a, nArgs, n)
}

func prepFuncCall(f *funcInfo, node *ast.FuncCallExp, a int) int {
	nArgs := len(node.Args)
	lastArgVarargOrFuncCall := false

	cgExp(f, node.PrefixExp, a, 1)

	if node.NameExp != nil {
		c := 0x100 + f.indexOfConstant(node.NameExp.Str)
		f.emitSelf(a, a, c)
	}

	for i, arg := range node.Args {
		tmp := f.allocReg()
		if i == nArgs - 1 && isVarargOrFuncCall(arg) {
			lastArgVarargOrFuncCall = true
			cgExp(f, arg, tmp, -1)
		}else{
			cgExp(f, arg, tmp, 1)
		}
	}
	f.freeRegs(nArgs)

	if node.NameExp != nil {
		nArgs++
	}
	if lastArgVarargOrFuncCall {
		nArgs = -1
	}
	return nArgs
}