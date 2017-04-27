package ast

import (
	"testing"

	"github.com/emb/play/monkey/token"
)

func TestString(t *testing.T) {
	myVar := &Identifier{
		Token: token.Token{
			Type:    token.IDENT,
			Literal: "myVar",
		},
		Value: "myVar",
	}
	program := &Program{
		Statements: []Statement{
			&LetStmt{
				Token: token.Token{
					Type:    token.LET,
					Literal: "let",
				},
				Name: myVar,
				Value: &Identifier{
					Token: token.Token{
						Type:    token.IDENT,
						Literal: "someVar",
					},
					Value: "someVar",
				},
			},
			&ReturnStmt{
				Token: token.Token{
					Type:    token.RETURN,
					Literal: "return",
				},
				Value: myVar,
			},
		},
	}

	want := `let myVar = someVar; return myVar;`
	if program.String() != want {
		t.Errorf("program.String() is %q, want %q", program, want)
	}
}
