package main


// Modeling the AST to contain tokens for now so it is easy to print an XML

type Program struct {
	Class  Token
	Name   Identifier
	LeftB  Token
	Vars   []ClassVarDecleration   `xml:"classVarDec"`
	Subs   []SubroutineDecleration `xml:"subroutineDec"`
	RightB Token
}

type ClassVarDecleration struct {
	DecType Token // static or field
	VarType Token // int, char, boolean or Class Name
	Names   []Identifier
	Semi    Token
}

type SubroutineDecleration struct {
	SubType Token // method, constructor or function
	RetType Token // type or void
	Name    Identifier
	LeftP   Token
	Params  Parameters `xml:"parameterList"`
	RightP  Token
	SubBody SubroutineBody `xml:"subroutineBody"`
}

type Parameters struct {
	Tokens []Token
}

type SubroutineBody struct {
	LeftB      Token
	VarDec     []VarDeclaration `xml:"varDec"`
	Statements Statements       `xml:"statements"`
	RightB     Token
}

type VarDeclaration struct {
	Var     Token
	VarType Token
	Names   []Identifier
	Semi    Token
}

type Statements struct {
	Statements []interface{}
}

type LetStatement struct {
	XMLName       struct{} `xml:"letStatement"`
	Let           Token
	Name          Identifier
	LeftB         *Token
	ArrExpression *Expression `xml:"arrexpression"`
	RightB        *Token
	Equal         Token
	Expression    Expression `xml:"expression"`
	Semi          Token
}

type IfStatement struct {
	XMLName        struct{} `xml:"ifStatement"`
	If             Token
	LeftP          Token
	Expression     Expression `xml:"expression"`
	RightP         Token
	LeftB          Token
	Statements     Statements `xml:"statements"`
	RightB         Token
	Else           *Token
	ElseLeftB      *Token
	ElseStatements *Statements `xml:"elsestatements"`
	ElseRightB     *Token
}

type WhileStatement struct {
	XMLName    struct{} `xml:"whileStatement"`
	While      Token
	LeftP      Token
	Expression Expression `xml:"expression"`
	RightP     Token
	LeftB      Token
	Statements Statements `xml:"statements"`
	RightB     Token
}

type DoStatement struct {
	XMLName struct{} `xml:"doStatement"`
	Do      Token
	SubroutineCall
	Semi Token
}

type SubroutineCall struct {
	Dest        *Identifier
	Dot         *Token
	Name        Identifier
	LeftP       Token
	Expressions *ExpressionList `xml:",omitempty"`
	RightP      Token
}

type ReturnStatement struct {
	XMLName    struct{} `xml:"returnStatement"`
	Return     Token
	Expression *Expression `xml:"expression"`
	Semi       Token
}

type ExpressionList struct {
	XMLName     struct{}      `xml:"expressionList"`
	Expressions []interface{} `xml:"expression"` // []Expression (to add commas)
}

type Expression struct {
	Term1 Term `xml:"term"`
	Op    *Token
	Term2 *Term `xml:"term2"`
}

type Term struct {
	// IntegerConstant, StringConstant, KeywordConstant, Identifier
	Token *Token

	// Array
	ArrayIdent  *Token
	ArrayLeftB  *Token
	ArrayExpr   *Expression `xml:"arrexpression"`
	ArrayRightB *Token

	// Subroutine Call
	SubroutineCall `xml:",omitempty"`

	// Parenthesized
	ParenLeftP      *Token
	ParenExpression *Expression `xml:"expression"`
	ParenRightP     *Token

	// Unary Operation
	UnaryOp      *Token
	UnaryOperand *Term `xml:"term"`
}

//type VarType struct {Token}
type Identifier struct{ Token }
