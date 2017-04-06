package lexer

import (
	"testing"

	"github.com/emb/play/monkey/token"
)

func TestNextToken(t *testing.T) {
	input := "=+(){},;"
	tests := []struct {
		wantType    token.Type
		wantLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tc := range tests {
		tok := l.NextToken()
		if tok.Type != tc.wantType {
			t.Fatalf("tests[%d] bad token type want=%q, got=%q",
				i, tc.wantType, tok.Type)
		}
		if tok.Literal != tc.wantLiteral {
			t.Fatalf("tests[%d] bad literal want=%q, got=%q",
				i, tc.wantLiteral, tok.Literal)
		}
	}
}
