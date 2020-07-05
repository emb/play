package main

import (
	"hackjackc/token"
)

// Modeling the AST to contain tokens for now so it is easy to print an XML

type Program struct {
	Class  token.Token
	Name   Identifier
	LeftB  token.Token
	Vars   []ClassVarDecleration   `xml:"classVarDec"`
	Subs   []SubroutineDecleration `xml:"subroutineDec"`
	RightB token.Token
}

type ClassVarDecleration struct {
	DecType token.Token // static or field
	VarType token.Token // int, char, boolean or Class Name
	Names   []Identifier
	Semi    token.Token
}

type SubroutineDecleration struct {
	SubType token.Token // method, constructor or function
	RetType token.Token // type or void
	Name    Identifier
	LeftP   token.Token
	Params  Parameters `xml:"parameterList"`
	RightP  token.Token
	SubBody SubroutineBody `xml:"subroutineBody"`
}

type Parameters struct {
	Tokens []token.Token
}

type SubroutineBody struct {
	LeftB      token.Token
	VarDec     []VarDeclaration `xml:"varDec"`
	Statements Statements       `xml:"statements"`
	RightB     token.Token
}

type VarDeclaration struct {
	Var     token.Token
	VarType token.Token
	Names   []Identifier
	Semi    token.Token
}

type Statements struct {
	Statements []interface{}
}

type LetStatement struct {
	XMLName       struct{} `xml:"letStatement"`
	Let           token.Token
	Name          Identifier
	LeftB         *token.Token
	ArrExpression *Expression `xml:"arrexpression"`
	RightB        *token.Token
	Equal         token.Token
	Expression    Expression `xml:"expression"`
	Semi          token.Token
}

type IfStatement struct {
	XMLName        struct{} `xml:"ifStatement"`
	If             token.Token
	LeftP          token.Token
	Expression     Expression `xml:"expression"`
	RightP         token.Token
	LeftB          token.Token
	Statements     Statements `xml:"statements"`
	RightB         token.Token
	Else           *token.Token
	ElseLeftB      *token.Token
	ElseStatements *Statements `xml:"elsestatements"`
	ElseRightB     *token.Token
}

type WhileStatement struct {
	XMLName    struct{} `xml:"whileStatement"`
	While      token.Token
	LeftP      token.Token
	Expression Expression `xml:"expression"`
	RightP     token.Token
	LeftB      token.Token
	Statements Statements `xml:"statements"`
	RightB     token.Token
}

type DoStatement struct {
	XMLName struct{} `xml:"doStatement"`
	Do      token.Token
	SubroutineCall
	Semi token.Token
}

type SubroutineCall struct {
	Dest        *Identifier
	Dot         *token.Token
	Name        Identifier
	LeftP       token.Token
	Expressions *ExpressionList `xml:",omitempty"`
	RightP      token.Token
}

type ReturnStatement struct {
	XMLName    struct{} `xml:"returnStatement"`
	Return     token.Token
	Expression *Expression `xml:"expression"`
	Semi       token.Token
}

type ExpressionList struct {
	XMLName     struct{}      `xml:"expressionList"`
	Expressions []interface{} `xml:"expression"` // []Expression (to add commas)
}

type Expression struct {
	Term1 Term `xml:"term"`
	Op    *token.Token
	Term2 *Term `xml:"term2"`
}

type Term struct {
	// IntegerConstant, StringConstant, KeywordConstant, Identifier
	Token *token.Token

	// Array
	ArrayIdent  *token.Token
	ArrayLeftB  *token.Token
	ArrayExpr   *Expression `xml:"arrexpression"`
	ArrayRightB *token.Token

	// Subroutine Call
	SubroutineCall `xml:",omitempty"`

	// Parenthesized
	ParenLeftP      *token.Token
	ParenExpression *Expression `xml:"expression"`
	ParenRightP     *token.Token

	// Unary Operation
	UnaryOp      *token.Token
	UnaryOperand *Term `xml:"term"`
}

//type VarType struct {token.Token}
type Identifier struct{ token.Token }
