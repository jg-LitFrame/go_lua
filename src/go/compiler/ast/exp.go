package ast

type Exp interface{}

type NilExp struct{ Line int }
type TrueExp struct{ Line int }
type False struct{ Line int }
type VarargExp struct{ Line int }
type IntegerExp struct{ Line int }
type FloatExp struct {
	Line int
	Val  int64
}
type String struct {
	Line int
	Val  string
}
type NameExp struct {
	Line int
	Name string
}

type UnopExp struct {
	Line int
	Op   int
	Exp  Exp
}

type BinopExp struct {
	Line int
	Op   int
	Exp1 Exp
	Exp2 Exp
}

type ConcatExp struct {
	Line int
	Exps []Exp
}

type TableConstructorExp struct {
	Line     int
	LastLine int
	KeyExps  []Exp
	ValExp   []Exp
}

type FuncDefExp struct {
	Line     int
	LastLine int
	ParList  []string
	IsVararg bool
	Block    *Block
}

type ParentsExp struct {
	Exp Exp
}

type TableAccessExp struct {
	LastLine  int
	PrefixExp Exp
	KeyExp    Exp
}

type FuncCallExp struct {
	Line      int
	LastLine  int
	PrefixExp Exp
	NameExp   Exp
	Args      []Exp
}
