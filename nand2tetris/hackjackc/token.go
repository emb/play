// Package token declares a bunch of useful constants that need a
// namespace hence this package.
package main

import (
	"encoding/xml"
	"fmt"
)

// Token represents a Jack token
type Token struct {
	Literal string
	Type    TokenType
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q)", t.Type, t.Literal)
}

// Return which type of Lexical Element this token belongs to.
func (t Token) Element() string {
	if t.Type == IdentifierToken {
		return "identifier"
	}
	if t.Type == StringConstant {
		return "stringConstant"
	}
	if t.Type == IntegerConstant {
		return "integerConstant"
	}
	if t.Type >= ClassToken && t.Type <= ReturnToken {
		return "keyword"
	}
	if t.Type >= LeftBraceToken && t.Type <= TildeToken {
		return "symbol"
	}
	return "unknown"
}

func (t Token) MarshalXML(enc *xml.Encoder, s xml.StartElement) error {
	if t.Type == EOF {
		return nil
	}
	s.Name.Local = t.Element()
	return enc.EncodeElement(t.Literal, s)
}

// TokenType an enumeration of all jack tokens.
type TokenType int

const (
	// Useful unexpected tokens.
	EOF TokenType = iota
	IllegalToken

	// Keywords
	ClassToken
	ConstructorToken
	FunctionToken
	MethodToken
	FieldToken
	StaticToken
	VarToken
	IntToken
	CharToken
	BooleanToken
	VoidToken
	TrueToken
	FalseToken
	NullToken
	ThisToken
	LetToken
	DoToken
	IfToken
	ElseToken
	WhileToken
	ReturnToken

	// Symbols
	LeftBraceToken
	RightBraceToken
	LeftParenToken
	RightParenToken
	LeftBraketToken
	RightBraketToken
	DotToken
	CommaToken
	SemiColonToken
	PlusToken
	MinusToken
	MultiplyToken
	DivideToken
	AmpersandToken
	PipeToken
	GreaterThanToken
	LessThanToken
	EqualToken
	TildeToken

	// Constants
	IntegerConstant
	StringConstant

	// Identifier
	IdentifierToken
)

var keywords = map[string]TokenType {
	"class": ClassToken,
		"constructor": ConstructorToken,
		"function": FunctionToken,
		"method": MethodToken,
		"field": FieldToken,
		"static": StaticToken,
		"var": VarToken,
		"int": IntToken,
		"char": CharToken,
		"boolean": BooleanToken,
		"void": VoidToken,
		"true": TrueToken,
		"false": FalseToken,
		"null": NullToken,
		"this": ThisToken,
		"let": LetToken,
		"do": DoToken,
		"if": IfToken,
		"else": ElseToken,
		"while": WhileToken,
		"return": ReturnToken,
}

// KeywordOrIdentifier returns a token based on the incoming string.
func KeywordOrIdentifier(k string) Token {
	t, ok := keywords[k]
	if ok {
		return Token{Type: t, Literal: k}
	}
	return Token{Type: IdentifierToken, Literal: k}
}
