// Package repl provides REPL utilities for the monkey language.
package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/emb/play/monkey/lexer"
	"github.com/emb/play/monkey/token"
)

func prompt(w io.Writer) {
	fmt.Fprint(w, ">> ")
}

// Start starts the Read, Eval, Print, Loop.
func Start(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for prompt(out); scanner.Scan(); prompt(out) {
		line := scanner.Text()
		l := lexer.New(line)
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Fprintf(out, "%+v\n", tok)
		}
	}
	return scanner.Err()
}
