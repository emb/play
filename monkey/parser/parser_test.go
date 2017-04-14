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
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("stmt.Expression is of type %T, want *ast.Identifier",
			stmt.Expression)
	}
	if ident.Value != "monkeySee" {
		t.Errorf("ident.Value is %s, want monkeySee", ident.Value)
	}
	if ident.TokenLiteral() != "monkeySee" {
		t.Errorf("ident.TokenLiteral() is %s, want monkeySee",
			ident.TokenLiteral())
	}
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

func TestParsingPrefixExpressions(t *testing.T) {
	tests := []struct {
		input  string
		wantOp string
		want   int64
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
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
		testIntegerLiteral(t, expr.Right, tc.want)
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	tests := []struct {
		input     string
		wantLeft  int64
		wantOp    string
		wantRight int64
	}{
		{"1 + 2;", 1, "+", 2},
		{"1 - 2;", 1, "-", 2},
		{"1 * 2;", 1, "*", 2},
		{"1 / 2;", 1, "/", 2},
		{"1 > 2;", 1, ">", 2},
		{"1 < 2;", 1, "<", 2},
		{"1 == 2;", 1, "==", 2},
		{"1 != 2;", 1, "!=", 2},
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
		expr, ok := stmt.Expression.(*ast.InfixExpr)
		if !ok {
			t.Fatalf("stmt.Expression is of type %T, want *ast.InfixExpr",
				stmt.Expression)
		}
		testIntegerLiteral(t, expr.Left, tc.wantLeft)
		if expr.Operator != tc.wantOp {
			t.Errorf("expr.Operator is %s, want %s",
				expr.Operator, tc.wantOp)

		}
		testIntegerLiteral(t, expr.Right, tc.wantRight)
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
		b, ok := stmt.Expression.(*ast.Boolean)
		if !ok {
			t.Fatalf("stmt.Expression is of type %T, want *ast.Boolean",
				stmt.Expression)
		}
		if b.Value != tc.want {
			t.Errorf(".Value is %t, want %t", b.Value, tc.want)
		}
		literal := fmt.Sprintf("%t", tc.want)
		if b.Token.Literal != literal {
			t.Errorf(".Token.Literal is %s, want %s",
				b.Token.Literal, literal)
		}
	}
}
