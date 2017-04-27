// Package ast describes the Abstract Syntax Tree of the Monkey Language.
//
// The monkey language has 2 main statements; let and return. The rest
// of the language is made up of expressions. Even the if conditionals
// are expressions.
package ast

import (
	"bytes"
	"fmt"
	"strings"

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

// IntegerLiteral describes an integer in the Monkey language
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

// TokenLiteral returns the literal value of IntegerLiteral
func (i *IntegerLiteral) TokenLiteral() string { return i.Token.Literal }

// String returns a string representation of IntegerLiteral
func (i *IntegerLiteral) String() string { return i.Token.Literal }

// StringLiteral describes a string in the Monkey languae
type StringLiteral struct {
	Token token.Token
	Value string
}

// TokenLiteral returns the literal string value
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }

// String returns a string representation of StringLiteral
func (s *StringLiteral) String() string { return s.Token.Literal }

// ArrayLiteral describes an array within the language
type ArrayLiteral struct {
	// Token stores `[` token
	Token    token.Token
	Elements []Expression
}

// TokenLiteral return the literal toke `[`
func (a *ArrayLiteral) TokenLiteral() string { return a.Token.Literal }

// String returns a string representation of an array
func (a *ArrayLiteral) String() string {
	if a == nil {
		return ""
	}
	var buf bytes.Buffer
	es := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		es[i] = e.String()
	}
	buf.WriteByte('[')
	buf.WriteString(strings.Join(es, ", "))
	buf.WriteByte(']')
	return buf.String()
}

// HashLiteral describes hashes in monkey
type HashLiteral struct {
	// Token stores the `{`
	Token token.Token
	Pairs map[Expression]Expression
}

// TokenLiteral returns the `{` literal
func (h *HashLiteral) TokenLiteral() string { return h.Token.Literal }

// String returns a string representation of the hash
func (h *HashLiteral) String() string {
	if h == nil {
		return ""
	}
	var buf bytes.Buffer
	pairs := []string{}
	for k, v := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s:%s", k, v))
	}
	buf.WriteByte('{')
	buf.WriteString(strings.Join(pairs, ", "))
	buf.WriteByte('}')
	return buf.String()
}

// IndexExpr describes an expression of the form myarray[1]
type IndexExpr struct {
	// Token is the `[`
	Token token.Token
	Left  Expression
	Index Expression
}

// TokenLiteral return the literal token `[`
func (i *IndexExpr) TokenLiteral() string { return i.Token.Literal }

// String returns a string representation of an index expression
func (i *IndexExpr) String() string {
	return fmt.Sprintf("(%s[%s])", i.Left, i.Index)
}

// PrefixExpr describes a prefix expressions of form -5.
type PrefixExpr struct {
	// Token describes the prefix token; ! or -
	Token    token.Token
	Operator string
	Right    Expression
}

// TokenLiteral returns the literal value of the prefix expression
func (p *PrefixExpr) TokenLiteral() string { return p.Token.Literal }

// String reconstructs the input code of a prefix expression
func (p *PrefixExpr) String() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("(%s%s)", p.Operator, p.Right)
}

// InfixExpr describes expressions of the form 7 + 9
type InfixExpr struct {
	// Token describes the infix operator
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

// TokenLiteral returns the literal value of the infix operator
func (i *InfixExpr) TokenLiteral() string { return i.Token.Literal }

// String reconstructs the input code of an infix expression
func (i *InfixExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", i.Left, i.Operator, i.Right)
}

// Boolean describes binary expressions such as, true and false
type Boolean struct {
	Token token.Token
	Value bool
}

// TokenLiteral returns the expression literal
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }

// String construct the input code for boolean expressions
func (b *Boolean) String() string {
	if b == nil {
		return ""
	}
	return b.Token.Literal
}

// IfExpr describes if expressions. These are not statements they
// produce a value.
type IfExpr struct {
	// Token hold the `if` token.
	Token       token.Token
	Condition   Expression
	Consequence *BlockStmt
	Alternative *BlockStmt
}

// TokenLiteral returns a string representing the first token of an if
// expression.
func (i *IfExpr) TokenLiteral() string { return i.Token.Literal }

// String returns a string representing the if expression code.
func (i *IfExpr) String() string {
	var buf bytes.Buffer
	buf.WriteString("if ")
	buf.WriteString(i.Condition.String())
	buf.WriteByte(' ')
	buf.WriteString(i.Consequence.String())
	if i.Alternative != nil {
		buf.WriteString(" else ")
		buf.WriteString(i.Alternative.String())
	}
	return buf.String()
}

// FunctionLiteral describes functions in the monkey language.
type FunctionLiteral struct {
	// Token holds the `fn` string
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStmt
}

// TokenLiteral returns a string representing the fn token.
func (f *FunctionLiteral) TokenLiteral() string { return f.Token.Literal }

// String returns a string representing the function code
func (f *FunctionLiteral) String() string {
	if f == nil {
		return ""
	}
	var buf bytes.Buffer
	params := make([]string, len(f.Parameters))
	for i, p := range f.Parameters {
		params[i] = p.String()
	}
	buf.WriteString(f.TokenLiteral())
	buf.WriteByte('(')
	buf.WriteString(strings.Join(params, ", "))
	buf.WriteByte(')')
	buf.WriteString(f.Body.String())
	return buf.String()
}

// BlockStmt describes a list of statements that belongs to IfExpr and
// FnExpr.
type BlockStmt struct {
	// Token describes the opening brace `{` of the block statement.
	Token      token.Token
	Statements []Statement
}

// TokenLiteral returns a string representing the opening of a block
// statement.
func (b *BlockStmt) TokenLiteral() string { return b.Token.Literal }

// String returns a string representation of block statement code.
func (b *BlockStmt) String() string {
	if b == nil {
		return ""
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for _, s := range b.Statements {
		buf.WriteString(s.String())
	}
	buf.WriteByte('}')
	return buf.String()
}

// CallExpr describes a function call in the language
type CallExpr struct {
	// Token will hold Left Parenthesis ( in a call
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

// TokenLiteral returns the first parenthesis of a call expression.
func (c *CallExpr) TokenLiteral() string { return c.Token.Literal }

// String returns a string representing a function call
func (c *CallExpr) String() string {
	if c == nil {
		return ""
	}
	var buf bytes.Buffer
	args := make([]string, len(c.Arguments))
	for i, a := range c.Arguments {
		args[i] = a.String()
	}
	buf.WriteString(c.Function.String())
	buf.WriteByte('(')
	buf.WriteString(strings.Join(args, ", "))
	buf.WriteByte(')')
	return buf.String()
}
