package main

import (
	"fmt"
	"go/compiler/lexer"
	"go/compiler/parser"
	"io/ioutil"
	"encoding/json"
)

func main() {

	//testPrintLexer()
	testParse()
}

func testParse() {
	chunk, chunkName := "", ""
	astLua := parser.Parse(chunk, chunkName)
	b, err := json.Marshal(astLua)
	if err != nil {
		panic(err)
	}
	println(string(b))
}

func testPrintLexer() {
	luabytePath := "D:/work_space/go_lua/src/lua/ch14/test.lua"
	data, err := ioutil.ReadFile(luabytePath)
	if err != nil {
		panic(err)
	}
	testLexer(string(data), "chunk")
}

func testLexer(chunk, chunkName string) {
	l := lexer.NewLexer(chunk, chunkName)
	for {
		line, kind, token := l.NextToken()
		fmt.Printf("[%2d] [%-10s] %s\n",
			line, kindToCategory(kind), token)
		if kind == lexer.TOKEN_EOF {
			break
		}
	}

}

func kindToCategory(kind int) string {
	switch {
	case kind < lexer.TOKEN_SEP_SEMI:
		return "other"
	case kind <= lexer.TOKEN_SEP_RCURLY:
		return "separator"
	case kind <= lexer.TOKEN_OP_NOT:
		return "operator"
	case kind <= lexer.TOKEN_KW_WHILE:
		return "keyword"
	case kind == lexer.TOKEN_IDENTIFIER:
		return "identifier"
	case kind == lexer.TOKEN_NUMBER:
		return "number"
	case kind == lexer.TOKEN_STRING:
		return "string"
	default:
		return "other"
	}
}
