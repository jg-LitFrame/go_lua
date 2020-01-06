package codegen

import (
	"go/luavm"
	"go/lexer"
)

var arithAndBitwiseBinops = map[int]int{
	lexer.TOKEN_OP_ADD:  luavm.OP_ADD,
	lexer.TOKEN_OP_SUB:  luavm.OP_SUB,
	lexer.TOKEN_OP_MUL:  luavm.OP_MUL,
	lexer.TOKEN_OP_MOD:  luavm.OP_MOD,
	lexer.TOKEN_OP_POW:  luavm.OP_POW,
	lexer.TOKEN_OP_DIV:  luavm.OP_DIV,
	lexer.TOKEN_OP_IDIV: luavm.OP_IDIV,
	lexer.TOKEN_OP_BAND: luavm.OP_BAND,
	lexer.TOKEN_OP_BOR:  luavm.OP_BOR,
	lexer.TOKEN_OP_BXOR: luavm.OP_BXOR,
	lexer.TOKEN_OP_SHL:  luavm.OP_SHL,
	lexer.TOKEN_OP_SHR:  luavm.OP_SHR,
}

type localVarInfo struct {
	pre      *localVarInfo
	name     string
	scopeLv  int
	slot     int
	captured bool
	endPC 	 int
}

type upvalInfo struct {
	locVarSlot int
	upvalIndex int
	index      int
}

type funcInfo struct {
	constants map[interface{}]int
	useRegs   int
	maxRegs   int
	scopeLv   int
	locVars   []*localVarInfo
	locNames  map[string]*localVarInfo
	breaks    [][]int
	parent    *funcInfo
	upvalues  map[string]upvalInfo
	insts     []uint32
	line      int
	lastLine  int
	numParams int
	isVararg  bool
	subFuncs  []*funcInfo
}

func newFuncInfo(parent *funcInfo, fd *FuncDefExp) *funcInfo {
	return &funcInfo{
		parent:    parent,
		subFuncs:  []*funcInfo{},
		locVars:   make([]*localVarInfo, 0, 8),
		locNames:  map[string]*localVarInfo{},
		upvalues:  map[string]upvalInfo{},
		constants: map[interface{}]int{},
		breaks:    make([][]int, 1),
		insts:     make([]uint32, 0, 8),
		numParams: len(fd.ParList),
		isVararg:  fd.IsVararg,
	}
}


func (f *funcInfo) indexOfConstant(k interface{}) int {
	if idx, ok := f.constants[k]; ok {
		return idx
	}
	idx := len(f.constants)
	f.constants[k] = idx
	return idx
}

func (f *funcInfo) allocReg() int {
	f.useRegs++
	if f.useRegs >= 255 {
		panic("function or expression need to much regs!")
	}
	if f.useRegs > f.maxRegs {
		f.maxRegs = f.useRegs
	}
	return f.useRegs - 1
}

func (f *funcInfo) freeReg() {
	f.useRegs--
}

func (f *funcInfo) allocRegs(n int) int {
	for i := 0; i < n; i++ {
		f.allocReg()
	}
	return f.useRegs - n
}

func (f *funcInfo) freeRegs(n int) {
	for i := 0; i < n; i++ {
		f.freeReg()
	}
}

func (f *funcInfo) enterScope(breakable bool) {
	f.scopeLv++
	if breakable {
		f.breaks = append(f.breaks, []int{})
	} else {
		f.breaks = append(f.breaks, nil)
	}
}

func (f *funcInfo) addLocalVar(name string) int {
	newVar := &localVarInfo{
		name:    name,
		pre:     f.locNames[name],
		scopeLv: f.scopeLv,
		slot:    f.allocReg(),
	}
	f.locVars = append(f.locVars, newVar)
	f.locNames[name] = newVar
	return newVar.slot
}

func (f *funcInfo) slotOfLocVar(name string) int {
	if locVar, found := f.locNames[name]; found {
		return locVar.slot
	}
	return -1
}

func (f *funcInfo) exitScope() {

	//TODO 没理解咋回事
	pendingBreakJmps := f.breaks[len(f.breaks)-1]
	f.breaks = f.breaks[:len(f.breaks)-1]

	a := f.getJmpArgA()
	for _, pc := range pendingBreakJmps {
		sBx := f.pc() - pc
		i := (sBx+luavm.MAXARG_sBx)<<14 | a<<6 | luavm.OP_JMP
		f.insts[pc] = uint32(i)
	}

	f.scopeLv--
	for _, locVar := range f.locNames {
		if locVar.scopeLv > f.scopeLv {
			f.removeLocVar(locVar)
		}
	}
}

func (f *funcInfo) removeLocVar(locVar *localVarInfo) {
	f.freeReg()
	if locVar.pre == nil {
		delete(f.locNames, locVar.name)
	} else if locVar.pre.scopeLv == locVar.scopeLv {
		f.removeLocVar(locVar.pre)
	} else {
		f.locNames[locVar.name] = locVar.pre
	}
}

