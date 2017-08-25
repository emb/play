package scan

import (
	"strconv"
	"testing"
)

type tok struct {
	typ Type
	lit string
}

const whitespace = " \t \n\n"

var testTokens = [...]tok{
	{Comment, "# General identifiers\n"},
	{Identifier, "foobar"},
	{Identifier, "foobar_l33t"},
	{Identifier, "x"},
	{String, `"eek!"`},
	{String, `"monkey 'the gopher' twain"`},
	{String, `"monkey \"the gopher\" twain"`},
	{String, `"It is 12\" long."`},
	{Integer, "1"},
	{Integer, "94876678"},

	{Comment, "# Keyword identifiers\n"},
	{Let, "let"},
	{Function, "fn"},
	{True, "true"},
	{False, "false"},
	{If, "if"},
	{Else, "else"},
	{Return, "return"},

	{Comment, "# Comma, braces and colons\n"},
	{Comma, ","},
	{Colon, ":"},
	{Semicolon, ";"},
	{LeftBrack, "["},
	{LeftParen, "("},
	{LeftBrace, "{"},
	{RightBrack, "]"},
	{RightParen, ")"},
	{RightBrace, "}"},
}

var source = func() (s string) {
	for _, t := range testTokens {
		s = s + t.lit + whitespace
	}
	return
}()

func countNewLines(s string) (n int) {
	for _, ch := range s {
		if ch == '\n' {
			n++
		}
	}
	return
}

func testToken(t *testing.T, got Token, want tok, l int) {
	if got.Type != want.typ {
		t.Errorf("bad token type, got: %q want: %q", got.Type, want.typ)
	}
	if got.Text != want.lit {
		t.Errorf("bad literal for %q: got %q, want: %q", want.typ, got.Text, want.lit)
	}
	if got.Line != l {
		t.Errorf("bad line count for %q: got %d, want %d", want.typ, got.Line, l)
	}
}

func TestScan(t *testing.T) {
	t.Logf("input\n%s", source)
	scanner := NewScanner(source)
	line := 1
	for _, tt := range testTokens {
		t.Run(tt.typ.String(), func(t *testing.T) {
			testToken(t, scanner.Next(), tt, line)
		})
		// source joins tokenList with whitespace which has 2 new lines
		line += countNewLines(tt.lit + whitespace)
	}
	testToken(t, Token{}, tok{EOF, ""}, 0)
}

// TestScanString tests some string scanning edge cases.
func TestScanString(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`"hello world!`},
		{`"hello world\"`},
		{`"hi theer\`},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s := NewScanner(tc.input)
			tk := s.Next()
			testToken(t, tk, tok{Error, "unterminated quoted string"}, 1)
		})
	}
}
