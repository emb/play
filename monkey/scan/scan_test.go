package scan

import "testing"

type tok struct {
	typ Type
	lit string
}

const whitespace = " \t \n\n"

var testTokens = [...]tok{
	{Comment, "# General identifiers\n"},
	{Identifier, "some_x"},

	{Comment, "# Keyword identifiers\n"},
	{Let, "let"},
	{Function, "fn"},
	{True, "true"},
	{False, "false"},
	{If, "if"},
	{Else, "else"},
	{Return, "return"},
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
