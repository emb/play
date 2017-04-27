package parser

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/emb/play/monkey/ast"
	"github.com/emb/play/monkey/lexer"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input     string
		wantIdent string
		wantValue interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foo = bar", "foo", "bar"},
	}
	for i, tc := range tests {
		t.Logf("test[%d] input %q", i, tc.input)
		parse := New(lexer.New(tc.input))
		program := parse.Program()
		checkErrors(t, parse)

		ensureStatements(t, program, 1)
		stmt := program.Statements[0]
		testLet(t, stmt, tc.wantIdent)
		v := stmt.(*ast.LetStmt).Value
		testLiteralExpr(t, v, tc.wantValue)
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
	}{
		{"return 5;", 5},
		{"return false;", false},
		{"return foo;", "foo"},
	}
	for i, tc := range tests {
		t.Logf("test[%d] input %q", i, tc.input)
		parse := New(lexer.New(tc.input))
		program := parse.Program()
		checkErrors(t, parse)

		ensureStatements(t, program, 1)
		stmt := program.Statements[0]
		testReturn(t, stmt)
		v := stmt.(*ast.ReturnStmt).Value
		testLiteralExpr(t, v, tc.want)
	}
	input := `
return 5;
return 10;
return 94897597;
`
	parser := New(lexer.New(input))
	program := parser.Program()
	checkErrors(t, parser)

	ensureStatements(t, program, 3)
	for _, stmt := range program.Statements {
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			t.Errorf("statement type is %T, want *ast.Return", stmt)
			continue
		}
		if ret.TokenLiteral() != "return" {
			t.Errorf(`.TokenLiteral() is %q, want "return"`,
				ret.TokenLiteral())
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "monkeySee;" // A valid expression with an identifier
	parser := New(lexer.New(input))
	program := parser.Program()
	checkErrors(t, parser)

	stmt := firstExpression(t, program)
	testIdent(t, stmt.Expression, "monkeySee")
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "7;"
	parser := New(lexer.New(input))
	program := parser.Program()
	checkErrors(t, parser)

	stmt := firstExpression(t, program)
	testIntegerLiteral(t, stmt.Expression, 7)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello string!";`
	parse := New(lexer.New(input))
	program := parse.Program()
	checkErrors(t, parse)

	stmt := firstExpression(t, program)
	str, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is of type %T, want *ast.StringLiteral",
			stmt.Expression)
	}
	if str.Value != "hello string!" {
		t.Errorf(`str.Value is %q, want "hello string!"`, str.Value)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	tests := []struct {
		input  string
		wantOp string
		want   interface{}
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for i, tc := range tests {
		t.Logf("test[%d] input %s", i, tc.input)

		parser := New(lexer.New(tc.input))
		program := parser.Program()
		checkErrors(t, parser)

		stmt := firstExpression(t, program)
		expr, ok := stmt.Expression.(*ast.PrefixExpr)
		if !ok {
			t.Fatalf("stmt.Expression is of type %T, want *ast.PrefixExpr",
				stmt.Expression)
		}
		if expr.Operator != tc.wantOp {
			t.Errorf("expr.Operator is %s, want %s",
				expr.Operator, tc.wantOp)
		}
		testLiteralExpr(t, expr.Right, tc.want)
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	tests := []struct {
		input     string
		wantLeft  interface{}
		wantOp    string
		wantRight interface{}
	}{
		{"1 + 2;", 1, "+", 2},
		{"1 - 2;", 1, "-", 2},
		{"1 * 2;", 1, "*", 2},
		{"1 / 2;", 1, "/", 2},
		{"1 > 2;", 1, ">", 2},
		{"1 < 2;", 1, "<", 2},
		{"1 == 2;", 1, "==", 2},
		{"1 != 2;", 1, "!=", 2},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}
	for i, tc := range tests {
		t.Logf("test[%d] %s", i, tc.input)
		parse := New(lexer.New(tc.input))
		program := parse.Program()
		checkErrors(t, parse)

		stmt := firstExpression(t, program)
		testInfix(t, stmt.Expression, tc.wantLeft, tc.wantOp, tc.wantRight)
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b / c", "(a + (b / c))"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4; - 5 * 5", "(3 + 4)((-5) * 5)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))"},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3 > 5) == false)"},
		{"3 < 5 == true", "((3 < 5) == true)"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 / (5 + 5)", "(2 / (5 + 5))"},
		{"-(5 + 5)", "(-(5 + 5))"},
		{"!(true == true)", "(!(true == true))"},
		{"a + add(b * c) + d", "((a + add((b * c))) + d)"},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			parse := New(lexer.New(tc.input))
			program := parse.Program()
			checkErrors(t, parse)

			if program.String() != tc.want {
				t.Errorf("test[%d] parse.Program() returns %q, want %q",
					i, program, tc.want)
			}

		})
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true;", true},
		{"false;", false},
	}
	for i, tc := range tests {
		t.Logf("test[%d] input %q", i, tc.input)
		parse := New(lexer.New(tc.input))
		program := parse.Program()
		checkErrors(t, parse)

		stmt := firstExpression(t, program)
		testBooleanLiteral(t, stmt.Expression, tc.want)
	}
}

