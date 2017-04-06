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
	IDENT = "IDENT"
	INT   = "INT"
)

// Operations
const (
	ASSIGN = "="
	PLUS   = "+"
)

// Delimiters
const (
	COMMA     = ","
	SEMICOLON = ";"
)

// Parenthesis
const (
	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
)

// Keywords
const (
	FUNCTION = "FUNCTION"
	LET      = "LET"
)
