package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"

)

// Lexer provides an easy access type to tokenize a Jack source file.
type Lexer struct {
	buf *bufio.Reader
}

// NewLexer constructs a lexer
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{buf: bufio.NewReader(r)}
}

// Next returns the next token. Returns EOF when no more tokens
// are available.
func (l *Lexer) Next() (tok Token) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				// Ensure we set the error ignoring
				// EOF which will return zero value
				if v != io.EOF {
					tok = Token{
						Type:    IllegalToken,
						Literal: fmt.Sprintf("%v", v),
					}
				}
			default:
				tok = Token{
					Type:    IllegalToken,
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
	if err == io.EOF {
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
func (l *Lexer) advance() Token {
	switch r := l.read(); {
	case unicode.IsSpace(r):
		return l.advance()
	case r == '{':
		return Token{Type: LeftBraceToken, Literal: string(r)}
	case r == '}':
		return Token{Type: RightBraceToken, Literal: string(r)}
	case r == '(':
		return Token{Type: LeftParenToken, Literal: string(r)}
	case r == ')':
		return Token{Type: RightParenToken, Literal: string(r)}
	case r == '[':
		return Token{Type: LeftBraketToken, Literal: string(r)}
	case r == ']':
		return Token{Type: RightBraketToken, Literal: string(r)}
	case r == '.':
		return Token{Type: DotToken, Literal: string(r)}
	case r == ',':
		return Token{Type: CommaToken, Literal: string(r)}
	case r == ';':
		return Token{Type: SemiColonToken, Literal: string(r)}
	case r == '+':
		return Token{Type: PlusToken, Literal: string(r)}
	case r == '-':
		return Token{Type: MinusToken, Literal: string(r)}
	case r == '*':
		return Token{Type: MultiplyToken, Literal: string(r)}
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
		return Token{Type: DivideToken, Literal: string(r)}
	case r == '&':
		return Token{Type: AmpersandToken, Literal: string(r)}
	case r == '|':
		return Token{Type: PipeToken, Literal: string(r)}
	case r == '>':
		return Token{Type: GreaterThanToken, Literal: string(r)}
	case r == '<':
		return Token{Type: LessThanToken, Literal: string(r)}
	case r == '=':
		return Token{Type: EqualToken, Literal: string(r)}
	case r == '~':
		return Token{Type: TildeToken, Literal: string(r)}
	case r == '"':
		s := l.consume('"')
		return Token{Type: StringConstant, Literal: s}
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
			return Token{Type: IllegalToken, Literal: err.Error()}
		}
		return Token{Type: IntegerConstant, Literal: strconv.Itoa(i)}
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
		return KeywordOrIdentifier(buf.String())
	}
	return Token{}
}
