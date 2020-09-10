/*
hackjackc is the Jack language compiler. See chapter 10/11 of the
elements of computer systems.

Usage: hackjackc [Path/to/jack/files]
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"encoding/json"
)

var printAST = flag.Bool("ast", false, "print the AST as a JSON object")

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s PATH_TO_JACK_DIR_OR_FILE\n", os.Args[0])
		os.Exit(1)
	}
	status := 0
	for _, p := range flag.Args() {
		if err := process(p); err != nil {
			status++
			fmt.Fprintf(os.Stderr, "Error: %s", err)
		}
	}
	os.Exit(status)
}

func process(p string) error {
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("process path %q: %w", p, err)
	}
	files := []string{}
	if info.IsDir() {
		matches, err := filepath.Glob(filepath.Join(p, "*.jack"))
		if err != nil {
			return fmt.Errorf("process path %q: %w", p, err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("process path %q: no .jack files", p)
		}
		files = append(files, matches...)
	} else {
		files = append(files, p)
	}
	for _, f := range files {
		if err := compile(f); err != nil {
			return fmt.Errorf("process path %q: %w", p, err)
		}
	}
	return nil
}

func compile(f string) error {
	if filepath.Ext(f) != ".jack" {
		return fmt.Errorf("compile: can only compile .jack files")
	}
	r, err := os.Open(f)
	if err != nil {
		return fmt.Errorf("compile: %w", err)
	}
	defer r.Close()
	
	parser := NewParser(NewLexer(r))
	prog, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("compile: %w", err)
	}
	// print the ast to stdout.
	if *printAST {
		if err := json.NewEncoder(os.Stdout).Encode(prog); err != nil {
			return fmt.Errorf("compile: %w", err)
		}
	}
	outpath := strings.TrimSuffix(f, filepath.Ext(f))
	out, err := os.Create(fmt.Sprintf("%s.vm", outpath))
	if err != nil {
		return fmt.Errorf("compile: %w", err)
	}
	defer out.Close()
	
	ctx := Context{Name: prog.Name.Literal, sym: NewSymbols()}
	w := writer{Writer:out}
	return prog.Compile(&ctx, &w)
}
