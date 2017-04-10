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

	let, ok := s.(*ast.Let)
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
