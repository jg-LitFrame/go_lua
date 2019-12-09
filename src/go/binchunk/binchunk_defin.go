package binchunk

const (
	LUA_SIGNATURE    = "\x1bLua"
	LUA_VERSION      = 0x53
	LUA_FORMAT       = 0
	LUA_DATA         = "\x19\x93\x0D\x0A\x1A\x0A"
	CINT_SIZE        = 4
	CSIZET_SIZE      = 4
	INSTRUCTION_SIZE = 4
	LUA_INTEGER_SIZE = 8
	LUA_NUMBER_SIZE  = 8
	LUAC_INT         = int64(0x5678)
	LUAC_NUM         = 370.5
)

const (
	TAG_NIL       = 0x00
	TAG_BOOLEAN   = 0x01
	TAG_NUMBER    = 0x03
	TAG_INTEGER   = 0x13
	TAG_SHORT_STR = 0x04
	TAG_LONG_STR  = 0x14
)

type header struct {
	signature       [4]byte
	version         byte
	format          byte
	luacData        [6]byte
	cintSize        byte
	sizetSize       byte
	instructionSize byte
	luaIntegerSize  byte
	luaNumberSize   byte
	luacInt         int64
	luacNum         float64
}

type Prototype struct {
	Source          string //for debug
	LineDefine      uint32
	LastLineDefined uint32
	NumParams       byte
	IsVarargs       byte
	MaxStatckSize   byte
	Code            []uint32
	Constants       []interface{}
	Upvalues        []Upvalue
	Protos          []*Prototype
	LineInfo        []uint32 //debug
	LocVars         []LocVar
	UpvalueNames    []string
}

type Upvalue struct {
	Instatck byte
	Idx      byte
}

type LocVar struct {
	VarName string
	StartPC uint32
	EndPC   uint32
}
