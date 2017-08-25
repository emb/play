// Package scan provides a scanner that tokenises the Monkey language.
package scan

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eof = -1

// stateFn represents the state of the scanner as a function that
// returns the next state.
type stateFn func(*Scanner) stateFn

// Scanner scans input for language tokens.
type Scanner struct {
	input  string     // the string being scanned
	tokens chan Token // channel for scanned tokens.

	line  int // current line in input
	pos   int // current position
	start int // start position of the current token
	width int // width of the last read rune
}

// NewScanner creates a scanner.
func NewScanner(input string) *Scanner {
	s := &Scanner{
		input:  input,
		tokens: make(chan Token, 2),
		line:   1,
	}
	go s.run()
	return s
}

// Next returns the next token.
func (s *Scanner) Next() Token {
	return <-s.tokens
}

// next return the next rune in the input
func (s *Scanner) next() rune {
	// terminate
	if s.pos >= len(s.input) {
		s.width = 0
		return eof
	}
	// advance
	r, w := utf8.DecodeRuneInString(s.input[s.pos:])
	s.width = w
	s.pos += s.width
	return r
}

// ignore skips oven input to the current point.
func (s *Scanner) ignore() { s.start = s.pos }

// backup steps back one rune. Should only be called once per next()
// call. The reason being is we use the width of a rune to backup and
// we never store more than the current runes' width.
func (s *Scanner) backup() { s.pos -= s.width }

// errorf emmits an error token and continues scanning.
func (s *Scanner) errorf(format string, args ...interface{}) stateFn {
	s.tokens <- Token{Error, s.line, fmt.Sprintf(format, args...)}
	return scanAny
}

// peek return the next rune without advancing the current position.
func (s *Scanner) peek() rune {
	r := s.next()
	s.backup()
	return r
}

// accept consumes the next rune if it is within a valid set.
func (s *Scanner) accept(valid string) bool {
	if strings.IndexRune(valid, s.next()) >= 0 {
		return true
	}
	s.backup() // otherwise we haven't accepted anything.
	return false
}

func (s *Scanner) emmit(t Type) {
	s.tokens <- Token{t, s.line, s.input[s.start:s.pos]}
	s.start = s.pos
}

// run scans the input by executing state functions until the state is
// nil.
func (s *Scanner) run() {
	for state := scanAny; state != nil; {
		state = state(s)
	}
	close(s.tokens) // no more tokens
}

// scanAny initial state function. Attempt at scanning tokens.
func scanAny(s *Scanner) stateFn {
	switch r := s.next(); {
	case r == eof:
		return nil
	case r == '\n':
		s.line++
		s.ignore()
		return scanAny(s)
	case r == '#':
		return scanComment(s)
	case isSpace(r):
		return scanSpace(s)
	case r == '"':
		return scanString(s)
	case isAlphaNumeric(r):
		return scanIdentifier(s)
	case r == ':':
		s.emmit(Colon)
		return scanAny(s)
	case r == ';':
		s.emmit(Semicolon)
		return scanAny(s)
	case r == ',':
		s.emmit(Comma)
		return scanAny(s)
	case r == '[':
		s.emmit(LeftBrack)
		return scanAny(s)
	case r == '(':
		s.emmit(LeftParen)
		return scanAny(s)
	case r == '{':
		s.emmit(LeftBrace)
		return scanAny(s)
	case r == ']':
		s.emmit(RightBrack)
		return scanAny(s)
	case r == ')':
		s.emmit(RightParen)
		return scanAny(s)
	case r == '}':
		s.emmit(RightBrace)
		return scanAny(s)
	default:
		return s.errorf("bad character at line %d: %#U", s.line, r)
	}
}

// scanSpace scan a run of spaces.
func scanSpace(s *Scanner) stateFn {
	for isSpace(s.peek()) {
		s.next()
	}
	s.ignore()
	return scanAny
}

// scanComment scans a comment.
func scanComment(s *Scanner) stateFn {
	for {
		r := s.next()
		if r == '\n' || r == eof {
			break
		}
	}
	s.emmit(Comment)
	// each comment starts with # and end with a newline increment
	// the line after emitting a comment to ensure that the
	// comment token has the correct line number.
	s.line++
	return scanSpace
}

// scanIdentifier scan an alphanumeric identifier.
func scanIdentifier(s *Scanner) stateFn {
	for {
		r := s.next()
		if !isAlphaNumeric(r) {
			s.backup()
			break
		}
	}
	if all(s.input[s.start:s.pos], unicode.IsDigit) {
		// all characters are digits
		s.emmit(Integer)
	} else {
		// could be a keyword or an identifier
		s.emmit(Lookup(s.input[s.start:s.pos]))
	}
	return scanAny
}

// scanString scans a quoted string, assumes a double quote being
// consumed.
func scanString(s *Scanner) stateFn {
Loop:
	for {
		switch s.next() {
		case '\\': // an escaped character.
			if r := s.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return s.errorf("unterminated quoted string")
		case '"': // end of the string
			break Loop
		}
	}
	s.emmit(String)
	return scanAny
}

// isSpace returns true if r is a space character
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isAlphaNumeric returns true if the identifier is a valid
// alphanumeric unicode character or `_` character.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsDigit(r) || unicode.IsLetter(r)
}

// all returns true if applying f to all elements of xs returns true.
//
// useful for something like all(xs, unicode.IsLetter)
func all(xs string, f func(rune) bool) bool {
	for _, x := range xs {
		if !f(x) {
			return false
		}
	}
	return true
}
