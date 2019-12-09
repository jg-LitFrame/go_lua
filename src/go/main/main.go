package main

import (
	"fmt"
	"go/luaapi"
	"go/state"
	"io/ioutil"
)

func main() {
	// testStateOp()
	// testVM()
	//testTable()
	//testCall()
	//testPrint()
	//testMeta()
	//testTFor()
	testPCall()
}

func testPCall() {
	luabytePath := "D:/work_space/go_lua/src/lua/ch13/luac.out"
	data, err := ioutil.ReadFile(luabytePath)
	if err != nil {
		panic(err)
	}

	ls := state.New()
	ls.Register("pcall", pCall)
	ls.Register("print", print)
	ls.Register("error", error)
	ls.Load(data, "chunk", "b")
	ls.Call(0, 0)
}

func error(ls luaapi.LuaState) int {
	return ls.Error()
}

func pCall(ls luaapi.LuaState) int {
	nArgs := ls.GetTop() - 1
	status := ls.PCall(nArgs, -1, 0)
	ls.PushBoolean(status == luaapi.LUA_OK)
	ls.Insert(1)
	return ls.GetTop()
}

func testTFor() {
	luabytePath := "D:/work_space/go_lua/src/lua/ch12/luac.out"
	data, err := ioutil.ReadFile(luabytePath)
	if err != nil {
		panic(err)
	}

	ls := state.New()
	ls.Register("print", print)
	ls.Register("getmetatable", getMetatable)
	ls.Register("setmetatable", setMetatable)
	ls.Register("next", next)
	ls.Register("pairs", pairs)
	ls.Register("ipairs", iPairs)
	ls.Load(data, "chunk", "b")
	ls.Call(0, 0)
}

func setMetatable(ls luaapi.LuaState) int {
	ls.SetMetatable(1)
	return 1
}

func next(ls luaapi.LuaState) int {
	ls.SetTop(2) /* create a 2nd argument if there isn't one */
	if ls.Next(1) {
		return 2
	} else {
		ls.PushNil()
		return 1
	}
}

func pairs(ls luaapi.LuaState) int {
	ls.PushGoFunction(next, 0) /* will return generator, */
	ls.PushValue(1)            /* state, */
	ls.PushNil()
	return 3
}

func iPairs(ls luaapi.LuaState) int {
	ls.PushGoFunction(_iPairsAux, 0) /* iteration function */
	ls.PushValue(1)                  /* state */
	ls.PushInteger(0)                /* initial value */
	return 3
}

func _iPairsAux(ls luaapi.LuaState) int {
	i := ls.ToInteger(2) + 1
	ls.PushInteger(i)
	if ls.GetI(1, i) == luaapi.LUA_TNIL {
		return 1
	} else {
		return 2
	}
}

func testMeta() {
	luabytePath := "D:/work_space/go_lua/src/lua/ch11/luac.out"
	data, err := ioutil.ReadFile(luabytePath)
	if err != nil {
		panic(err)
	}

	ls := state.New()
	ls.Register("print", print)
	ls.Register("getmetatable", getMetatable)
	ls.Register("setmetatable", setMetatable)
	ls.Load(data, "chunk", "b")
	ls.Call(0, 0)
}

func getMetatable(ls luaapi.LuaState) int {
	if !ls.GetMetatable(1) {
		ls.PushNil()
	}
	return 1
}

func testPrint() {
	luabytePath := "D:/work_space/go_lua/src/lua/ch10/luac.out"
	data, err := ioutil.ReadFile(luabytePath)
	if err != nil {
		panic(err)
	}
	ls := state.New()
	ls.Register("print", print)
	ls.Load(data, "test01", "b")
	ls.Call(0, 0)
}

func print(ls luaapi.LuaState) int {
	nArgs := ls.GetTop()
	for i := 1; i <= nArgs; i++ {
		if ls.IsBoolean(i) {
			fmt.Printf("%t", ls.ToBoolean(i))
		} else if ls.IsString(i) {
			fmt.Printf(ls.ToString(i))
		} else {
			fmt.Printf(ls.TypeName(ls.Type(i)))
		}
		if i < nArgs {
			fmt.Print("\t")
		}
	}
	fmt.Println()
	return 0
}

func testCall() {
	luabytePath := "D:/work_space/go_lua/src/lua/ch08/luac.out"
	data, err := ioutil.ReadFile(luabytePath)
	if err != nil {
		panic(err)
	}
	ls := state.New()
	ls.Load(data, "test01", "b")
	ls.Call(0, 0)
}

/*
func testTable() {
	luabytePath := "D:/work_space/go_lua/src/lua/ch07/luac.out"
	data, err := ioutil.ReadFile(luabytePath)
	if err != nil {
		panic(err)
	}
	proto := binchunk.Undump(data)
	luaMain(proto)
}

func testVM() {
	luabytePath := "D:/work_space/go_lua/src/lua/ch06/luac.out"
	data, err := ioutil.ReadFile(luabytePath)
	if err != nil {
		panic(err)
	}
	proto := binchunk.Undump(data)
	luaMain(proto)
}

func luaMain(proto *binchunk.Prototype) {
	nRegs := int(proto.MaxStatckSize)
	ls := state.NewS(nRegs+8, proto)
	ls.SetTop(nRegs)
	for {
		pc := ls.PC()
		inst := luavm.Instruction(ls.Fetch())
		if inst.OpCode() != luavm.OP_RETURN {
			inst.Execute(ls)
			fmt.Printf("[%02d] %s", pc+1, inst.OpName())
			printStack(ls)
		} else {
			break
		}

	}
}
*/
func testStateOp() {
	ls := state.New()
	ls.PushInteger(1)
	ls.PushString("2.0")
	ls.PushString("3.0")
	ls.PushNumber(4.0)
	printStack(ls)

	ls.Arith(luaapi.LUA_OPADD)
	printStack(ls)
	ls.Arith(luaapi.LUA_OPBNOT)
	printStack(ls)
	ls.Len(2)
	printStack(ls)
	ls.Concat(3)
	printStack(ls)
	ls.PushBoolean(ls.Compare(1, 2, luaapi.LUA_OPEQ))
	printStack(ls)
}

func testStatePrint() {
	ls := state.New()
	ls.PushBoolean(true)
	printStack(ls)
	ls.PushInteger(10)
	printStack(ls)
	ls.PushNil()
	printStack(ls)
	ls.PushString("hello")
	printStack(ls)
	ls.PushValue(-4)
	printStack(ls)
	ls.Replace(3)
	printStack(ls)
	ls.SetTop(6)
	printStack(ls)
	ls.Remove(-3)
	printStack(ls)
	ls.SetTop(-5)
	printStack(ls)
}

func printStack(ls luaapi.LuaState) {
	top := ls.GetTop()
	for i := 1; i <= top; i++ {
		t := ls.Type(i)
		switch t {
		case luaapi.LUA_TBOOLEAN:
			fmt.Printf("[%t]", ls.ToBoolean(i))
		case luaapi.LUA_TNUMBER:
			fmt.Printf("[%g]", ls.ToNumber(i))
		case luaapi.LUA_TSTRING:
			fmt.Printf("[%q]", ls.ToString(i))
		default:
			fmt.Printf("[%s]", ls.TypeName(t))
		}
	}
	fmt.Println("")
}
