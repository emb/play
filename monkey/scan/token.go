package scan

import "fmt"

// Type represents the underlying token type
type Type int

// Define the list of tokens.
const (
	// Special tokens
	EOF Type = iota // Start here. This is our zero value too.
	Error
	Comment

	literalbeg
	Identifier // Alphanumeric identifier
	String     // String representations
	Integer    // Integer representations
	literalend

	operatorbeg
	Assign    // =
	Add       // +
	Multiply  // *
	Substract // -
	Divide    // /

	Equal   // ==
	Greater // >
	Lesser  // <
	Not     // !

	NotEqual       // !=
	GreaterOrEqual // >=
	LesserOrEqual  // <=

	Comma     // ,
	Colon     // :
	Semicolon // ;

	LeftBrack  // [
	LeftParen  // (
	LeftBrace  // {
	RightBrack // ]
	RightParen // )
	RightBrace // }
	operatorend

	keywordbeg
	Function
	Let
	True
	False
	If
	Else
	Return
	keywordend
)

var tokens = [...]string{
	EOF:     "EOF",
	Error:   "ERROR",
	Comment: "COMMENT",

	Identifier: "IDENT",
	Integer:    "INT",
	String:     "STRING",

	Assign:    "=",
	Add:       "+",
	Multiply:  "*",
	Substract: "-",
	Divide:    "/",

	Equal:   "==",
	Greater: ">",
	Lesser:  "<",
	Not:     "!",

	NotEqual:       "!=",
	GreaterOrEqual: ">=",
	LesserOrEqual:  "<=",

	Comma:     ",",
	Colon:     ":",
	Semicolon: ";",

	LeftBrack:  "[",
	LeftParen:  "(",
	LeftBrace:  "{",
	RightBrack: "]",
	RightParen: ")",
	RightBrace: "}",

	Function: "fn",
	Let:      "let",
	True:     "true",
	False:    "false",
	If:       "if",
	Else:     "else",
	Return:   "return",
}

var keywords map[string]Type

func init() {
	keywords = make(map[string]Type)
	for i := keywordbeg + 1; i < keywordend; i++ {
		keywords[tokens[i]] = i
	}
}

// String returns the string corresponding to the token
// tok. Operators, keywords and delimiters will have the actual string
// representing the tokens. Otherwise a string representing the
// contsant name such as IDENT.
func (t Type) String() string {
	var s string
	if 0 <= t && t < Type(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = fmt.Sprintf("token(%d)", int(t))
	}
	return s
}

// IsLiteral returns true for tokens that are identifiers and basic
// types (string, literals).
func (t Type) IsLiteral() bool { return literalbeg < t && t < literalend }

// IsOperator returns true for tokens that are operators and
// delimiters.
func (t Type) IsOperator() bool { return operatorbeg < t && t < operatorend }

// IsKeyword returns true for tokens that language keywords.
func (t Type) IsKeyword() bool { return keywordbeg < t && t < keywordend }

// Lookup returns a keyword or IDENT for a string.
func Lookup(ident string) Type {
	if k, isKeyword := keywords[ident]; isKeyword {
		return k
	}
	return Identifier
}

// Token represents a token returned from the scanner.
type Token struct {
	Type Type   // The type of this token
	Line int    // The line number of which this token appears
	Text string // The literal string of this token
}

// String returns a string representation of the token.
func (tok Token) String() string {
	switch {
	case tok.Type == EOF:
		return tok.Type.String()
	case tok.Type == Error:
		return "error: " + tok.Text
	case len(tok.Text) > 10:
		return fmt.Sprintf("%s: %.10q...", tok.Type, tok.Text)
	default:
		return fmt.Sprintf("%s: %q", tok.Type, tok.Text)
	}
}
