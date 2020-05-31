package main

import (
	"bufio"
	"io"
	"os"
)

// readInstructions creates a hack assembly reader. The reader emits
// instructions or an error. The resulting channel will be closed once
// reading is done. Note the value of the error will be set after the
// channel is close.
func readInstructions(stop <-chan os.Signal, r io.Reader) (<-chan *instruction, error) {
	ch := make(chan *instruction, 4) // create a bufferred channel of 4 instructions.
	var cherr error
	go func() {
		scanner := bufio.NewScanner(r)
		defer func() {
			close(ch) // finished processing
			if err := scanner.Err(); err != nil && cherr == nil {
				cherr = err
			}
		}()
		var addr uint16
		var line uint
		for scanner.Scan() {
			inst := newInstruction(scanner.Text(), line, addr)
			line++
			if inst == nil {
				continue
			}
			select {
			case <-stop:
				return // stop processing if we recieve a signal
			case ch <- inst:
				// only advance if we have an
				// instruction and not whitespace and
				// the instruction is not a label.
				if inst.typ != Label {
					addr++
				}
			}
		}
	}()
	return ch, cherr
}
