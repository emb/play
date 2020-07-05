/*
hackjackc is the Jack language compiler. See chapter 10/11 of the
elements of computer systems.

Usage: hackjackc [Path/to/jack/files]
*/
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s PATH_TO_JACK_DIR_OR_FILE\n", os.Args[0])
		os.Exit(1)
	}
	status := 0
	for _, p := range os.Args[1:] {
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
	parser := NewParser(NewLexer(r))
	prog, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("compile: %w", err)
	}

	c := struct {
		Program
		XMLName struct{} `xml:"class"`
	}{Program: *prog}
	x, err := xml.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	// Hack to fix xml expression layout.
	x = bytes.ReplaceAll(x, []byte("arrexpression"), []byte("expression"))
	x = bytes.ReplaceAll(x, []byte("elsestatements"), []byte("statements"))
	x = bytes.ReplaceAll(x, []byte("term2"), []byte("term"))
	x = bytes.ReplaceAll(x, []byte("<expressionList></expressionList>"),
		[]byte("<expressionList>\n</expressionList>"))
	x = bytes.ReplaceAll(x, []byte("<parameterList></parameterList>"),
		[]byte("<parameterList>\n</parameterList>"))

	outpath := strings.TrimSuffix(f, filepath.Ext(f))
	out, err := os.Create(fmt.Sprintf("%s.xml", outpath))
	if err != nil {
		return fmt.Errorf("compile: %w", err)
	}
	if _, err := out.Write(x); err != nil {
		return fmt.Errorf("compile: %w", err)
	}
	return nil
}
