package main

import (
	"fmt"
	"strings"
	"testing"

	"hackjackc/token"
)

func tkn(typ token.TokenType, l string) token.Token {
	return token.Token{Type: typ, Literal: l}
}

func TestLexer(t *testing.T) {
	tests := []struct {
		in   string
		want token.Token
	}{
		{"", tkn(token.EOF, "")},
		{" ", tkn(token.EOF, "")},
		{"{", tkn(token.LeftBrace, "{")},
		{"}", tkn(token.RightBrace, "}")},
		{"(", tkn(token.LeftParen, "(")},
		{")", tkn(token.RightParen, ")")},
		{"[", tkn(token.LeftBraket, "[")},
		{"]", tkn(token.RightBraket, "]")},
		{".", tkn(token.Dot, ".")},
		{",", tkn(token.Comma, ",")},
		{";", tkn(token.SemiColon, ";")},
		{"+", tkn(token.Plus, "+")},
		{"-", tkn(token.Minus, "-")},
		{"*", tkn(token.Multiply, "*")},
		{"// foo bar\n// x + y", tkn(token.EOF, "")},
		{"/** foo bar **/", tkn(token.EOF, "")},
		{"/", tkn(token.Divide, "/")},
		{"&", tkn(token.Ampersand, "&")},
		{"|", tkn(token.Pipe, "|")},
		{">", tkn(token.GreaterThan, ">")},
		{"<", tkn(token.LessThan, "<")},
		{"=", tkn(token.Equal, "=")},
		{"~", tkn(token.Tilde, "~")},
		{"456", tkn(token.IntegerConstant, "456")},
		{`"abc"`, tkn(token.StringConstant, "abc")},
		{"f1", tkn(token.Identifier, "f1")},
		{"foo", tkn(token.Identifier, "foo")},
		{"class", tkn(token.Class, "class")},
		{"constructor", tkn(token.Constructor, "constructor")},
		{"function", tkn(token.Function, "function")},
		{"method", tkn(token.Method, "method")},
		{"field", tkn(token.Field, "field")},
		{"static", tkn(token.Static, "static")},
		{"var", tkn(token.Var, "var")},
		{"int", tkn(token.Int, "int")},
		{"char", tkn(token.Char, "char")},
		{"boolean", tkn(token.Boolean, "boolean")},
		{"void", tkn(token.Void, "void")},
		{"true", tkn(token.True, "true")},
		{"false", tkn(token.False, "false")},
		{"null", tkn(token.Null, "null")},
		{"this", tkn(token.This, "this")},
		{"let", tkn(token.Let, "let")},
		{"do", tkn(token.Do, "do")},
		{"if", tkn(token.If, "if")},
		{"else", tkn(token.Else, "else")},
		{"while", tkn(token.While, "while")},
		{"return", tkn(token.Return, "return")},
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
		`Class("class")`,
		`Identifier("Foo")`,
		`LeftBrace("{")`,
		`Static("static")`,
		`Int("int")`,
		`Identifier("foo")`,
		`SemiColon(";")`,
		`Field("field")`,
		`Char("char")`,
		`Identifier("a")`,
		`Comma(",")`,
		`Identifier("b")`,
		`SemiColon(";")`,
		`Constructor("constructor")`,
		`LeftParen("(")`,
		`RightParen(")")`,
		`LeftBrace("{")`,
		`RightBrace("}")`,
		`Function("function")`,
		`LeftParen("(")`,
		`RightParen(")")`,
		`LeftBrace("{")`,
		`RightBrace("}")`,
		`Method("method")`,
		`LeftParen("(")`,
		`RightParen(")")`,
		`LeftBrace("{")`,
		`Let("let")`,
		`Identifier("a")`,
		`Equal("=")`,
		`Identifier("foo")`,
		`LeftBraket("[")`,
		`IntegerConstant("2")`,
		`RightBraket("]")`,
		`SemiColon(";")`,
		`Let("let")`,
		`Identifier("b")`,
		`Equal("=")`,
		`StringConstant("bar")`,
		`SemiColon(";")`,
		`If("if")`,
		`LeftParen("(")`,
		`Tilde("~")`,
		`Identifier("x")`,
		`RightParen(")")`,
		`LeftBrace("{")`,
		`Return("return")`,
		`SemiColon(";")`,
		`RightBrace("}")`,
		`Do("do")`,
		`Identifier("X")`,
		`Dot(".")`,
		`Identifier("Y")`,
		`LeftParen("(")`,
		`RightParen(")")`,
		`SemiColon(";")`,
		`Return("return")`,
		`SemiColon(";")`,
		`RightBrace("}")`,
		`RightBrace("}")`,
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
