package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"

	"hackjackc/token"
)

// Lexer provides an easy access type to tokenize a Jack source file.
type Lexer struct {
	buf *bufio.Reader
}

// NewLexer constructs a lexer
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{buf: bufio.NewReader(r)}
}

// Next returns the next token. Returns token.EOF when no more tokens
// are available.
func (l *Lexer) Next() (tok token.Token) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				// Ensure we set the error ignoring
				// EOF which will return zero value
				if !errors.Is(v, io.EOF) {
					tok = token.Token{
						Type:    token.Illegal,
						Literal: fmt.Sprintf("%v", v),
					}
				}
			default:
				tok = token.Token{
					Type:    token.Illegal,
					Literal: fmt.Sprintf("%v", v),
				}
			}
		}
	}()
	tok = l.advance()
	return
}

// peek returns the next rune preserving where we are in the buffer.
func (l *Lexer) peek() rune {
	r, _, err := l.buf.ReadRune()
	if errors.Is(err, io.EOF) {
		return 0
	} else if err != nil {
		panic(err)
	}
	l.unread()
	return r
}

// read returns the next rune
func (l *Lexer) read() rune {
	r, _, err := l.buf.ReadRune()
	if err != nil {
		panic(err)
	}
	return r
}

// unread goes back a rune in the buffer.
func (l *Lexer) unread() {
	if err := l.buf.UnreadRune(); err != nil {
		panic(err)
	}
}

// consume reads a string until the occurrence of the first delimiter.
func (l *Lexer) consume(delim byte) string {
	s, err := l.buf.ReadString(delim)
	if err != nil && len(s) > 0 {
		return s
	} else if err != nil {
		panic(err)
	}
	return s[:len(s)-1]
}

// advance returns the next token
func (l *Lexer) advance() token.Token {
	switch r := l.read(); {
	case unicode.IsSpace(r):
		return l.advance()
	case r == '{':
		return token.Token{Type: token.LeftBrace, Literal: string(r)}
	case r == '}':
		return token.Token{Type: token.RightBrace, Literal: string(r)}
	case r == '(':
		return token.Token{Type: token.LeftParen, Literal: string(r)}
	case r == ')':
		return token.Token{Type: token.RightParen, Literal: string(r)}
	case r == '[':
		return token.Token{Type: token.LeftBraket, Literal: string(r)}
	case r == ']':
		return token.Token{Type: token.RightBraket, Literal: string(r)}
	case r == '.':
		return token.Token{Type: token.Dot, Literal: string(r)}
	case r == ',':
		return token.Token{Type: token.Comma, Literal: string(r)}
	case r == ';':
		return token.Token{Type: token.SemiColon, Literal: string(r)}
	case r == '+':
		return token.Token{Type: token.Plus, Literal: string(r)}
	case r == '-':
		return token.Token{Type: token.Minus, Literal: string(r)}
	case r == '*':
		return token.Token{Type: token.Multiply, Literal: string(r)}
	case r == '/':
		// deal with comments.
		if l.peek() == '/' {
			l.consume('\n') // ignore result
			return l.advance()
		}
		if l.peek() == '*' {
			for {
				l.consume('*')
				if r := l.read(); r == '/' {
					return l.advance()
				}
			}
		}
		return token.Token{Type: token.Divide, Literal: string(r)}
	case r == '&':
		return token.Token{Type: token.Ampersand, Literal: string(r)}
	case r == '|':
		return token.Token{Type: token.Pipe, Literal: string(r)}
	case r == '>':
		return token.Token{Type: token.GreaterThan, Literal: string(r)}
	case r == '<':
		return token.Token{Type: token.LessThan, Literal: string(r)}
	case r == '=':
		return token.Token{Type: token.Equal, Literal: string(r)}
	case r == '~':
		return token.Token{Type: token.Tilde, Literal: string(r)}
	case r == '"':
		s := l.consume('"')
		return token.Token{Type: token.StringConstant, Literal: s}
	case unicode.IsNumber(r):
		var buf bytes.Buffer
		buf.WriteRune(r)
		for {
			if unicode.IsNumber(l.peek()) {
				buf.WriteRune(l.read())
				continue
			}
			break
		}
		i, err := strconv.Atoi(buf.String())
		if err != nil {
			return token.Token{Type: token.Illegal, Literal: err.Error()}
		}
		return token.Token{Type: token.IntegerConstant, Literal: strconv.Itoa(i)}
	default:
		var buf bytes.Buffer
		buf.WriteRune(r)
		for {
			p := l.peek()
			if unicode.IsDigit(p) || unicode.IsLetter(p) || p == '_' {
				buf.WriteRune(l.read())
				continue
			}
			break
		}
		return token.KeywordOrIdentifier(buf.String())
	}
	return token.Token{}
}
