// Package token declares a bunch of useful constants that need a
// namespace hence this package.
package token

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
	if t.Type == Identifier {
		return "identifier"
	}
	if t.Type == StringConstant {
		return "stringConstant"
	}
	if t.Type == IntegerConstant {
		return "integerConstant"
	}
	if t.Type >= Class && t.Type <= Return {
		return "keyword"
	}
	if t.Type >= LeftBrace && t.Type <= Tilde {
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
	Illegal

	// Keywords
	Class
	Constructor
	Function
	Method
	Field
	Static
	Var
	Int
	Char
	Boolean
	Void
	True
	False
	Null
	This
	Let
	Do
	If
	Else
	While
	Return

	// Symbols
	LeftBrace
	RightBrace
	LeftParen
	RightParen
	LeftBraket
	RightBraket
	Dot
	Comma
	SemiColon
	Plus
	Minus
	Multiply
	Divide
	Ampersand
	Pipe
	GreaterThan
	LessThan
	Equal
	Tilde

	// Constants
	IntegerConstant
	StringConstant

	// Identifier
	Identifier
)

var keywords = map[string]TokenType {
	"class": Class,
		"constructor": Constructor,
		"function": Function,
		"method": Method,
		"field": Field,
		"static": Static,
		"var": Var,
		"int": Int,
		"char": Char,
		"boolean": Boolean,
		"void": Void,
		"true": True,
		"false": False,
		"null": Null,
		"this": This,
		"let": Let,
		"do": Do,
		"if": If,
		"else": Else,
		"while": While,
		"return": Return,
}

// KeywordOrIdentifier returns a token based on the incoming string.
func KeywordOrIdentifier(k string) Token {
	t, ok := keywords[k]
	if ok {
		return Token{Type: t, Literal: k}
	}
	return Token{Type: Identifier, Literal: k}
}
