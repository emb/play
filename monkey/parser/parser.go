// Package parser defines the parsing utilities for an AST.
package parser

import (
	"fmt"
	"strconv"

	"github.com/emb/play/monkey/ast"
	"github.com/emb/play/monkey/lexer"
	"github.com/emb/play/monkey/token"
)

// New crates a new parser given a lexer l.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []error{},
		prefixParseFns: map[token.Type]prefixParseFn{},
		infixParseFns:  map[token.Type]infixParseFn{},
	}

	// Read two tokens, so both c,p are set.
	p.next()
	p.next()

	// Register expression parsing functions.
	p.registerPrefix(token.IDENT, p.ident)
	p.registerPrefix(token.INT, p.int)
	p.registerPrefix(token.STRING, p.str)
	p.registerPrefix(token.BANG, p.prefix)
	p.registerPrefix(token.MINUS, p.prefix)
	p.registerPrefix(token.TRUE, p.bool)
	p.registerPrefix(token.FALSE, p.bool)
	p.registerPrefix(token.LPAREN, p.grouped)
	p.registerPrefix(token.IF, p.ifexpr)
	p.registerPrefix(token.FUNCTION, p.fn)
	p.registerPrefix(token.LBRACKET, p.array)
	p.registerPrefix(token.LBRACE, p.hash)

	p.registerInfix(token.PLUS, p.infix)
	p.registerInfix(token.MINUS, p.infix)
	p.registerInfix(token.ASTERISK, p.infix)
	p.registerInfix(token.SLASH, p.infix)
	p.registerInfix(token.EQ, p.infix)
	p.registerInfix(token.NEQ, p.infix)
	p.registerInfix(token.GT, p.infix)
	p.registerInfix(token.LT, p.infix)
	p.registerInfix(token.LPAREN, p.call)
	p.registerInfix(token.LBRACKET, p.index)

	return p
}

// expression parsing functions for infix/prefix parsing.
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser parses tokens into the monkey AST
type Parser struct {
	l *lexer.Lexer

	c token.Token // Current
	p token.Token // Next/Peek token

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn

	errors []error
}

// registerPrefix registers a prefix parsing function
func (p *Parser) registerPrefix(t token.Type, fn prefixParseFn) {
	p.prefixParseFns[t] = fn
}

// registerInfix register an infix parsing function
func (p *Parser) registerInfix(t token.Type, fn infixParseFn) {
	p.infixParseFns[t] = fn
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

// Errors returns a list of parser errors
func (p *Parser) Errors() []error { return p.errors }

func (p *Parser) statement() ast.Statement {
	switch p.c.Type {
	case token.LET:
		return p.letStmt()
	case token.RETURN:
		return p.retStmt()
	default:
		return p.exprStmt()
	}
}

func (p *Parser) letStmt() *ast.LetStmt {
	stmt := &ast.LetStmt{Token: p.c}
	if !p.nextIfPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.c, Value: p.c.Literal}
	if !p.nextIfPeek(token.ASSIGN) {
		return nil
	}
	p.next()
	stmt.Value = p.expr(Lowest)
	if p.peekIs(token.SEMICOLON) {
		p.next()
	}
	return stmt
}

func (p *Parser) retStmt() *ast.ReturnStmt {
	stmt := &ast.ReturnStmt{Token: p.c}
	p.next()
	stmt.Value = p.expr(Lowest)
	if p.peekIs(token.SEMICOLON) {
		p.next()
	}
	return stmt
}

func (p *Parser) exprStmt() *ast.ExpressionStmt {
	stmt := &ast.ExpressionStmt{Token: p.c}
	stmt.Expression = p.expr(Lowest)
	if p.peekIs(token.SEMICOLON) {
		p.next()
	}
	return stmt
}

func (p *Parser) expr(prec precedence) ast.Expression {
	prefix := p.prefixParseFns[p.c.Type]
	if prefix == nil {
		p.err("missing parse function for token %s", p.c.Type)
		return nil
	}
	left := prefix()
	for !p.peekIs(token.SEMICOLON) && prec < p.pp() {
		infix := p.infixParseFns[p.p.Type]
		if infix == nil {
			return left
		}
		p.next()
		left = infix(left)
	}
	return left
}

func (p *Parser) ident() ast.Expression {
	return &ast.Identifier{Token: p.c, Value: p.c.Literal}
}

func (p *Parser) int() ast.Expression {
	i, err := strconv.ParseInt(p.c.Literal, 10, 64)
	if err != nil {
		p.err("error parsing an integer %q: %s", p.c.Literal, err)
		return nil
	}
	return &ast.IntegerLiteral{Token: p.c, Value: i}
}

func (p *Parser) str() ast.Expression {
	return &ast.StringLiteral{Token: p.c, Value: p.c.Literal}
}

func (p *Parser) prefix() ast.Expression {
	expr := &ast.PrefixExpr{
		Token:    p.c,
		Operator: p.c.Literal,
	}
	p.next()
	expr.Right = p.expr(Prefix)
	return expr
}

func (p *Parser) infix(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpr{
		Token:    p.c,
		Left:     left,
		Operator: p.c.Literal,
	}
	prec := p.cp()
	p.next()
	expr.Right = p.expr(prec)
	return expr
}

func (p *Parser) bool() ast.Expression {
	return &ast.Boolean{Token: p.c, Value: p.currentIs(token.TRUE)}
}

