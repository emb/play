// Package ast describes the Abstract Syntax Tree of the Monkey Language.
//
// The monkey language has 2 main statements; let and return. The rest
// of the language is made up of expressions. Even the if conditionals
// are expressions.
package ast

import (
	"bytes"
	"fmt"

	"github.com/emb/play/monkey/token"
)

// Node an interface describing a node in the AST.
type Node interface {
	fmt.Stringer
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

// String constructs a string representing the program
// statements. Useful for debugging and testing.
func (p *Program) String() string {
	if p == nil {
		return ""
	}
	var buf bytes.Buffer
	for _, s := range p.Statements {
		buf.WriteString(s.String())
	}
	return buf.String()
}

// LetStmt describes a Let statement.
type LetStmt struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

// TokenLiteral returns the token literal underlying let statement.
func (l *LetStmt) TokenLiteral() string { return l.Token.Literal }

// String reconstructs let statement into valid code.
func (l *LetStmt) String() string {
	if l == nil {
		return ""
	}
	return fmt.Sprintf("%s %s = %s;", l.TokenLiteral(), l.Name, l.Value)
}

// Identifier describes user defined names used during variable
// bindings in the language.
type Identifier struct {
	Token token.Token
	Value string
}

// TokenLiteral returns the literal value of an identifier token.
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// String returns the value of identifier
func (i *Identifier) String() string {
	if i == nil {
		return ""
	}
	return i.Value
}

// ReturnStmt describes a Return statement.
type ReturnStmt struct {
	Token token.Token
	Value Expression
}

// TokenLiteral returns the token literal underlying return statement.
func (r *ReturnStmt) TokenLiteral() string { return r.Token.Literal }

// String reconstructs return statement into valid code.
func (r *ReturnStmt) String() string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf(" %s %s;", r.TokenLiteral(), r.Value)
}

// ExpressionStmt describes an Expression statement. Unlike the main two
// statements of the language this is a wrapper. Since the following
// code is valid Monkey code.
// 	let x = 4
//	x + 3
type ExpressionStmt struct {
	// Token stores the first token of an expression.
	Token      token.Token
	Expression Expression
}

// TokenLiteral returns the literal value of the first token in an
// expression.
func (e *ExpressionStmt) TokenLiteral() string { return e.Token.Literal }

// String reconstruct an Expression statement into valid code.
func (e *ExpressionStmt) String() string {
	if e == nil {
		return ""
	}
	return e.Expression.String()
}