func (f *funcInfo) addBreakJmp(pc int) {
	for i := f.scopeLv; i >= 0; i-- {
		if f.breaks[i] != nil {
			f.breaks[i] = append(f.breaks[i], pc)
		}
	}
	panic("<breaks> at line ? not inside a loop")
}

func (f *funcInfo) indexOfUpval(name string) int {
	if upval, ok := f.upvalues[name]; ok {
		return upval.index
	}

	if f.parent != nil {
		if locVar, found := f.parent.locNames[name]; found {
			idx := len(f.upvalues)
			f.upvalues[name] = upvalInfo{locVar.slot, -1, idx}
			locVar.captured = true
			return idx
		}
		if uvIdx := f.parent.indexOfUpval(name); uvIdx >= 0 {
			idx := len(f.upvalues)
			f.upvalues[name] = upvalInfo{-1, uvIdx, idx}
			return idx
		}
	}
	return -1
}

/* code */

func (f *funcInfo) pc() int {
	return len(f.insts) - 1
}

func (f *funcInfo) fixSbx(pc, sBx int) {
	i := f.insts[pc]
	i = i << 18 >> 18                  // clear sBx
	i = i | uint32(sBx+luavm.MAXARG_sBx)<<14 // reset sBx
	f.insts[pc] = i
}

// todo: rename?
func (f *funcInfo) fixEndPC(name string, delta int) {
	for i := len(f.locVars) - 1; i >= 0; i-- {
		locVar := f.locVars[i]
		if locVar.name == name {
			locVar.endPC += delta
			return
		}
	}
}

func (f *funcInfo) emitABC(opcode, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | opcode
	f.insts = append(f.insts, uint32(i))
}

func (f *funcInfo) emitABx(opcode, a, bx int) {
	i := bx<<14 | a<<6 | opcode
	f.insts = append(f.insts, uint32(i))
}

func (f *funcInfo) emitAsBx(opcode, a, b int) {
	i := (b+luavm.MAXARG_sBx)<<14 | a<<6 | opcode
	f.insts = append(f.insts, uint32(i))
}

func (f *funcInfo) emitAx(opcode, ax int) {
	i := ax<<6 | opcode
	f.insts = append(f.insts, uint32(i))
}

// r[a] = r[b]
func (f *funcInfo) emitMove( a, b int) {
	f.emitABC( luavm.OP_MOVE, a, b, 0)
}

// r[a], r[a+1], ..., r[a+b] = nil
func (f *funcInfo) emitLoadNil( a, n int) {
	f.emitABC( luavm.OP_LOADNIL, a, n-1, 0)
}

// r[a] = (bool)b; if (c) pc++
func (f *funcInfo) emitLoadBool( a, b, c int) {
	f.emitABC( luavm.OP_LOADBOOL, a, b, c)
}