func (p *Parser) grouped() ast.Expression {
	p.next()
	expr := p.expr(Lowest)
	if !p.nextIfPeek(token.RPAREN) {
		return nil
	}
	return expr
}

func (p *Parser) ifexpr() ast.Expression {
	expr := &ast.IfExpr{Token: p.c}
	if !p.nextIfPeek(token.LPAREN) {
		return nil
	}
	p.next()
	expr.Condition = p.expr(Lowest)
	if !p.nextIfPeek(token.RPAREN) {
		return nil
	}
	if !p.nextIfPeek(token.LBRACE) {
		return nil
	}
	expr.Consequence = p.block()
	if p.peekIs(token.ELSE) {
		p.next()
		if !p.nextIfPeek(token.LBRACE) {
			return nil
		}
		expr.Alternative = p.block()
	}
	return expr
}

func (p *Parser) fn() ast.Expression {
	expr := &ast.FunctionLiteral{
		Token:      p.c,
		Parameters: []*ast.Identifier{},
	}
	if !p.nextIfPeek(token.LPAREN) {
		return nil
	}
	expr.Parameters = p.params()
	if !p.nextIfPeek(token.LBRACE) {
		return nil
	}
	expr.Body = p.block()
	return expr
}

func (p *Parser) block() *ast.BlockStmt {
	block := &ast.BlockStmt{Token: p.c, Statements: []ast.Statement{}}
	p.next()
	for !p.currentIs(token.RBRACE) && !p.currentIs(token.EOF) {
		stmt := p.statement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.next()
	}
	return block
}

func (p *Parser) params() []*ast.Identifier {
	idents := []*ast.Identifier{}
	if p.peekIs(token.RPAREN) {
		p.next()
		return idents
	}
	p.next()
	idents = append(idents, &ast.Identifier{Token: p.c, Value: p.c.Literal})
	for p.peekIs(token.COMMA) {
		p.next() // move to comma
		p.next() // move to next identifier
		idents = append(idents, &ast.Identifier{
			Token: p.c,
			Value: p.c.Literal,
		})
	}
	if !p.nextIfPeek(token.RPAREN) {
		return nil
	}
	return idents
}

func (p *Parser) call(callable ast.Expression) ast.Expression {
	return &ast.CallExpr{
		Token:     p.c,
		Function:  callable,
		Arguments: p.listExprs(token.RPAREN),
	}
}

func (p *Parser) listExprs(end token.Type) []ast.Expression {
	args := []ast.Expression{}
	if p.peekIs(end) { // Empty arguments
		p.next()
		return args
	}
	p.next()
	args = append(args, p.expr(Lowest))
	for p.peekIs(token.COMMA) {
		p.next() // advance to comma
		p.next() // prepare to parse next expression
		args = append(args, p.expr(Lowest))
	}
	if !p.nextIfPeek(end) {
		return nil
	}
	return args
}

func (p *Parser) array() ast.Expression {
	return &ast.ArrayLiteral{
		Token:    p.c,
		Elements: p.listExprs(token.RBRACKET),
	}
}

func (p *Parser) index(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpr{Token: p.c, Left: left}
	p.next()
	exp.Index = p.expr(Lowest)
	if !p.nextIfPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) hash() ast.Expression {
	hash := &ast.HashLiteral{
		Token: p.c,
		Pairs: make(map[ast.Expression]ast.Expression),
	}
	for !p.peekIs(token.RBRACE) {
		p.next()
		key := p.expr(Lowest)
		if !p.nextIfPeek(token.COLON) {
			return nil
		}
		p.next()
		value := p.expr(Lowest)
		hash.Pairs[key] = value
		if !p.peekIs(token.RBRACE) && !p.nextIfPeek(token.COMMA) {
			return nil
		}
	}
	if !p.nextIfPeek(token.RBRACE) {
		return nil
	}
	return hash
}

// nextIfPeek checks if the next/peek token type matches t then call
// next.
func (p *Parser) nextIfPeek(t token.Type) bool {
	if p.peekIs(t) {
		p.next()
		return true
	}
	p.err("expected next token to be %s, got %s instead", t, p.p.Type)
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

// error append an error to the list of errors in the parser.
func (p *Parser) err(msg string, a ...interface{}) {
	p.errors = append(p.errors, fmt.Errorf(msg, a))
}

// cp returns the current token precedence
func (p *Parser) cp() precedence {
	if prec, ok := precedences[p.c.Type]; ok {
		return prec
	}
	return Lowest
}

// pp returns the peek token precedence
func (p *Parser) pp() precedence {
	if prec, ok := precedences[p.p.Type]; ok {
		return prec
	}
	return Lowest
}

// precedence is going to be used define a list of operator precedence
// used when parsing expressions
type precedence int

// List operator precedence
const (
	Lowest      precedence = iota
	Equals                 // ==
	LessGreater            // > or <
	Sum                    // +
	Product                // *
	Prefix                 // -X or !X
	Call                   // myFunction(x)
	Index
)

var precedences = map[token.Type]precedence{
	token.EQ:       Equals,
	token.NEQ:      Equals,
	token.LT:       LessGreater,
	token.GT:       LessGreater,
	token.PLUS:     Sum,
	token.MINUS:    Sum,
	token.SLASH:    Product,
	token.ASTERISK: Product,
	token.LPAREN:   Call,
	token.LBRACKET: Index,
}
