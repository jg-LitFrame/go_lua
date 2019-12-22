package parser

import (
	"go/compiler/ast"
	"go/compiler/lexer"
	"go/number"
	"math"
)

func Parse(chunk, chunkName string) *ast.Block{
	le := lexer.NewLexer(chunk, chunkName)
	block := parseBlock(le)
	le.NextTokenOfKind(lexer.TOKEN_EOF)
	return block
}

func parseBlock(le *lexer.Lexer) *ast.Block {
	return &ast.Block{
		Stats:    parseStats(le),
		RetExps:  parseRetExps(le),
		LastLine: le.Line(),
	}
}

func parseStats(lexer *lexer.Lexer) []ast.Stat {
	stats := make([]ast.Stat, 0, 8)
	for !_isReturnOrBlockEnd(lexer.LookAhead()) {
		stat := parseStat(lexer)
		if _, ok := stat.(*ast.EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}
func _isReturnOrBlockEnd(tokenKind int) bool {
	switch tokenKind {
	case lexer.TOKEN_KW_RETURN, lexer.TOKEN_EOF, lexer.TOKEN_KW_END, lexer.TOKEN_KW_ELSE, lexer.TOKEN_KW_ELSEIF, lexer.TOKEN_KW_UNTIL:
		return true
	}
	return false
}

func parseRetExps(le *lexer.Lexer) []ast.Exp {
	if le.LookAhead() != lexer.TOKEN_KW_RETURN {
		return nil
	}
	le.NextToken()
	switch le.LookAhead() {
	case lexer.TOKEN_EOF, lexer.TOKEN_KW_END, lexer.TOKEN_KW_ELSE, lexer.TOKEN_KW_ELSEIF, lexer.TOKEN_KW_UNTIL:
		return []ast.Exp{}
	case lexer.TOKEN_SEP_SEMI:
		return []ast.Exp{}
	default:
		exps := parseExpList(le)
		if le.LookAhead() == lexer.TOKEN_SEP_SEMI {
			le.NextToken()
		}
		return exps
	}
	return nil
}

func parseExpList(le *lexer.Lexer) []ast.Exp {
	exps := make([]ast.Exp, 0, 4)
	exps = append(exps, parseExp(le))
	for le.LookAhead() == lexer.TOKEN_SEP_COMMA {
		le.NextToken()
		exps = append(exps, parseExp(le))
	}
	return exps
}

func parseStat(le *lexer.Lexer) ast.Stat {
	switch le.LookAhead() {
	case lexer.TOKEN_SEP_SEMI:
		return parseEmptyStat(le)
	case lexer.TOKEN_KW_BREAK:
		return parseBreakStat(le)
	case lexer.TOKEN_SEP_LABEL:
		return parseLabelStat(le)
	case lexer.TOKEN_KW_GOTO:
		return parseGoToStat(le)
	case lexer.TOKEN_KW_DO:
		return parseDoStat(le)
	case lexer.TOKEN_KW_WHILE:
		return parseWhileStat(le)
	case lexer.TOKEN_KW_REPEAT:
		return parseRepeatStat(le)
	case lexer.TOKEN_KW_IF:
		return parseIfStat(le)
	case lexer.TOKEN_KW_FOR:
		return parseForStat(le)
	case lexer.TOKEN_KW_FUNCTION:
		return parseFuncDefStat(le)
	case lexer.TOKEN_KW_LOCAL:
		return parseLocalAssignOrFuncDefStat(le)
	default:
		return parseAssignOrCallStat(le)
	}
}

func parseEmptyStat(le *lexer.Lexer) *ast.EmptyStat {
	le.NextTokenOfKind(lexer.TOKEN_SEP_SEMI)
	return &ast.EmptyStat{}
}

func parseBreakStat(le *lexer.Lexer) *ast.BreakStat {
	le.NextTokenOfKind(lexer.TOKEN_KW_BREAK)
	return &ast.BreakStat{le.Line()}
}

func parseLabelStat(le *lexer.Lexer) *ast.LabelStat {
	le.NextTokenOfKind(lexer.TOKEN_SEP_LABEL)
	_, name := le.NextIdentifier()
	le.NextTokenOfKind(lexer.TOKEN_SEP_LABEL)
	return &ast.LabelStat{name}
}

func parseGoToStat(le *lexer.Lexer) *ast.GotoStat {
	le.NextTokenOfKind(lexer.TOKEN_KW_GOTO)
	_, name := le.NextIdentifier()
	return &ast.GotoStat{name}
}

func parseDoStat(le *lexer.Lexer) *ast.DoStat {
	le.NextTokenOfKind(lexer.TOKEN_KW_DO)
	block := parseBlock(le)
	le.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.DoStat{block}
}

func parseWhileStat(le *lexer.Lexer) *ast.WhileStat {
	le.NextTokenOfKind(lexer.TOKEN_KW_WHILE)
	exp := parseExp(le)
	le.NextTokenOfKind(lexer.TOKEN_KW_DO)
	block := parseBlock(le)
	le.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.WhileStat{exp, block}
}

func parseRepeatStat(le *lexer.Lexer) *ast.RepeatStat {
	le.NextTokenOfKind(lexer.TOKEN_KW_REPEAT)
	block := parseBlock(le)
	le.NextTokenOfKind(lexer.TOKEN_KW_UNTIL)
	exp := parseExp(le)
	return &ast.RepeatStat{block, exp}
}

func parseIfStat(le *lexer.Lexer) *ast.IfStat {
	exps := make([]ast.Exp, 0, 4)
	blocks := make([]*ast.Block, 0, 4)

	le.NextTokenOfKind(lexer.TOKEN_KW_IF)
	exps = append(exps, parseExp(le))
	le.NextTokenOfKind(lexer.TOKEN_KW_THEN)
	blocks = append(blocks, parseBlock(le))

	for le.LookAhead() == lexer.TOKEN_KW_ELSEIF {
		le.NextToken()
		exps = append(exps, parseExp(le))
		le.NextTokenOfKind(lexer.TOKEN_KW_THEN)
		blocks = append(blocks, parseBlock(le))
	}

	if le.LookAhead() == lexer.TOKEN_KW_ELSE {
		le.NextToken()
		exps = append(exps, &ast.TrueExp{le.Line()})
		blocks = append(blocks, parseBlock(le))
	}
	le.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.IfStat{exps, blocks}
}

func parseForStat(le *lexer.Lexer) ast.Stat {
	lineOfFor, _ := le.NextTokenOfKind(lexer.TOKEN_KW_FOR)
	_, name := le.NextIdentifier()
	if le.LookAhead() == lexer.TOKEN_OP_ASSIGN {
		return _finishForNumStat(le, lineOfFor, name)
	} else {
		return _finishForInStat(le, lineOfFor, name)
	}
}

func _finishForNumStat(le *lexer.Lexer, lineOfFor int, varName string) *ast.ForNumStat {
	le.NextTokenOfKind(lexer.TOKEN_OP_ASSIGN)
	initExp := parseExp(le)
	le.NextTokenOfKind(lexer.TOKEN_SEP_COMMA)
	limitExp := parseExp(le)

	var stepExp ast.Exp
	if le.LookAhead() == lexer.TOKEN_SEP_COMMA {
		le.NextToken()
		stepExp = parseExp(le)
	} else {
		stepExp = &ast.IntegerExp{le.Line(), 1}
	}
	lineOfDo, _ := le.NextTokenOfKind(lexer.TOKEN_KW_DO)
	block := parseBlock(le)
	le.NextTokenOfKind(lexer.TOKEN_KW_END)

	return &ast.ForNumStat{
		LineOfFor: lineOfFor,
		LineOfDo:  lineOfDo,
		VarName:   varName,
		InitExp:   initExp,
		LimitExp:  limitExp,
		StepExp:   stepExp,
		Block:     block,
	}
}

func _finishForInStat(le *lexer.Lexer, lineOfFor int, varName string) *ast.ForInStat {
	nameList := _finishNameList(le, varName)
	le.NextTokenOfKind(lexer.TOKEN_KW_IN)
	expList := parseExpList(le)
	lineOfDo, _ := le.NextTokenOfKind(lexer.TOKEN_KW_DO)
	block := parseBlock(le)
	le.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.ForInStat{lineOfDo, nameList, expList, block}
}

func _finishNameList(le *lexer.Lexer, name0 string) []string {
	names := []string{name0}
	if le.LookAhead() == lexer.TOKEN_SEP_COMMA {
		le.NextToken()
		_, name := le.NextIdentifier()
		names = append(names, name)
	}
	return names
}

func parseLocalAssignOrFuncDefStat(le *lexer.Lexer) ast.Stat {
	le.NextTokenOfKind(lexer.TOKEN_KW_LOCAL)
	if le.LookAhead() == lexer.TOKEN_KW_FUNCTION {
		return _finishLocalFuncDefStat(le)
	} else {
		return _finishLocalVarDeclStat(le)
	}
}

func _finishLocalFuncDefStat(le *lexer.Lexer) *ast.LocalFuncDefStat {
	le.NextTokenOfKind(lexer.TOKEN_KW_FUNCTION)
	_, name := le.NextIdentifier()
	fsExp := parseFuncDefExp(le)
	return &ast.LocalFuncDefStat{name, fsExp}
}

func _finishLocalVarDeclStat(le *lexer.Lexer) *ast.LocalVarDeclStat {
	_, name0 := le.NextIdentifier()
	nameList := _finishNameList(le, name0)
	var expList []ast.Exp = nil
	if le.LookAhead() == lexer.TOKEN_OP_ASSIGN {
		le.NextToken()
		expList = parseExpList(le)
	}
	lastLine := le.Line()
	return &ast.LocalVarDeclStat{lastLine, nameList, expList}
}

func parseAssignOrCallStat(le *lexer.Lexer) ast.Stat {
	prefixExp := parsePrefixExp(le)
	if fc, ok := prefixExp.(*ast.FuncCallExp); ok {
		return fc
	} else {
		return parseAssignStat(le, prefixExp)
	}
}

func parseAssignStat(le *lexer.Lexer, var0 ast.Exp) *ast.AssignStat {
	varList := _finishVarList(le, var0)
	le.NextTokenOfKind(lexer.TOKEN_OP_ASSIGN)
	expList := parseExpList(le)
	lastLine := le.Line()
	return &ast.AssignStat{lastLine, varList, expList}

}

func _finishVarList(le *lexer.Lexer, var0 ast.Exp) []ast.Exp {
	vars := []ast.Exp{_checkVar(le, var0)}
	for le.LookAhead() == lexer.TOKEN_SEP_COMMA {
		le.NextToken()
		exp := parsePrefixExp(le)
		vars = append(vars, _checkVar(le, exp))
	}
	return vars
}

func _checkVar(le *lexer.Lexer, exp ast.Exp) ast.Exp {
	switch exp.(type) {
	case *ast.NameExp, *ast.TableAccessExp:
		return exp
	}
	le.NextTokenOfKind(-1)
	panic("unreachable!!!!")
}

func parseFuncDefStat(le *lexer.Lexer) *ast.AssignStat {
	le.NextTokenOfKind(lexer.TOKEN_KW_FUNCTION)
	fnExp, hasColon := _parseFuncName(le)
	fdExp := parseFuncDefExp(le)
	if hasColon {
		fdExp.ParList = append(fdExp.ParList, "")
		copy(fdExp.ParList[1:], fdExp.ParList)
		fdExp.ParList[0] = "self"
	}
	return &ast.AssignStat{
		LastLine: fdExp.Line,
		VarList:  []ast.Exp{fnExp},
		ExpList:  []ast.Exp{fdExp},
	}
}

func _parseFuncName(le *lexer.Lexer) (exp ast.Exp, hasColon bool) {
	line, name := le.NextIdentifier()
	exp = &ast.NameExp{line, name}
	for le.LookAhead() == lexer.TOKEN_SEP_COMMA {
		le.NextToken()
		line, name := le.NextIdentifier()
		idx := &ast.StringExp{line, name}
		exp = &ast.TableAccessExp{line, exp, idx}
	}
	if le.LookAhead() == lexer.TOKEN_SEP_COMMA {
		le.NextToken()
		line, name := le.NextIdentifier()
		idx := &ast.StringExp{line, name}
		exp = &ast.TableAccessExp{line, exp, idx}
		hasColon = true
	}
	return
}

//========================================================= Exp ===============================================
func parseExp(le *lexer.Lexer) ast.Exp {
	return parseExp12(le)
}

func parseExp12(le *lexer.Lexer) ast.Exp {
	exp := parseExp11(le)
	for le.LookAhead() == lexer.TOKEN_OP_OR {
		line, op, _ := le.NextToken()
		_exp := &ast.BinopExp{line, op, exp, parseExp11(le)}
		exp = optimizeLogicalOr(_exp)
	}
	return exp
}

func parseExp11(le *lexer.Lexer) ast.Exp {
	exp := parseExp10(le)
	for le.LookAhead() == lexer.TOKEN_OP_AND {
		line, op, _ := le.NextToken()
		_exp := &ast.BinopExp{line, op, exp, parseExp10(le)}
		exp = optimizeLogicalAnd(_exp)
	}
	return exp
}

func parseExp10(le *lexer.Lexer) ast.Exp {
	exp := parseExp9(le)
	for isExp10Op(le.LookAhead()) {
		line, op, _ := le.NextToken()
		exp = &ast.BinopExp{line, op, exp, parseExp9(le)}
	}
	return exp
}

func isExp10Op(opType int) bool {
	switch opType {
	case lexer.TOKEN_OP_LT, lexer.TOKEN_OP_GT, lexer.TOKEN_OP_LE, lexer.TOKEN_OP_GE, lexer.TOKEN_OP_NE, lexer.TOKEN_OP_EQ:
		return true
	}
	return false
}

// A | B
func parseExp9(le *lexer.Lexer) ast.Exp {
	exp := parseExp8(le)
	for le.LookAhead() == lexer.TOKEN_OP_BOR {
		line, op, _ := le.NextToken()
		_exp := &ast.BinopExp{line, op, exp, parseExp8(le)}
		exp = optimizeBitwiseBinaryOp(_exp)
	}
	return exp
}

func parseExp8(le *lexer.Lexer) ast.Exp {
	exp := parseExp7(le)
	for le.LookAhead() == lexer.TOKEN_OP_WAVE {
		line, op, _ := le.NextToken()
		exp = &ast.BinopExp{line, op, exp, parseExp7(le)}
	}
	return exp
}

// A & B
func parseExp7(le *lexer.Lexer) ast.Exp {
	exp := parseExp6(le)
	for le.LookAhead() == lexer.TOKEN_OP_BAND {
		line, op, _ := le.NextToken()
		_exp := &ast.BinopExp{line, op, exp, parseExp6(le)}
		exp = optimizeBitwiseBinaryOp(_exp)
	}
	return exp
}

// A << B , A >>B
func parseExp6(le *lexer.Lexer) ast.Exp {
	exp := parseExp5(le)
	for le.LookAhead() == lexer.TOKEN_OP_SHR || le.LookAhead() == lexer.TOKEN_OP_SHL {
		line, op, _ := le.NextToken()
		_exp := &ast.BinopExp{line, op, exp, parseExp5(le)}
		exp = optimizeBitwiseBinaryOp(_exp)
	}
	return exp
}

func parseExp5(le *lexer.Lexer) ast.Exp {
	exp := parseExp4(le)
	if le.LookAhead() != lexer.TOKEN_OP_CONCAT {
		return exp
	}
	var line int = 0
	exps := []ast.Exp{exp}
	for le.LookAhead() == lexer.TOKEN_OP_CONCAT {
		line, _, _ = le.NextToken()
		exps = append(exps, parseExp4(le))
	}
	return &ast.ConcatExp{line, exps}
}

// A + B,  A - B
func parseExp4(le *lexer.Lexer) ast.Exp {
	exp := parseExp3(le)
	if le.LookAhead() == lexer.TOKEN_OP_ADD || le.LookAhead() == lexer.TOKEN_OP_MINUS {
		line, op, _ := le.NextToken()
		_exp := &ast.BinopExp{line, op, exp, parseExp3(le)}
		exp = optimizeArithBinaryOp(_exp)
	}
	return exp
}

func parseExp3(le *lexer.Lexer) ast.Exp {
	exp := parseExp2(le)
	for le.LookAhead() == lexer.TOKEN_OP_MUL || le.LookAhead() == lexer.TOKEN_OP_DIV || le.LookAhead() == lexer.TOKEN_OP_IDIV || le.LookAhead() == lexer.TOKEN_OP_MOD {
		line, op, _ := le.NextToken()
		_exp := &ast.BinopExp{line, op, exp, parseExp2(le)}
		exp = optimizeArithBinaryOp(_exp)
	}
	return exp
}

func parseExp2(le *lexer.Lexer) ast.Exp {
	switch le.LookAhead() {
	case lexer.TOKEN_OP_LEN, lexer.TOKEN_OP_NOT, lexer.TOKEN_OP_MINUS, lexer.TOKEN_OP_BNOT:
		line, op, _ := le.NextToken()
		exp := &ast.UnopExp{line, op, parseExp2(le)}
		return optimizeUnaryOp(exp)
	}
	return parseExp1(le)
}

func parseExp1(le *lexer.Lexer) ast.Exp {
	exp := parseExp0(le)
	if le.LookAhead() == lexer.TOKEN_OP_POW {
		line, op, _ := le.NextToken()
		exp = &ast.BinopExp{line, op, exp, parseExp1(le)}
		exp = optimizePow(exp)
	}
	return exp
}

func parseExp0(le *lexer.Lexer) ast.Exp {
	switch le.LookAhead() {
	case lexer.TOKEN_VARARG:
		line, _, _ := le.NextToken()
		return &ast.VarargExp{line}
	case lexer.TOKEN_KW_NIL:
		line, _, _ := le.NextToken()
		return &ast.NilExp{line}
	case lexer.TOKEN_KW_FALSE:
		line, _, _ := le.NextToken()
		return &ast.FalseExp{line}
	case lexer.TOKEN_KW_TRUE:
		line, _, _ := le.NextToken()
		return &ast.TrueExp{line}
	case lexer.TOKEN_STRING:
		line, _, str := le.NextToken()
		return &ast.StringExp{line, str}
	case lexer.TOKEN_NUMBER:
		return parseNumberExp(le)
	case lexer.TOKEN_SEP_LCURLY:
		return parseTableConstructorExp(le)
	case lexer.TOKEN_KW_FUNCTION:
		le.NextToken()
		return parseFuncDefExp(le)
	default:
		return parsePrefixExp(le)
	}
}

func parseNumberExp(le *lexer.Lexer) ast.Exp {
	line, _, token := le.NextToken()
	if i, ok := number.ParseInteger(token); ok {
		return &ast.IntegerExp{line, i}
	} else if f, ok := number.ParseFloat(token); ok {
		return &ast.FloatExp{line, f}
	} else {
		panic("not a number: " + token)
	}
}



func parseFuncDefExp(le *lexer.Lexer) *ast.FuncDefExp {
	line := le.Line()
	le.NextTokenOfKind(lexer.TOKEN_SEP_LPAREN)
	parList, isVararg := _parseParList(le)
	le.NextTokenOfKind(lexer.TOKEN_SEP_RPAREN)
	block := parseBlock(le)
	lastLine, _ := le.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.FuncDefExp{line, lastLine, parList, isVararg, block}
}

func _parseParList(le *lexer.Lexer) (names []string, isVararg bool) {
	switch le.LookAhead() {
	case lexer.TOKEN_SEP_RPAREN:
		return nil, false
	case lexer.TOKEN_VARARG:
		return nil, true
	}
	_, name := le.NextIdentifier()
	names = append(names, name)
	for le.LookAhead() == lexer.TOKEN_SEP_COMMA {
		le.NextToken()
		if le.LookAhead() == lexer.TOKEN_IDENTIFIER {
			_, name := le.NextIdentifier()
			names = append(names, name)
		} else {
			le.NextTokenOfKind(lexer.TOKEN_VARARG)
			isVararg = true
			break
		}
	}
	return
}

func parseTableConstructorExp(le *lexer.Lexer) *ast.TableConstructorExp {
	line := le.Line()
	le.NextTokenOfKind(lexer.TOKEN_SEP_LCURLY)
	ks, vs := _parseFieldList(le)
	le.NextTokenOfKind(lexer.TOKEN_SEP_RCURLY)
	lastLine := le.Line()
	return &ast.TableConstructorExp{line, lastLine, ks, vs}
}

func _parseFieldList(le *lexer.Lexer) (ks, vs []ast.Exp) {
	if le.LookAhead() != lexer.TOKEN_SEP_RCURLY {
		k, v := _parseField(le)
		ks = append(ks, k)
		vs = append(vs, v)
		for _isFieldSep(le.LookAhead()) {
			le.NextToken()
			if le.LookAhead() != lexer.TOKEN_SEP_RCURLY {
				k, v := _parseField(le)
				ks = append(ks, k)
				vs = append(vs, v)
			} else {
				break
			}
		}
	}
	return
}

func _isFieldSep(tokenKind int) bool {
	return tokenKind == lexer.TOKEN_SEP_COMMA || tokenKind == lexer.TOKEN_SEP_SEMI
}

func _parseField(le *lexer.Lexer) (k, v ast.Exp) {
	if le.LookAhead() == lexer.TOKEN_SEP_LBRACK {
		le.NextToken()
		k = parseExp(le)
		le.NextTokenOfKind(lexer.TOKEN_SEP_RBRACK)
		le.NextTokenOfKind(lexer.TOKEN_OP_ASSIGN)
		v = parseExp(le)
		return
	}
	exp := parseExp(le)
	if nameExp, ok := exp.(*ast.NameExp); ok {
		le.NextToken()
		k = &ast.StringExp{nameExp.Line, nameExp.Name}
		v = parseExp(le)
		return
	}
	return nil, exp
}

func parsePrefixExp(le *lexer.Lexer) ast.Exp {
	var exp ast.Exp
	if le.LookAhead() == lexer.TOKEN_IDENTIFIER {
		line, name := le.NextIdentifier()
		exp = &ast.NameExp{line, name}
	} else {
		exp = parseParensExp(le)
	}
	return _finishPrefixExp(le, exp)
}

func _finishPrefixExp(le *lexer.Lexer, exp ast.Exp) ast.Exp {
	for {
		tokenType := le.LookAhead()
		switch tokenType {
		case lexer.TOKEN_SEP_LBRACK:
			le.NextToken()
			keyExp := parseExp(le)
			le.NextTokenOfKind(lexer.TOKEN_SEP_RBRACK)
			exp = &ast.TableAccessExp{le.Line(), exp, keyExp}
		case lexer.TOKEN_SEP_DOT:
			le.NextToken()
			line, name := le.NextIdentifier()
			keyExp := &ast.StringExp{line, name}
			exp = &ast.TableAccessExp{le.Line(), exp, keyExp}
		case lexer.TOKEN_SEP_COLON, lexer.TOKEN_SEP_LPAREN, lexer.TOKEN_SEP_LCURLY, lexer.TOKEN_STRING:
			exp = _finishFuncCallExp(le, exp)
		default:
			return exp
		}
	}
	return exp
}

func parseParensExp(le *lexer.Lexer) ast.Exp {
	le.NextTokenOfKind(lexer.TOKEN_SEP_LPAREN)
	exp := parseExp(le)
	le.NextTokenOfKind(lexer.TOKEN_SEP_RPAREN)
	switch exp.(type) {
	case *ast.VarargExp, *ast.FuncCallExp, *ast.NameExp, *ast.TableAccessExp:
		return &ast.ParensExp{exp}
	}
	return exp
}

func _finishFuncCallExp(le *lexer.Lexer, prefixExp ast.Exp) *ast.FuncCallExp {
	nameExp := _parseNameExp(le)
	line := le.Line()
	args := _parseArgs(le)
	lastLine := le.Line()
	return &ast.FuncCallExp{line, lastLine, prefixExp, nameExp, args}
}

func _parseNameExp(le *lexer.Lexer) *ast.StringExp {
	if le.LookAhead() == lexer.TOKEN_SEP_COLON {
		le.NextToken()
		line, name := le.NextIdentifier()
		return &ast.StringExp{line, name}
	}
	return nil
}

func _parseArgs(le *lexer.Lexer) (args []ast.Exp) {
	switch le.LookAhead() {
	case lexer.TOKEN_SEP_LPAREN:
		le.NextToken()
		if le.LookAhead() != lexer.TOKEN_SEP_RPAREN {
			args = parseExpList(le)
		}
		le.NextTokenOfKind(lexer.TOKEN_SEP_RPAREN)
	case lexer.TOKEN_SEP_LCURLY:
		args = []ast.Exp{parseTableConstructorExp(le)}
	default:
		line, str := le.NextTokenOfKind(lexer.TOKEN_STRING)
		args = []ast.Exp{&ast.StringExp{line, str}}
	}
	return
}

//{{{{{{{ 优化相关
func optimizeLogicalOr(exp *ast.BinopExp) ast.Exp {
	if isTrue(exp.Exp1) {
		return exp.Exp1 // true or x => true
	}
	if isFalse(exp.Exp1) && !isVarargOrFuncCall(exp.Exp2) {
		return exp.Exp2 // false or x => x
	}
	return exp
}

func optimizeLogicalAnd(exp *ast.BinopExp) ast.Exp {
	if isFalse(exp.Exp1) {
		return exp.Exp1 // false and x => false
	}
	if isTrue(exp.Exp1) && !isVarargOrFuncCall(exp.Exp2) {
		return exp.Exp2 // true and x => x
	}
	return exp
}

func optimizeBitwiseBinaryOp(exp *ast.BinopExp) ast.Exp {
	if i, ok := castToInt(exp.Exp1); ok {
		if j, ok := castToInt(exp.Exp2); ok {
			switch exp.Op {
			case lexer.TOKEN_OP_BAND:
				return &ast.IntegerExp{exp.Line, i & j}
			case lexer.TOKEN_OP_BOR:
				return &ast.IntegerExp{exp.Line, i | j}
			case lexer.TOKEN_OP_BXOR:
				return &ast.IntegerExp{exp.Line, i ^ j}
			case lexer.TOKEN_OP_SHL:
				return &ast.IntegerExp{exp.Line, number.ShiftLeft(i, j)}
			case lexer.TOKEN_OP_SHR:
				return &ast.IntegerExp{exp.Line, number.ShiftRight(i, j)}
			}
		}
	}
	return exp
}

func optimizeArithBinaryOp(exp *ast.BinopExp) ast.Exp {
	if x, ok := exp.Exp1.(*ast.IntegerExp); ok {
		if y, ok := exp.Exp2.(*ast.IntegerExp); ok {
			switch exp.Op {
			case lexer.TOKEN_OP_ADD:
				return &ast.IntegerExp{exp.Line, x.Val + y.Val}
			case lexer.TOKEN_OP_SUB:
				return &ast.IntegerExp{exp.Line, x.Val - y.Val}
			case lexer.TOKEN_OP_MUL:
				return &ast.IntegerExp{exp.Line, x.Val * y.Val}
			case lexer.TOKEN_OP_IDIV:
				if y.Val != 0 {
					return &ast.IntegerExp{exp.Line, number.IFloorDiv(x.Val, y.Val)}
				}
			case lexer.TOKEN_OP_MOD:
				if y.Val != 0 {
					return &ast.IntegerExp{exp.Line, number.IMod(x.Val, y.Val)}
				}
			}
		}
	}
	if f, ok := castToFloat(exp.Exp1); ok {
		if g, ok := castToFloat(exp.Exp2); ok {
			switch exp.Op {
			case lexer.TOKEN_OP_ADD:
				return &ast.FloatExp{exp.Line, f + g}
			case lexer.TOKEN_OP_SUB:
				return &ast.FloatExp{exp.Line, f - g}
			case lexer.TOKEN_OP_MUL:
				return &ast.FloatExp{exp.Line, f * g}
			case lexer.TOKEN_OP_DIV:
				if g != 0 {
					return &ast.FloatExp{exp.Line, f / g}
				}
			case lexer.TOKEN_OP_IDIV:
				if g != 0 {
					return &ast.FloatExp{exp.Line, number.FFloorDiv(f, g)}
				}
			case lexer.TOKEN_OP_MOD:
				if g != 0 {
					return &ast.FloatExp{exp.Line, number.FMod(f, g)}
				}
			case lexer.TOKEN_OP_POW:
				return &ast.FloatExp{exp.Line, math.Pow(f, g)}
			}
		}
	}
	return exp
}

func optimizePow(exp ast.Exp) ast.Exp {
	if binop, ok := exp.(*ast.BinopExp); ok {
		if binop.Op == lexer.TOKEN_OP_POW {
			binop.Exp2 = optimizePow(binop.Exp2)
		}
		return optimizeArithBinaryOp(binop)
	}
	return exp
}

func optimizeUnaryOp(exp *ast.UnopExp) ast.Exp {
	switch exp.Op {
	case lexer.TOKEN_OP_UNM:
		return optimizeUnm(exp)
	case lexer.TOKEN_OP_NOT:
		return optimizeNot(exp)
	case lexer.TOKEN_OP_BNOT:
		return optimizeBnot(exp)
	default:
		return exp
	}
}

func optimizeUnm(exp *ast.UnopExp) ast.Exp {
	switch x := exp.Exp.(type) { // number?
	case *ast.IntegerExp:
		x.Val = -x.Val
		return x
	case *ast.FloatExp:
		if x.Val != 0 {
			x.Val = -x.Val
			return x
		}
	}
	return exp
}

func optimizeNot(exp *ast.UnopExp) ast.Exp {
	switch exp.Exp.(type) {
	case *ast.NilExp, *ast.FalseExp: // false
		return &ast.TrueExp{exp.Line}
	case *ast.TrueExp, *ast.IntegerExp, *ast.FloatExp, *ast.StringExp: // true
		return &ast.FalseExp{exp.Line}
	default:
		return exp
	}
}

func optimizeBnot(exp *ast.UnopExp) ast.Exp {
	switch x := exp.Exp.(type) { // number?
	case *ast.IntegerExp:
		x.Val = ^x.Val
		return x
	case *ast.FloatExp:
		if i, ok := number.FloatToInteger(x.Val); ok {
			return &ast.IntegerExp{x.Line, ^i}
		}
	}
	return exp
}

func isFalse(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.FalseExp, *ast.NilExp:
		return true
	default:
		return false
	}
}

func isTrue(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.TrueExp, *ast.IntegerExp, *ast.FloatExp, *ast.StringExp:
		return true
	default:
		return false
	}
}

// todo
func isVarargOrFuncCall(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.VarargExp, *ast.FuncCallExp:
		return true
	}
	return false
}

func castToInt(exp ast.Exp) (int64, bool) {
	switch x := exp.(type) {
	case *ast.IntegerExp:
		return x.Val, true
	case *ast.FloatExp:
		return number.FloatToInteger(x.Val)
	default:
		return 0, false
	}
}

func castToFloat(exp ast.Exp) (float64, bool) {
	switch x := exp.(type) {
	case *ast.IntegerExp:
		return float64(x.Val), true
	case *ast.FloatExp:
		return x.Val, true
	default:
		return 0, false
	}
}

//}}}}}}}
