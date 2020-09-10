package main


// Modeling the AST to contain tokens for now so it is easy to print an XML

type Program struct {
	Name   Identifier
	Vars   []ClassVarDecleration 
	Subs   []SubroutineDecleration
}

type ClassVarDecleration struct {
	DecType Token // static or field
	VarType Token // int, char, boolean or Class Name
	Names   []Identifier
}

type SubroutineDecleration struct {
	SubType Token // method, constructor or function
	RetType Token // type or void
	Name    Identifier
	Params  []Parameter
	SubBody SubroutineBody
}

type Parameter struct {
	Name Identifier
	Type Token
}

type SubroutineBody struct {
	VarDec     []VarDeclaration
	Statements Statements      
}

// NVars eturn the number of local variables
func (sb *SubroutineBody) NVars() int {
	if sb == nil {
		return 0
	}
	var count int
	for _, vd := range sb.VarDec {
		count += len(vd.Names)
	}
	return count
}

type VarDeclaration struct {
	VarType Token
	Names   []Identifier
}

type Compiler interface {
	Compile(*Context, *writer) error
}

type Statements struct {
	Statements []Compiler
}

type LetStatement struct {
	Name          Identifier
	ArrExpression *Expression
	Expression    Expression
}

type IfStatement struct {
	Expression     Expression
	Statements     Statements
	ElseStatements *Statements
}

type WhileStatement struct {
	Expression Expression
	Statements Statements
}

type DoStatement struct {
	SubCall SubroutineCall
}

type SubroutineCall struct {
	Dest        *Identifier
	Name        Identifier
	Expressions []Expression
}

type ReturnStatement struct {
	Expression *Expression
}

type Expression struct {
	Term1 Term
	Op    *Token
	Term2 *Term
}

type Term struct {
	// IntegerConstant, StringConstant, KeywordConstant, Identifier
	Token *Token

	// Array
	ArrayIdent  *Token
	ArrayExpr   *Expression

	// Subroutine Call
	SubCall *SubroutineCall

	// Parenthesized
	ParenExpression *Expression

	// Unary Operation
	UnaryOp      *Token
	UnaryOperand *Term 
}

type Identifier struct{ Token }
