// Package ast describes the Abstract Syntax Tree of the Monkey Language.
package ast

import "github.com/emb/play/monkey/token"

// Node an interface describing a node in the AST.
type Node interface {
	TokenLiteral() string
}

// Statement a type of nodes in an AST.
type Statement interface {
	Node
}

// Expression a type of nodes in an AST
type Expression interface {
	Node
}

// Program describes the root node of AST
type Program struct {
	Statements []Statement
}

// TokenLiteral returns the first statement token literal.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// Let describes a Let statement.
type Let struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

// TokenLiteral returns the token literal underlying let statement.
func (l *Let) TokenLiteral() string {
	return l.Token.Literal
}

// Identifier describes user defined names used during variable
// bindings in the language.
type Identifier struct {
	Token token.Token
	Value string
}

// TokenLiteral returns the literal value of an identifier token.
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

// Return describes a Return statement.
type Return struct {
	Token token.Token
	Value Expression
}

// TokenLiteral returns the token literal underlying return statement.
func (r *Return) TokenLiteral() string {
	return r.Token.Literal
}
