package main

import (
	"fmt"
	"go/binchunk"
	"go/luavm"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Tmain() {
	chunkPath := "D:/work_space/go_lua/src/lua/ch01/luac.out"
	data, err := ioutil.ReadFile(chunkPath)
	if err != nil {
		panic(err)
	}
	proto := binchunk.Undump(data)
	list(proto)
}

func list(proto *binchunk.Prototype) {
	printHeader(proto)
	printCode(proto)
	printDetails(proto)
	for _, p := range proto.Protos {
		list(p)
	}
}

func printDetails(f *binchunk.Prototype) {
	fmt.Printf("constants (%d):\n", len(f.Constants))
	for i, k := range f.Constants {
		fmt.Printf("\t%d\t%s\n", i+1, constantToString(k))
	}

	fmt.Printf("locals (%d):\n", len(f.LocVars))
	for i, locVar := range f.LocVars {
		fmt.Printf("\t%d\t%s\t%d\t%d\n", i, locVar.VarName, locVar.StartPC+1,
			locVar.EndPC+1)
	}

	fmt.Printf("upvalues (%d)\n", len(f.Upvalues))
	for i, upval := range f.Upvalues {
		fmt.Printf("\t%d\t%s\t%d\t%d\n", i, upvalueName(f, i), upval.Instatck, upval.Idx)
	}
}

func upvalueName(f *binchunk.Prototype, idx int) string {
	if len(f.UpvalueNames) > 0 {
		return f.UpvalueNames[idx]
	}
	return "-"
}

func constantToString(k interface{}) string {
	switch k.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprintf("%t", k)
	case float64:
		return fmt.Sprintf("%g", k)
	case int64:
		return fmt.Sprintf("%d", k)
	case string:
		return fmt.Sprintf("%q", k)
	default:
		return "?"
	}
}

func printCode(f *binchunk.Prototype) {
	for pc, c := range f.Code {
		line := "-"
		if len(f.LineInfo) > 0 {
			line = fmt.Sprintf("%d", f.LineInfo[pc])
		}
		oc := luavm.Instruction(c)
		fmt.Printf("\t%d\t[%s]\t%s\t%08X\t", pc+1, line, oc.OpName(), c)
		printOperands(oc)
		fmt.Println("")
	}
}

func printOperands(i luavm.Instruction) {
	switch i.OpMode() {
	case luavm.IABC:
		a, b, c := i.ABC()
		//fmt.Printf("a=%d b=%d c=%d \t", a, b, c)
		fmt.Printf("%d", a)
		if i.BMode() != luavm.OpArgN {
			if b > 0xFF {
				fmt.Printf("  %d", -1-b&0xFF)
			} else {
				fmt.Printf("  %d", b)
			}
			if i.CMode() != luavm.OpArgN {
				if c > 0xFF {
					fmt.Printf("  %d", -1-c&0xFF)
				} else {
					fmt.Printf("  %d", c)
				}
			}
		}
	case luavm.IABx:
		a, bx := i.ABx()
		fmt.Printf("%d", a)
		if i.BMode() == luavm.OpArgK {
			fmt.Printf("  %d", -1-bx)
		} else if i.BMode() == luavm.OpArgU {
			fmt.Printf("  %d", bx)
		}
	case luavm.IAsBx:
		a, sbx := i.AsBx()
		fmt.Printf("%d  %d", a, sbx)
	case luavm.IAx:
		ax := i.Ax()
		fmt.Printf("%d", -1-ax)
	}
}

func printHeader(proto *binchunk.Prototype) {
	funcType := "main"
	if proto.LineDefine > 0 {
		funcType = "function"
	}
	varArgsFlag := ""
	if proto.IsVarargs > 0 {
		varArgsFlag = "+"
	}
	fmt.Printf("\n%s <%s:%d,%d> (%d instruction)\n", funcType, proto.Source,
		proto.LineDefine, proto.LastLineDefined, len(proto.Code))
	fmt.Printf("%d%s params, %d slots, %d upvalues, ", proto.NumParams,
		varArgsFlag, proto.MaxStatckSize, len(proto.Upvalues))
	fmt.Printf("%d locals, %d constants, %d functions\n", len(proto.LocVars),
		len(proto.Constants), len(proto.Protos))
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
