package binchunk

import (
	"encoding/binary"
	"math"
)

type chunkReader struct {
	data []byte
}

func (reader *chunkReader) readByte() byte {
	b := reader.data[0]
	reader.data = reader.data[1:]
	return b
}

func (reader *chunkReader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(reader.data)
	reader.data = reader.data[4:]
	return i
}

func (reader *chunkReader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(reader.data)
	reader.data = reader.data[8:]
	return i
}

func (reader *chunkReader) readLuaInteger() int64 {
	return int64(reader.readUint64())
}

func (reader *chunkReader) readLuaNumber() float64 {
	return math.Float64frombits(reader.readUint64())
}

func (reader *chunkReader) readBytes(n uint) []byte {
	bytes := reader.data[:n]
	reader.data = reader.data[n:]
	return bytes
}

func (reader *chunkReader) readString() string {
	size := uint(reader.readByte())
	if size == 0 {
		return ""
	}
	if size == 0xFF { //长字符串
		size = uint(reader.readUint64())
	}
	bytes := reader.readBytes(size - 1)
	return string(bytes)
}

func (reader *chunkReader) checkHeader() {
	if str := string(reader.readBytes(4)); str != LUA_SIGNATURE {
		panic("not a lua chunk file!!" + str)
	} else if reader.readByte() != LUA_VERSION {
		panic("Lua Version not match!!!")
	} else if reader.readByte() != LUA_FORMAT {
		panic("Lua Format not match!")
	} else if string(reader.readBytes(6)) != LUA_DATA {
		panic("LUA_DATA not match")
	} else if reader.readByte() != CINT_SIZE {
		panic("CINT_SIZE not match")
	} else if reader.readByte() != CSIZET_SIZE {
		panic("CSIZET_SIZE not match")
	} else if reader.readByte() != INSTRUCTION_SIZE {
		panic("INSTRUCTION_SIZE not match")
	} else if reader.readByte() != LUA_INTEGER_SIZE {
		panic("LUA_INTEGER_SIZE not match")
	} else if reader.readByte() != LUA_NUMBER_SIZE {
		panic("LUA_NUMBER_SIZE not match")
	} else if reader.readLuaInteger() != LUAC_INT {
		panic("LUAC_INT not match")
	} else if num := reader.readLuaNumber(); num != LUAC_NUM {
		println(num)
		panic("LUAC_NUM not match: ")
	}
	println("===== check header succ =======")
}

func (reader *chunkReader) readProto(parentSource string) *Prototype {
	source := reader.readString()
	if source == "" {
		source = parentSource
	}
	return &Prototype{
		Source:          source,
		LineDefine:      reader.readUint32(),
		LastLineDefined: reader.readUint32(),
		NumParams:       reader.readByte(),
		IsVarargs:       reader.readByte(),
		MaxStatckSize:   reader.readByte(),
		Code:            reader.readCode(),
		Constants:       reader.readConstants(),
		Upvalues:        reader.readUpvalues(),
		Protos:          reader.readProtos(source),
		LineInfo:        reader.readLineInfo(),
		LocVars:         reader.readLocVars(),
		UpvalueNames:    reader.readUpvalueNames(),
	}
}

func (reader *chunkReader) readUpvalueNames() []string {
	names := make([]string, reader.readUint32())
	for i := range names {
		names[i] = reader.readString()
	}
	return names
}

func (reader *chunkReader) readLocVars() []LocVar {
	vars := make([]LocVar, reader.readUint32())
	for i := range vars {
		vars[i] = LocVar{
			VarName: reader.readString(),
			StartPC: reader.readUint32(),
			EndPC:   reader.readUint32(),
		}
	}
	return vars
}

func (reader *chunkReader) readLineInfo() []uint32 {
	lineInfos := make([]uint32, reader.readUint32())
	for i := range lineInfos {
		lineInfos[i] = reader.readUint32()
	}
	return lineInfos
}

func (reader *chunkReader) readProtos(source string) []*Prototype {
	protos := make([]*Prototype, reader.readUint32())
	for i := range protos {
		protos[i] = reader.readProto(source)
	}
	return protos
}

func (reader *chunkReader) readUpvalues() []Upvalue {
	upvalues := make([]Upvalue, reader.readUint32())

	for i := range upvalues {
		upvalues[i].Instatck = reader.readByte()
		upvalues[i].Idx = reader.readByte()
	}
	return upvalues
}

func (reader *chunkReader) readCode() []uint32 {
	code := make([]uint32, reader.readUint32())
	for i := range code {
		code[i] = reader.readUint32()
	}
	return code
}

func (reader *chunkReader) readConstants() []interface{} {
	constants := make([]interface{}, reader.readUint32())
	for i := range constants {
		constants[i] = reader.readConstant()
	}
	return constants
}

func (reader *chunkReader) readConstant() interface{} {
	switch reader.readByte() {
	case TAG_NIL:
		return nil
	case TAG_BOOLEAN:
		return reader.readByte() != 0
	case TAG_NUMBER:
		return reader.readLuaNumber()
	case TAG_INTEGER:
		return reader.readLuaInteger()
	case TAG_SHORT_STR:
		return reader.readString()
	case TAG_LONG_STR:
		return reader.readString()
	default:
		panic("undefine lua val type!!!")
	}
}

//Undump 解析lua chunk文件
func Undump(data []byte) *Prototype {
	reader := &chunkReader{data}
	reader.checkHeader()

	reader.readByte()
	return reader.readProto("")
}