func TestIfExpresssion(t *testing.T) {
	input := `if (x < y) { x }`
	parse := New(lexer.New(input))
	program := parse.Program()
	checkErrors(t, parse)

	stmt := firstExpression(t, program)
	ifexpr, ok := stmt.Expression.(*ast.IfExpr)
	if !ok {
		t.Fatalf("ifexpr is of type %T, want *ast.IfExpr", ifexpr)
	}
	testInfix(t, ifexpr.Condition, "x", "<", "y")
	if len(ifexpr.Consequence.Statements) != 1 {
		t.Fatalf("ifexp.Consequence.Statements has %d statements, want 1",
			len(ifexpr.Consequence.Statements))
	}
	consq, ok := ifexpr.Consequence.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("Consequence.Statements[0] is of type %T, want *ast.ExpressionStmt",
			ifexpr.Consequence.Statements[0])
	}
	testIdent(t, consq.Expression, "x")
	if ifexpr.Alternative != nil {
		t.Errorf("ifexpr.Alternative is %+v, want nil",
			ifexpr.Alternative)
	}
}

func TestIfAlternativeExpresssion(t *testing.T) {
	input := `if (x < y) { x } else { y }`
	parse := New(lexer.New(input))
	program := parse.Program()
	checkErrors(t, parse)

	stmt := firstExpression(t, program)
	ifexpr, ok := stmt.Expression.(*ast.IfExpr)
	if !ok {
		t.Fatalf("ifexpr is of type %T, want *ast.IfExpr", ifexpr)
	}
	if ifexpr.Alternative == nil {
		t.Error("ifexpr.Alternative is nil")
	}
	if len(ifexpr.Alternative.Statements) != 1 {
		t.Fatalf("ifexp.Alternative.Statements has %d statements, want 1",
			len(ifexpr.Alternative.Statements))
	}
	alt, ok := ifexpr.Alternative.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("Alternative.Statements[0] is of type %T, want *ast.ExpressionStmt",
			ifexpr.Consequence.Statements[0])
	}
	testIdent(t, alt.Expression, "y")
}

func TestFuctionLiteralExpression(t *testing.T) {
	input := `fn(x, y) { x+y; }`
	parse := New(lexer.New(input))
	program := parse.Program()
	checkErrors(t, parse)

	stmt := firstExpression(t, program)
	fn, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is of type %T, want *ast.FunctionLiteral",
			stmt.Expression)
	}
	if len(fn.Parameters) != 2 {
		t.Fatalf("fn has %d parameters, want 2", len(fn.Parameters))
	}
	testLiteralExpr(t, fn.Parameters[0], "x")
	testLiteralExpr(t, fn.Parameters[1], "y")
	if len(fn.Body.Statements) != 1 {
		t.Fatalf("fn.Body has %d statements, want 1",
			len(fn.Body.Statements))
	}
	body, ok := fn.Body.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("body.Statement[0] is of type %T, want *ast.ExpressionStmt",
			fn.Body.Statements[0])
	}
	testInfix(t, body.Expression, "x", "+", "y")
}

func TestFunctionParametersParsing(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{input: "fn() {};", want: []string{}},
		{input: "fn(x) {};", want: []string{"x"}},
		{input: "fn(x,y,z) {};", want: []string{"x", "y", "z"}},
	}
	for i, tc := range tests {
		t.Logf("test[%d] input %q", i, tc.input)
		parse := New(lexer.New(tc.input))
		program := parse.Program()
		checkErrors(t, parse)

		stmt := firstExpression(t, program)
		fn := stmt.Expression.(*ast.FunctionLiteral)
		if len(fn.Parameters) != len(tc.want) {
			t.Errorf("fn has %d params, want %d",
				len(fn.Parameters), len(tc.want))
		}
		for i, p := range tc.want {
			testLiteralExpr(t, fn.Parameters[i], p)
		}
	}
}

func TestCallExpression(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"
	parse := New(lexer.New(input))
	program := parse.Program()
	checkErrors(t, parse)

	stmt := firstExpression(t, program)
	call, ok := stmt.Expression.(*ast.CallExpr)
	if !ok {
		t.Fatalf("stmt.Expression is of type %T, want *ast.CallExpr",
			stmt.Expression)
	}
	testIdent(t, call.Function, "add")
	if len(call.Arguments) != 3 {
		t.Fatalf("call has %d arguments, want 3", len(call.Arguments))
	}
	testLiteralExpr(t, call.Arguments[0], 1)
	testInfix(t, call.Arguments[1], 2, "*", 3)
	testInfix(t, call.Arguments[2], 4, "+", 5)
}

