/*
hacksemble is an assembler for the hack computer.

Usage: hacksemble Foo.asm

Which will produce a Foo.hack file with the binary instructions that
can be loaded into the Hardware simulator.

The assembler is a multipass assembler that builds a symbol table as
well as decode instructions.
*/
package main

import (
	"fmt"
	"io"
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
		err := assemble(f)
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

func assemble(fname string) error {
	if filepath.Ext(fname) != ".asm" {
		return fmt.Errorf("assemble %q: can only process .asm files", fname)
	}

	// Input
	in, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer in.Close()

	// Output
	out, err := output(fname)
	if err != nil {
		return err
	}
	defer out.Close()

	insts, err := readInstructions(in)
	if err != nil {
		return err
	}

	// build symbols first pass
	s := readSymbols(insts)

	// write output
	for _, inst := range insts {
		if inst.typ == Label {
			continue
		}
		bits, err := inst.Encode(s.Get)
		if err != nil {
			return fmt.Errorf("assemble(%s): %w", fname, err)
		}
		fmt.Fprintf(out, "%016b\n", bits)
	}
	return nil
}

func output(fname string) (io.WriteCloser, error) {
	outname := fmt.Sprintf("%s.hack", strings.TrimSuffix(fname, filepath.Ext(fname)))
	return os.Create(outname)
}
