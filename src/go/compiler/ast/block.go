package ast

type Stat interface{}

type Block struct {
	LastLine int
	Stats    []Stat
	RetExps  []Exp
}

type EmptyStat struct{}
type BreakStat struct{ Line int }
type LabelStat struct{ Name string }
type GotoStat struct{ Name string }
type DoStat struct{ Block *Block }
type FuncCallStat = FuncCallExp

type WhileStat struct {
	Exp   Exp
	Block *Block
}

type RepeatStat struct {
	Block *Block
	Exp   Exp
}

type IfStat struct {
	Exps   []Exp
	Blocks []*Block
}

type ForNumStat struct {
	LineOfFor int
	LineOfDo  int
	VarName   string
	InitExp   Exp
	LimitExp  Exp
	StepExp    Exp
	Block     *Block
}

type ForInStat struct {
	LineOfDo int
	NameList []string
	ExpList  []Exp
	Block    *Block
}

type LocalVarDeclStat struct {
	LastLine int
	NameList []string
	ExpList  []Exp
}

type AssignStat struct {
	LastLine int
	VarList  []Exp
	ExpList  []Exp
}

type LocalFuncDefStat struct {
	Name string
	Exp  *FuncDefExp
}
