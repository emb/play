// Package lexer is an implementation for a Lexer for the Monkey
// language.
package lexer

import (
	"github.com/emb/play/monkey/token"
)

// New creates a new lexer
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// Lexer a type used to tokenize the monkey language.
type Lexer struct {
	input        string
	position     int  // current position
	readPosition int  // reading position after current char
	ch           byte // char being examined
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken returns a next token every time it is called on a given
// input. When tokens run out token.EOF is returned.
func (l *Lexer) NextToken() token.Token {
	new := func(t token.Type, ch byte) token.Token {
		return token.Token{Type: t, Literal: string(ch)}
	}

	var tok token.Token
	switch l.ch {
	case '=':
		tok = new(token.ASSIGN, l.ch)
	case '+':
		tok = new(token.PLUS, l.ch)
	case ';':
		tok = new(token.SEMICOLON, l.ch)
	case '(':
		tok = new(token.LPAREN, l.ch)
	case ')':
		tok = new(token.RPAREN, l.ch)
	case ',':
		tok = new(token.COMMA, l.ch)
	case '{':
		tok = new(token.LBRACE, l.ch)
	case '}':
		tok = new(token.RBRACE, l.ch)
	case 0:
		tok.Type = token.EOF
	default:
		tok.Type = token.ILLEGAL
	}
	l.readChar()
	return tok
}