func TestArrayLiteralExpressions(t *testing.T) {
	input := "[1, 2 * 3, 3 + 3]"
	parse := New(lexer.New(input))
	program := parse.Program()
	checkErrors(t, parse)

	stmt := firstExpression(t, program)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is of type %T, want *ast.ArrayLiteral",
			stmt.Expression)
	}
	if len(array.Elements) != 3 {
		t.Fatalf("len(array) is %d, want 3", len(array.Elements))
	}
	testIntegerLiteral(t, array.Elements[0], 1)
	testInfix(t, array.Elements[1], 2, "*", 3)
	testInfix(t, array.Elements[2], 3, "+", 3)
}

func TestIndexEpression(t *testing.T) {
	input := "mylist[3+3]"
	parse := New(lexer.New(input))
	program := parse.Program()
	checkErrors(t, parse)

	stmt := firstExpression(t, program)
	iexp, ok := stmt.Expression.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("stmt.Expression is of type %T, want *ast.IndexExpr",
			stmt.Expression)
	}
	testIdent(t, iexp.Left, "mylist")
	testInfix(t, iexp.Index, 3, "+", 3)
}

func ensureStatements(t *testing.T, program *ast.Program, n int) {
	if len(program.Statements) != n {
		t.Fatalf("program.Statements has %d, want %d",
			len(program.Statements), n)

	}
}

func firstExpression(t *testing.T, program *ast.Program) *ast.ExpressionStmt {
	ensureStatements(t, program, 1)
	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("program.Statements[0] is of type %T, want *ast.ExpressionStmt",
			program.Statements[0])
	}
	return stmt
}

func testIdent(t *testing.T, expr ast.Expression, value string) {
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expr is of type %T, want *ast.Identifier", expr)
	}
	if ident.Value != value {
		t.Errorf("ident.Value is %s, want %s", ident.Value, value)
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral() is %s, want %s",
			ident.TokenLiteral(), value)
	}
}

func testIntegerLiteral(t *testing.T, expr ast.Expression, value int64) {
	literal, ok := expr.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is of type %T, want *ast.IntegerLiteral",
			expr)
	}
	if literal.Value != value {
		t.Errorf("literal.Value is %d, want 7", literal.Value)
	}
	if literal.TokenLiteral() != strconv.FormatInt(value, 10) {
		t.Errorf("literal.TokenLiteral() is %s, want 7",
			literal.TokenLiteral())
	}
}

func testLet(t *testing.T, s ast.Statement, name string) {
	if s.TokenLiteral() != "let" {
		t.Errorf(`s.TokenLiteral has %q, want "let"`, s.TokenLiteral())
	}

	let, ok := s.(*ast.LetStmt)
	if !ok {
		t.Errorf("statement type is %T, want *ast.Let", s)
	}

	if let.Name.Value != name {
		t.Errorf(`.Name.Value is %q, want %q`, let.Name.Value, name)
	}

	if let.Name.TokenLiteral() != name {
		t.Errorf(".Name.TokenLiteral is %q, want %q",
			let.Name.TokenLiteral(), name)
	}
}

func testLiteralExpr(t *testing.T, expr ast.Expression, want interface{}) {
	switch v := want.(type) {
	case int:
		testIntegerLiteral(t, expr, int64(v))
	case int64:
		testIntegerLiteral(t, expr, v)
	case string:
		testIdent(t, expr, v)
	case bool:
		testBooleanLiteral(t, expr, v)
	default:
		t.Errorf("unexpected type of expression %T", expr)
	}
}

func testInfix(t *testing.T, expr ast.Expression, left interface{}, op string, right interface{}) {
	infix, ok := expr.(*ast.InfixExpr)
	if !ok {
		t.Fatalf("exp is of type %T, want *ast.InfixExpr", expr)
	}
	testLiteralExpr(t, infix.Left, left)
	if infix.Operator != op {
		t.Errorf("expr.Operator is %s, want %s", infix.Operator, op)

	}
	testLiteralExpr(t, infix.Right, right)

}

func testBooleanLiteral(t *testing.T, expr ast.Expression, value bool) {
	b, ok := expr.(*ast.Boolean)
	if !ok {
		t.Fatalf("expr is of type %T, want *ast.Boolean", expr)
	}
	if b.Value != value {
		t.Errorf(".Value is %t, want %t", b.Value, value)
	}
	literal := fmt.Sprintf("%t", value)
	if b.Token.Literal != literal {
		t.Errorf(".Token.Literal is %s, want %s",
			b.Token.Literal, literal)
	}

}

func testReturn(t *testing.T, stmt ast.Statement) {
	ret, ok := stmt.(*ast.ReturnStmt)
	if !ok {
		t.Errorf("statement type is %T, want *ast.Return", stmt)
	}
	if ret.TokenLiteral() != "return" {
		t.Errorf(`.TokenLiteral() is %q, want "return"`,
			ret.TokenLiteral())
	}
}

func checkErrors(t *testing.T, p *Parser) {
	errors := p.errors
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