// r[a] = kst[bx]
func (f *funcInfo) emitLoadK( a int, k interface{}) {
	idx := f.indexOfConstant(k)
	if idx < (1 << 18) {
		f.emitABx( luavm.OP_LOADK, a, idx)
	} else {
		f.emitABx( luavm.OP_LOADKX, a, 0)
		f.emitAx( luavm.OP_EXTRAARG, idx)
	}
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (f *funcInfo) emitVararg( a, n int) {
	f.emitABC( luavm.OP_VARARG, a, n+1, 0)
}

// r[a] = emitClosure(proto[bx])
func (f *funcInfo) emitClosure( a, bx int) {
	f.emitABx( luavm.OP_CLOSURE, a, bx)
}

// r[a] = {}
func (f *funcInfo) emitNewTable( a, nArr, nRec int) {
	f.emitABC( luavm.OP_NEWTABLE,
		a, luavm.Int2fb(nArr), luavm.Int2fb(nRec))
}

// r[a][(c-1)*FPF+i] := r[a+i], 1 <= i <= b
func (f *funcInfo) emitSetList( a, b, c int) {
	f.emitABC( luavm.OP_SETLIST, a, b, c)
}

// r[a] := r[b][rk(c)]
func (f *funcInfo) emitGetTable( a, b, c int) {
	f.emitABC( luavm.OP_GETTABLE, a, b, c)
}

// r[a][rk(b)] = rk(c)
func (f *funcInfo) emitSetTable( a, b, c int) {
	f.emitABC( luavm.OP_SETTABLE, a, b, c)
}

// r[a] = upval[b]
func (f *funcInfo) emitGetUpval( a, b int) {
	f.emitABC( luavm.OP_GETUPVAL, a, b, 0)
}

// upval[b] = r[a]
func (f *funcInfo) emitSetUpval( a, b int) {
	f.emitABC( luavm.OP_SETUPVAL, a, b, 0)
}

// r[a] = upval[b][rk(c)]
func (f *funcInfo) emitGetTabUp( a, b, c int) {
	f.emitABC( luavm.OP_GETTABUP, a, b, c)
}

// upval[a][rk(b)] = rk(c)
func (f *funcInfo) emitSetTabUp( a, b, c int) {
	f.emitABC( luavm.OP_SETTABUP, a, b, c)
}

// r[a], ..., r[a+c-2] = r[a](r[a+1], ..., r[a+b-1])
func (f *funcInfo) emitCall( a, nArgs, nRet int) {
	f.emitABC( luavm.OP_CALL, a, nArgs+1, nRet+1)
}

// return r[a](r[a+1], ... ,r[a+b-1])
func (f *funcInfo) emitTailCall( a, nArgs int) {
	f.emitABC( luavm.OP_TAILCALL, a, nArgs+1, 0)
}

// return r[a], ... ,r[a+b-2]
func (f *funcInfo) emitReturn( a, n int) {
	f.emitABC( luavm.OP_RETURN, a, n+1, 0)
}

// r[a+1] := r[b]; r[a] := r[b][rk(c)]
func (f *funcInfo) emitSelf( a, b, c int) {
	f.emitABC( luavm.OP_SELF, a, b, c)
}

// pc+=sBx; if (a) close all upvalues >= r[a - 1]
func (f *funcInfo) emitJmp( a, sBx int) int {
	f.emitAsBx( luavm.OP_JMP, a, sBx)
	return len(f.insts) - 1
}

// if not (r[a] <=> c) then pc++
func (f *funcInfo) emitTest( a, c int) {
	f.emitABC( luavm.OP_TEST, a, 0, c)
}

// if (r[b] <=> c) then r[a] := r[b] else pc++
func (f *funcInfo) emitTestSet( a, b, c int) {
	f.emitABC( luavm.OP_TESTSET, a, b, c)
}

func (f *funcInfo) emitForPrep( a, sBx int) int {
	f.emitAsBx( luavm.OP_FORPREP, a, sBx)
	return len(f.insts) - 1
}

func (f *funcInfo) emitForLoop( a, sBx int) int {
	f.emitAsBx( luavm.OP_FORLOOP, a, sBx)
	return len(f.insts) - 1
}

func (f *funcInfo) emitTForCall( a, c int) {
	f.emitABC( luavm.OP_TFORCALL, a, 0, c)
}

func (f *funcInfo) emitTForLoop( a, sBx int) {
	f.emitAsBx( luavm.OP_TFORLOOP, a, sBx)
}

// r[a] = op r[b]
func (f *funcInfo) emitUnaryOp( op, a, b int) {
	switch op {
	case lexer.TOKEN_OP_NOT:
		f.emitABC( luavm.OP_NOT, a, b, 0)
	case lexer.TOKEN_OP_BNOT:
		f.emitABC( luavm.OP_BNOT, a, b, 0)
	case lexer.TOKEN_OP_LEN:
		f.emitABC( luavm.OP_LEN, a, b, 0)
	case lexer.TOKEN_OP_UNM:
		f.emitABC( luavm.OP_UNM, a, b, 0)
	}
}

// r[a] = rk[b] op rk[c]
// arith & bitwise & relational
func (f *funcInfo) emitBinaryOp( op, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		f.emitABC( opcode, a, b, c)
	} else {
		switch op {
		case lexer.TOKEN_OP_EQ:
			f.emitABC( luavm.OP_EQ, 1, b, c)
		case lexer.TOKEN_OP_NE:
			f.emitABC( luavm.OP_EQ, 0, b, c)
		case lexer.TOKEN_OP_LT:
			f.emitABC( luavm.OP_LT, 1, b, c)
		case lexer.TOKEN_OP_GT:
			f.emitABC( luavm.OP_LT, 1, c, b)
		case lexer.TOKEN_OP_LE:
			f.emitABC( luavm.OP_LE, 1, b, c)
		case lexer.TOKEN_OP_GE:
			f.emitABC( luavm.OP_LE, 1, c, b)
		}
		f.emitJmp( 0, 1)
		f.emitLoadBool( a, 0, 1)
		f.emitLoadBool( a, 1, 0)
	}
}


func (f *funcInfo) getJmpArgA() int {
	hasCapturedLocVars := false
	minSlotOfLocVars := f.maxRegs
	for _, locVar := range f.locNames {
		if locVar.scopeLv == f.scopeLv {
			for v := locVar; v != nil && v.scopeLv == f.scopeLv; v = v.pre {
				if v.captured {
					hasCapturedLocVars = true
				}
				if v.slot < minSlotOfLocVars && v.name[0] != '(' {
					minSlotOfLocVars = v.slot
				}
			}
		}
	}
	if hasCapturedLocVars {
		return minSlotOfLocVars + 1
	} else {
		return 0
	}
}

func (f *funcInfo) closeOpenUpvals() {
	a := f.getJmpArgA()
	if a > 0 {
		f.emitJmp(a, 0)
	}
}