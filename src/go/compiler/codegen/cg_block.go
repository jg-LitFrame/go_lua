package codegen

import (
	"go/compiler/ast"
)

func cgBlock(f *funcInfo, node *ast.Block) {
	for _, stat := range node.Stat {
		cgStat(f, stat)
	}
	if node.RetExps != nil {
		cgRetStat(f, node.RetExps)
	}
}

func cgRetStat(f *funcInfo, exps []ast.Exp) {
	nExps := len(exps)
	if nExps == 0 {
		f.emitReturn(0, 0)
		return
	}
	multRet := isVarargOrFuncCall(exps[nExps - 1])
	for i, exp := range exps {
		r := f.allocReg()
		if i == nExps -1 && multRet {
			cgExp(f, exp, r, -1)
		}else{
			cgExp(f, exp, r, 1)
		}
	}
}

func isVarargOrFuncCall(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.VarargExp, *FuncCallExp:
		return true
	}
	return false
}

func cgStat(f *funcInfo, node ast.Stat) {
	switch stat := node.(type) {
	case *ast.FuncCallStat: cgFuncCallStat(f, stat)
	case *ast.BreakState: cgBreakStat(f, stat)
	case *ast.DoStat: cgDoStat(f, stat)
	case *ast.RepeatStat: cgRepeatStat(f, stat)
	case *ast.WhileStat: cgWhileStat(f, stat)
	case *ast.IfStat: cgIfStat(f, stat)
	case *ast.ForNumStat: cgForNumStat(f, stat)
	case *ast.ForInStat: cgForInStat(f, stat)
	case *ast.AssignStat: cgAssignStat(f, stat)
	case *ast.LocalVarDeclStat: cgLocalVarDeclStat(f, stat)
	case *ast.LocalFuncDefStat: cgLocalFuncDefStat(f, stat)
	case *ast.LabelStat, *ast.GotoStat: panic("not support!!")
	}
}

func cgLocalFuncDefStat(f *funcInfo, node *ast.LocalFuncDefStat){
	r := f.addLocalVar(node.Name)
	cgFuncDefExp(f, node.Exp, r)
}

func cgFuncCallStat(f *funcInfo, node *ast.FuncCallStat) {
	r := f.allocReg()
	cgFuncCallExp(f, node, r, 0)
	f.freeReg()
}

func cgBreakStat(f *funcInfo, node *ast.BreakStat) {
	pc := f.emitJmp(0, 0)
	f.addBreakJmp(pc)
}

func cgDoStat(f *funcInfo, node *ast.DoStat) {
	f.enterScope(false)
	cgBlock(f, node.Block)
	f.closeOpenUpvals()
	f.exitScope()
}

//While Stat
func cgWhileStat(f *funcInfo, node *ast.WhileStat) {
	pcBeforeExp := f.pc()

	r := f.allocReg()
	cgExp(f, node.Exp, r, 1)
	f.freeReg()

	f.emitTest(r, 0)
	pcJmpToEnd := f.emitJmp(0, 0)

	f.enterScope(true)
	cgBlock(f, node.Block)
	f.closeOpenUpvals()
	f.emitJmp(0, pcBeforeExp - f.pc() - 1)
	f.exitScope()

	f.fixSbx(pcJmpToEnd, f.pc() - pcJmpToEnd)
}


//RepeatStat
func cgRepeatStat(f *funcInfo, node *ast.RepeatStat) {
	f.enterScope(true)

	pcBeforeExp := f.pc()
	cgBlock(f, node.Block)

	r := f.allocReg()
	cgExp(f, node.Exp, r, 1)
	f.freeReg()

	f.emitTest(r, 0)
	f.emitJmp(f.getJmpArgA, pcBeforeExp - f.pc() - 1)
	f.closeOpenUpvals()

	f.exitScope()
}

//if Stat
func cgIfStat(f *funcInfo, node *ast.IfStat) {
	pcJmpToEnds := make([]int, len(node.exps))
	pcJmpToNextExp := -1

	for i, exp := range node.Exps {
		if pcJmpToNextExp >= 0 {
			f.fixSbx(pcJmpToNextExp, f.pc() - pcJmpToNextExp)
		}
		r := f.allocReg()
		cgExp(f, exp, r, 1)
		f.freeReg()

		f.emitTest(r, 0);
		pcJmpToNextExp = f.emitJmp(0, 0)

		f.enterScope(false)
		cgBlock(f, node.Blocks[i]);
		f.closeOpenUpvals()
		f.exitScope()

		if i < len(node.Exps) - 1 {
			pcJmpToEnds[i] = f.emitJmp(0, 0)
		}else{
			pcJmpToEnds[i] = pcJmpToNextExp
		}
	}

	for _, pc := range pcJmpToEnds {
		f.fixSbx(pc, f.pc() - pc)
	}
}

