// Package lexer is an implementation for a Lexer for the Monkey
// language.
package lexer

import (
	"github.com/emb/play/monkey/token"
)

// isLetter returns true if the underlying byte is an ASCII letter or
// underscore
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit returns true if the underlying byte is a number.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

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

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// skip skips white space
func (l *Lexer) skip() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// ident reads an identifier
func (l *Lexer) ident() string {
	pos := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *Lexer) number() string {
	pos := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *Lexer) string() string {
	pos := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[pos:l.position]
}

// NextToken returns a next token every time it is called on a given
// input. When tokens run out token.EOF is returned.
func (l *Lexer) NextToken() token.Token {
	new := func(t token.Type, ch byte) token.Token {
		return token.Token{Type: t, Literal: string(ch)}
	}

	l.skip()

	var tok token.Token
	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			eq := string(l.ch)
			l.readChar()
			eq += string(l.ch)
			tok.Type = token.EQ
			tok.Literal = eq
		} else {
			tok = new(token.ASSIGN, l.ch)
		}
	case '+':
		tok = new(token.PLUS, l.ch)
	case '-':
		tok = new(token.MINUS, l.ch)
	case '*':
		tok = new(token.ASTERISK, l.ch)
	case '/':
		tok = new(token.SLASH, l.ch)
	case '>':
		tok = new(token.GT, l.ch)
	case '<':
		tok = new(token.LT, l.ch)
	case '!':
		if l.peekChar() == '=' {
			neq := string(l.ch)
			l.readChar()
			neq += string(l.ch)
			tok.Type = token.NEQ
			tok.Literal = neq
		} else {
			tok = new(token.BANG, l.ch)
		}
	case ':':
		tok = new(token.COLON, l.ch)
	case ';':
		tok = new(token.SEMICOLON, l.ch)
	case ',':
		tok = new(token.COMMA, l.ch)
	case '(':
		tok = new(token.LPAREN, l.ch)
	case ')':
		tok = new(token.RPAREN, l.ch)
	case '{':
		tok = new(token.LBRACE, l.ch)
	case '}':
		tok = new(token.RBRACE, l.ch)
	case '[':
		tok = new(token.LBRACKET, l.ch)
	case ']':
		tok = new(token.RBRACKET, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.string()
	case 0:
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			i := l.ident()
			return token.Token{
				Type:    token.LookupIdent(i),
				Literal: i,
			}
		} else if isDigit(l.ch) {
			return token.Token{
				Type:    token.INT,
				Literal: l.number(),
			}
		}
		tok = new(token.ILLEGAL, l.ch)
	}
	l.readChar()
	return tok
}
