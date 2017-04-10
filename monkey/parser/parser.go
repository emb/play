// Package parser defines the parsing utilities for an AST.
package parser

import (
	"github.com/emb/play/monkey/ast"
	"github.com/emb/play/monkey/lexer"
	"github.com/emb/play/monkey/token"
)

// New crates a new parser given a lexer l.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	// Read two tokens, so both c,p are set.
	p.next()
	p.next()

	return p
}

// Parser parses tokens into the monkey AST
type Parser struct {
	l *lexer.Lexer

	c token.Token // Current
	p token.Token // Next/Peek token
}

// next advances the parser by a token.
func (p *Parser) next() {
	p.c = p.p
	p.p = p.l.NextToken()
}

// Program parses and returns an ast.Program which is the root of
// Monkey's AST.
func (p *Parser) Program() *ast.Program {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}
	for p.c.Type != token.EOF {
		stmt := p.statement()
		if stmt != nil {
			program.Statements = append(program.Statements,
				stmt)
		}
		p.next()
	}
	return program
}

func (p *Parser) statement() ast.Statement {
	switch p.c.Type {
	case token.LET:
		return p.let()
	default:
		return nil
	}
}

func (p *Parser) let() *ast.Let {
	stmt := &ast.Let{Token: p.c}
	if !p.nextIfPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.c, Value: p.c.Literal}
	if !p.nextIfPeek(token.ASSIGN) {
		return nil
	}
	// TODO: skipping expressions for now until a semicolon
	for !p.currentIs(token.SEMICOLON) {
		p.next()
	}
	return stmt
}

// nextIfPeek checks if the next/peek token type matches t then call
// next.
func (p *Parser) nextIfPeek(t token.Type) bool {
	if p.peekIs(t) {
		p.next()
		return true
	}
	return false
}

// currentIs checks the type of the current token.
func (p *Parser) currentIs(t token.Type) bool {
	return p.c.Type == t
}

// peekIs checks if the type of the next/peek token.
func (p *Parser) peekIs(t token.Type) bool {
	return p.p.Type == t
}