//for Num Stat
func cgForNumStat( f *funcInfo, node *ast.ForNumStat) {
	f.enterScope(true)

	cgLocalVarDeclStat(f, &ast.LocalVarDeclStat{
		NameList : []string{"(for index)", "(for limit)", "(for step)"},
		ExpList  : []ast.Exp{node.InitExp, node.LimitExp, node.StepExp},
	})

	f.addLocalVar(node.VarName)

	a := f.useRegs - 4
	pcForPrep := f.emitForPrep(a, 0)
	cgBlock(f, node.Block)
	f.closeOpenUpvals()
	pcForLoop := f.emitForLoop(a, 0)

	f.fixSbx(pcForPrep, pcForLoop - pcForPrep - 1)
	f.fixSbx(pcForLoop, pcForPrep - pcForLoop)

	f.exitScope()
}

//For comm Stat
func cgForInStat(f *funcInfo, node *ast.ForInStat) {
	f.enterScope(true)

	cgLocalVarDeclStat(f, &ast.LocalVarDeclStat{
		NameList : []string{"(for generator)", "(for state)", "(for control)"},
		ExpList  : node.ExpList,
	})

	for _,name := range node.NameList {
		f.addLocalVar(name)
	}

	pcJmpToTFC := f.emitJmp(0, 0)
	cgBlock(f, node.Block)
	f.closeOpenUpvals()
	f.fixSbx(pcJmpToTFC, f.pc() - pcJmpToTFC)

	rGenerator := f.slotOfLocVar("(for generator)")
	f.emitTForCall(rGenerator, len(node.NameList))
	f.emitTForLoop(rGenerator+2, pcJmpToTFC-f.pc()-1)

	f.exitScope()
}



func cgLocalVarDeclStat(f *funcInfo, node *ast.LocalVarDeclStat) {
	nExps := len(node.ExpList)
	nNames := len(node.NameList)

	oldRegs := f.useRegs
	if nExps == nNames {
		for _, exp := range node.ExpList {
			a := f.allocReg()
			cgExp(f, exp, a, 1)
		}
	}else if nExps > nNames {
		for i, exp := range node.ExpList {
			a := f.allocReg()
			if i == nExps - 1 && isVarargOrFuncCall(exp) {
				cgExp(f, exp, a, 0)
			}else{
				cgExp(f, exp, a, 1)
			}
		}
	}else {
		mulRet := false
		for i, exp := range node.ExpList {
			a := f.allocReg()
			if i == nExps - 1 && isVarargOrFuncCall(exp) {
				mulRet = true
				n := nNames - nExps + 1
				cgExp(f, exp, a, n)
				f.allocRegs(n - 1)
			}else{
				cgExp(f, exp, a, 1)
			}
			if !mulRet {
				n := nNames - nExps
				a := f.allocRegs(n)
				f.emitLoadNil(a, n)
			}
		}
	}
	f.useRegs = oldRegs
	for _, name := range node.NameList {
		f.addLocalVar(Name)
	}
}

func cgAssignStat(f *funcInfo, node *ast.AssignStat) {
	nExps := len(node.ExpList)
	nVars := len(node.VarList)

	oldRegs := f.useRegs
	tRegs := make([]int, nVars)
	kRegs := make([]int, nVars)
	vRegs := make([]int, nVars)

	for i, exp := range node.VarList {
		if taExp, ok := exp.(*ast.TableAccessExp); ok {
			tRegs[i] = f.allocReg()
			cgExp(f, taExp.PrefixExp, tRegs[i], 1)
			kRegs[i] = f.allocReg()
			cgExp(f, taExp.KeyExps, kRegs[i], 1)
		}
	}

	for i:=0; i < nVars; i++ {
		vRegs[i] = f.useRegs + i
	}

	if nExps >= nVars {
		for i, exp := range node.ExpList {
			a := f.allocReg()
			if i >= nVars && i == nRegs - 1 && isVarargOrFuncCall(exp) {
				cgExp(f, exp, a, 0)
			}else{
				cgExp(f, exp, a, 1)
			}
		}
	}else{
		multRet := false
		for i, exp := range exps {
			a := f.allocReg()
			if i == nExps - 1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1
				cgExp(f, exp, a, n)
				f.allocRegs(n - i)
			}else{
				cgExp(f, exp, a, 1)
			}
		}
		if !multRet {
			n := nVars - nExps
			a := f.allocRegs(n)
			f.emitLoadNil(a, n)
		}
	}
	for i, exp := range node.ExpList {
		if nameExp, ok := exp.(*ast.NameExp); ok {
			varName := nameExp.Name
			if a:= f.slotOfLocVar(varName); a > 0 {
				f.emitMove(a, vRegs[i])
			}else if b := f.indexOfUpval(varName); b > 0 {
				f.emitSetUpval(vRegs[i], b)
			}else{
				a := f.indexOfUpval("_Env")
				b := 0x100 + f.indexOfConstant(varName)
				f.emitSetTabUp(a, b, vRegs[i])
			}
		}else{
			f.emitSetTable(tRegs[i], kRegs[i], vRegs[i])
		}
	}
	f.useRegs = oldRegs
}