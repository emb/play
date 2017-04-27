// Package token defines the language tokens.
package token

// Type abstracts the type of tokens. A string was chosen to
// simplify printing despite the performance implication.
type Type string

// Token a structure describing a token in the language.
type Token struct {
	Type    Type
	Literal string
}

// Unexpected tokens
const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF" // End of file reached
)

// Identifiers and literals
const (
	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STRING"
)

// Operations
const (
	ASSIGN = "="
	PLUS   = "+"
	MINUS  = "-"
	SLASH  = "/"

	EQ  = "=="
	NEQ = "!="
	GT  = ">"
	LT  = "<"

	BANG     = "!"
	ASTERISK = "*"
)

// Delimiters
const (
	COMMA     = ","
	SEMICOLON = ";"
)

// Parenthesis
const (
	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"
)

// Keywords
const (
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)

var keywords = map[string]Type{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

// LookupIdent returns the type of a given identifier whether it is a
// keyword or not.
func LookupIdent(k string) Type {
	if t, ok := keywords[k]; ok {
		return t
	}
	return IDENT
}
