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
	"os/signal"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		exitErr(fmt.Errorf("Usage: %s File.asm", os.Args[0]))
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, os.Kill)

	status := 0
	for _, f := range os.Args[1:] {
		err := assemble(stop, f)
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

func assemble(stop <-chan os.Signal, fname string) error {
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

	// Go routine to read instructions
	chSymbols := make(chan *instruction)
	chOutput := make(chan *instruction, 10)
	insts, cerr := readInstructions(stop, in)

	// fan out
	go func(in <-chan *instruction, sym, out chan *instruction) {
		for inst := range in {
			sym<-inst // Send here first
			out<-inst // then here
		}
		close(sym)
		close(out)
	}(insts, chSymbols, chOutput)
	

	// Gut the symbols table.
	s := readSymbols(chSymbols)
	
	for inst := range chOutput {
		if inst.typ == Label {
			continue
		}
		bits, err := inst.Encode(s.Get)
		if err != nil {
			return fmt.Errorf("assemble(%s): %w", fname, err)
		}
		fmt.Fprintf(out, "%016b\n", bits)
	}
	return cerr
}

func output(fname string) (io.WriteCloser, error) {
	outname := fmt.Sprintf("%s.hack", strings.TrimSuffix(fname, filepath.Ext(fname)))
	return os.Create(outname)
}
