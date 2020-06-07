package main

import (
	"bufio"
	"io"
)

// readInstruction from r and build an instruction set in memory
func readInstructions(r io.Reader) ([]*instruction, error) {
	insts := []*instruction{}
	scanner := bufio.NewScanner(r)
	var addr uint16
	var line uint
	for scanner.Scan() {
		inst := newInstruction(scanner.Text(), line, addr)
		line++
		if inst == nil {
			continue
		}
		insts = append(insts, inst)
		// only advance if we have an instruction and not
		// whitespace and the instruction is not a label.
		if inst.typ != Label {
			addr++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return insts, nil
}
