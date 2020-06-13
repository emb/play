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

func process(fname string) error {
	if filepath.Ext(fname) != ".vm" {
		return fmt.Errorf("process %q: can only process .vm files", fname)
	}

	// Input
	in, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer in.Close()

	// Output
	name := strings.TrimSuffix(fname, filepath.Ext(fname))
	out, err := os.Create(fmt.Sprintf("%s.asm", name))
	if err != nil {
		return err
	}
	defer out.Close()

	p := NewParser(in)
	err = Translate(filepath.Base(name), p.Parse(), out)
	// Translate errors
	if err != nil {
		return fmt.Errorf("process %s: %s", fname, err)
	}
	// Parse errors
	if err := p.Err(); err != nil {
		return fmt.Errorf("process %s: %s", fname, err)
	}
	return nil
}
