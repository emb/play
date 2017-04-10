package parser

import (
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
