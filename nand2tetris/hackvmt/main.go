/*
hackvmt is a program that translates Hack virtual machine code into
Hack assembly code.

Usage: hackvmt Foo.vm

The command will produce a Foo.asm file that can be assembled and run
on the Hack machine.
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		exitErr(fmt.Errorf("Usage: %s File.asm", os.Args[0]))
	}

	status := 0
	for _, f := range os.Args[1:] {
		err := process(f)
		if err != nil {
			status = 1
			fmt.Fprintf(os.Stderr, "Error: %s", err)
		}
	}
	os.Exit(status)
}

func exitErr(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
}

func process(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("process path %q: %w", path, err)
	}

	files := []string{}
	var outpath string
	if info.IsDir() {
		matches, err := filepath.Glob(filepath.Join(path, "*.vm"))
		if err != nil {
			return fmt.Errorf("process: glob path: %w", err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("process %q no .vm files", path)
		}
		files = append(files, matches...)
		outpath = filepath.Join(path, filepath.Base(path))
	} else {
		files = append(files, path)
		outpath = strings.TrimSuffix(path, filepath.Ext(path))
	}

	// Output
	out, err := os.Create(fmt.Sprintf("%s.asm", outpath))
	if err != nil {
		return err
	}
	defer out.Close()

	// buffer so we don't have to process everything in memory.
	ch := make(chan *Command, 4)

	// Start a go routine that writes the asm file.
	errch := make(chan error)
	go func() {
		err := Translate(ch, out)
		if err != nil {
			errch <- err
		}
		close(errch)
	}()

	// Let the parser star working.
	err = parseFiles(files, ch)
	close(ch) // ensure to close the channel after we have finished parsing.
	if err != nil {
		// TODO: consider deferring os.Remove with the output file here.
		return err
	}

	// Now wait & check for translation errors.
	return <-errch
}

func parseFiles(files []string, output chan<- *Command) error {
	for _, f := range files {
		if filepath.Ext(f) != ".vm" {
			return fmt.Errorf("parse %q: can only process .vm files", f)
		}
		r, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		defer r.Close()
		namespace := filepath.Base(strings.TrimSuffix(f, filepath.Ext(f)))
		if err := Parse(namespace, r, output); err != nil {
			return err
		}
	}
	return nil
}
