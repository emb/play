package parser

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/emb/play/monkey/ast"
	"github.com/emb/play/monkey/lexer"
)

func TestLetStatements(t *testing.T) {
	input := `
let x = 5;
let y = 9;
let foobar = 989865;
`
	l := lexer.New(input)
	p := New(l)

	program := p.Program()
	checkErrors(t, p)
	if program == nil {
		t.Fatal("p.Progarm() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements has %d, want 3",
			len(program.Statements))
	}

	tests := []struct {
		wantIdent string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tc := range tests {
		stmt := program.Statements[i]
		testLet(t, stmt, tc.wantIdent)
	}
}

func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 10;
return 94897597;
`
	parser := New(lexer.New(input))
	program := parser.Program()
	checkErrors(t, parser)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements has %d, want 3",
			len(program.Statements))
	}

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

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("program.Statements[0] is of type %T, want *ast.ExpressionStmt",
			program.Statements[0])
	}
	testIdent(t, stmt.Expression, "monkeySee")
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "7;"
	parser := New(lexer.New(input))
	program := parser.Program()
	checkErrors(t, parser)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("program.Statements[0] is of type %T, want *ast.ExpressionStmt",
			program.Statements[0])
	}
	testIntegerLiteral(t, stmt.Expression, 7)
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

		if len(program.Statements) != 1 {
			t.Fatalf("program has %d statements, want 1",
				len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
		if !ok {
			t.Fatalf("program.Statements[0] is of type %T, want *ast.ExpressionStmt",
				program.Statements[0])
		}
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

		if len(program.Statements) != 1 {
			t.Fatalf("program has %d statements, want 1",
				len(program.Statements))

		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
		if !ok {
			t.Fatalf("program.Statements[0] is of type %T, want *ast.ExpressionStmt",
				program.Statements[0])
		}
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
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3 > 5) == false)"},
		{"3 < 5 == true", "((3 < 5) == true)"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 / (5 + 5)", "(2 / (5 + 5))"},
		{"-(5 + 5)", "(-(5 + 5))"},
		{"!(true == true)", "(!(true == true))"},
	}

	for i, tc := range tests {
		parse := New(lexer.New(tc.input))
		program := parse.Program()
		checkErrors(t, parse)

		if program.String() != tc.want {
			t.Errorf("test[%d] parse.Program() returns %q, want %q",
				i, program, tc.want)
		}
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

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements has %d, want 1",
				len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
		if !ok {
			t.Fatalf("program.Statements[0] is of type %T, want *ast.ExpressionStmt",
				program.Statements[0])
		}
		testBooleanLiteral(t, stmt.Expression, tc.want)
	}
}

func TestIfExpresssion(t *testing.T) {
	input := `if (x < y) { x }`
	parse := New(lexer.New(input))
	program := parse.Program()
	checkErrors(t, parse)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements has %d, want 1",
			len(program.Statements))

	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("program.Statements[0] is of type %T, want *ast.ExpressionStmt",
			program.Statements[0])
	}
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

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements has %d, want 1",
			len(program.Statements))

	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("program.Statements[0] is of type %T, want *ast.ExpressionStmt",
			program.Statements[0])
	}
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

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements has %d, want 1",
			len(program.Statements))

	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("program.Statements[0] is of type %T, want *ast.ExpressionStmt",
			program.Statements[0])
	}
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

		stmt := program.Statements[0].(*ast.ExpressionStmt)
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
