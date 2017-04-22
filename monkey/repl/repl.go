// Package repl provides REPL utilities for the monkey language.
package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/emb/play/monkey/evaluator"
	"github.com/emb/play/monkey/lexer"
	"github.com/emb/play/monkey/parser"
)

func prompt(w io.Writer) {
	fmt.Fprint(w, ">> ")
}

// Start starts the Read, Eval, Print, Loop.
func Start(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for prompt(out); scanner.Scan(); prompt(out) {
		line := scanner.Text()
		parse := parser.New(lexer.New(line))
		program := parse.Program()
		if errs := parse.Errors(); len(errs) != 0 {
			displayErrors(out, errs)
			continue
		}
		fmt.Fprintf(out, "eval(%s) ->\n", program)
		result, err := evaluator.Eval(program)
		if err != nil {
			displayErrors(out, []error{err})
		} else if result != nil {
			fmt.Fprintf(out, "%s\n", result.Inspect())
		}
	}
	return scanner.Err()
}

const monkey = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func displayErrors(out io.Writer, errs []error) {
	io.WriteString(out, monkey)
	fmt.Fprint(out, "Woops! We ran into some monkey business here!\n")
	fmt.Fprint(out, "   parser errors:\n")
	for _, err := range errs {
		fmt.Fprintf(out, "\t* %s\n", err)
	}
}
