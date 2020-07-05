package main

import (
	"fmt"
	"strings"
	"testing"

)

func tkn(typ TokenType, l string) Token {
	return Token{Type: typ, Literal: l}
}

func TestLexer(t *testing.T) {
	tests := []struct {
		in   string
		want Token
	}{
		{"", tkn(EOF, "")},
		{" ", tkn(EOF, "")},
		{"{", tkn(LeftBraceToken, "{")},
		{"}", tkn(RightBraceToken, "}")},
		{"(", tkn(LeftParenToken, "(")},
		{")", tkn(RightParenToken, ")")},
		{"[", tkn(LeftBraketToken, "[")},
		{"]", tkn(RightBraketToken, "]")},
		{".", tkn(DotToken, ".")},
		{",", tkn(CommaToken, ",")},
		{";", tkn(SemiColonToken, ";")},
		{"+", tkn(PlusToken, "+")},
		{"-", tkn(MinusToken, "-")},
		{"*", tkn(MultiplyToken, "*")},
		{"// foo bar\n// x + y", tkn(EOF, "")},
		{"/** foo bar **/", tkn(EOF, "")},
		{"/", tkn(DivideToken, "/")},
		{"&", tkn(AmpersandToken, "&")},
		{"|", tkn(PipeToken, "|")},
		{">", tkn(GreaterThanToken, ">")},
		{"<", tkn(LessThanToken, "<")},
		{"=", tkn(EqualToken, "=")},
		{"~", tkn(TildeToken, "~")},
		{"456", tkn(IntegerConstant, "456")},
		{`"abc"`, tkn(StringConstant, "abc")},
		{"f1", tkn(IdentifierToken, "f1")},
		{"foo", tkn(IdentifierToken, "foo")},
		{"class", tkn(ClassToken, "class")},
		{"constructor", tkn(ConstructorToken, "constructor")},
		{"function", tkn(FunctionToken, "function")},
		{"method", tkn(MethodToken, "method")},
		{"field", tkn(FieldToken, "field")},
		{"static", tkn(StaticToken, "static")},
		{"var", tkn(VarToken, "var")},
		{"int", tkn(IntToken, "int")},
		{"char", tkn(CharToken, "char")},
		{"boolean", tkn(BooleanToken, "boolean")},
		{"void", tkn(VoidToken, "void")},
		{"true", tkn(TrueToken, "true")},
		{"false", tkn(FalseToken, "false")},
		{"null", tkn(NullToken, "null")},
		{"this", tkn(ThisToken, "this")},
		{"let", tkn(LetToken, "let")},
		{"do", tkn(DoToken, "do")},
		{"if", tkn(IfToken, "if")},
		{"else", tkn(ElseToken, "else")},
		{"while", tkn(WhileToken, "while")},
		{"return", tkn(ReturnToken, "return")},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			l := NewLexer(strings.NewReader(tc.in))
			got := l.Next()
			if got != tc.want {
				t.Errorf("want %v got %v", tc.want, got)
			}
		})
	}
}

func TestLexingProgram(t *testing.T) {
	prog := strings.NewReader(`/**
Foo is a class
*/
class Foo {
    static int foo;
    field char a, b;

    constructor() {}
    function() {}
    method() {
        let a = foo[2];
        let b = "bar";
        if (~x) {
            return;
        }
        do X.Y();
        return;
    }
}
`)
	want := []string{
		`ClassToken("class")`,
		`IdentifierToken("Foo")`,
		`LeftBraceToken("{")`,
		`StaticToken("static")`,
		`IntToken("int")`,
		`IdentifierToken("foo")`,
		`SemiColonToken(";")`,
		`FieldToken("field")`,
		`CharToken("char")`,
		`IdentifierToken("a")`,
		`CommaToken(",")`,
		`IdentifierToken("b")`,
		`SemiColonToken(";")`,
		`ConstructorToken("constructor")`,
		`LeftParenToken("(")`,
		`RightParenToken(")")`,
		`LeftBraceToken("{")`,
		`RightBraceToken("}")`,
		`FunctionToken("function")`,
		`LeftParenToken("(")`,
		`RightParenToken(")")`,
		`LeftBraceToken("{")`,
		`RightBraceToken("}")`,
		`MethodToken("method")`,
		`LeftParenToken("(")`,
		`RightParenToken(")")`,
		`LeftBraceToken("{")`,
		`LetToken("let")`,
		`IdentifierToken("a")`,
		`EqualToken("=")`,
		`IdentifierToken("foo")`,
		`LeftBraketToken("[")`,
		`IntegerConstant("2")`,
		`RightBraketToken("]")`,
		`SemiColonToken(";")`,
		`LetToken("let")`,
		`IdentifierToken("b")`,
		`EqualToken("=")`,
		`StringConstant("bar")`,
		`SemiColonToken(";")`,
		`IfToken("if")`,
		`LeftParenToken("(")`,
		`TildeToken("~")`,
		`IdentifierToken("x")`,
		`RightParenToken(")")`,
		`LeftBraceToken("{")`,
		`ReturnToken("return")`,
		`SemiColonToken(";")`,
		`RightBraceToken("}")`,
		`DoToken("do")`,
		`IdentifierToken("X")`,
		`DotToken(".")`,
		`IdentifierToken("Y")`,
		`LeftParenToken("(")`,
		`RightParenToken(")")`,
		`SemiColonToken(";")`,
		`ReturnToken("return")`,
		`SemiColonToken(";")`,
		`RightBraceToken("}")`,
		`RightBraceToken("}")`,
	}
	l := NewLexer(prog)
	tok := l.Next()
	for _, w := range want {
		if tok.String() != w {
			t.Errorf("want: %s\tgot %s", w, tok.String())
		}
		tok = l.Next()
	}
}
