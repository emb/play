package main

import (
	"strconv"
	"sync"
)

// Symbols stores hack computer symbols. Symbols is safe for
// concurrent use and can do partial reads if desired.
type Symbols struct {
	table map[string]uint16
	ready chan struct{}

	mux sync.RWMutex
}

// Add adds a symbol to the smybol table.
func (s *Symbols) Add(label string, addr uint16) {
	s.mux.Lock()
	if s.table == nil {
		s.table = make(map[string]uint16)
		s.ready = make(chan struct{})
	}
	s.table[label] = addr
	s.mux.Unlock()
}

// get gets an addres from the symbol table.
func (s *Symbols) get(label string) (addr uint16, ok bool) {
	if s == nil {
		return
	}
	s.mux.RLock()
	addr, ok = s.table[label]
	s.mux.RUnlock()
	return
}

// Complete signals that we have finished reading the table.
func (s *Symbols) Complete() {
	if s.ready == nil {
		return
	}
	close(s.ready)
}

// Get waits for the table to be complete before returning if
// a value does not exist.
func (s *Symbols) Get(label string) (uint16, bool) {
	addr, ok := s.get(label)
	if ok {
		return addr, ok
	}
	if s.ready != nil {
		<-s.ready // block
	}
	return s.get(label)
}

// predefinedSymbols are symbols predefined by hack computer
var predefinedSymbols = map[string]uint16{
	"SP":     0,
	"LCL":    1,
	"ARG":    2,
	"THIS":   3,
	"THAT":   4,
	"R0":     0,
	"R1":     1,
	"R2":     2,
	"R3":     3,
	"R4":     4,
	"R5":     5,
	"R6":     6,
	"R7":     7,
	"R8":     8,
	"R9":     9,
	"R10":    10,
	"R11":    11,
	"R12":    12,
	"R13":    13,
	"R14":    14,
	"R15":    15,
	"SCREEN": 16384,
	"KBD":    24576,
}

// readSymbols reads A-Instructions/Labels from input channel ch and
// builds a symbol table.
func readSymbols(ch <-chan *instruction) *Symbols {
	// Load predefined symbols
	var s Symbols
	for k, v := range predefinedSymbols {
		s.Add(k, v)
	}

	go func(ch <-chan *instruction, s*Symbols) {
		vars := map[string]struct{}{}
		for inst := range ch {
			// Don't process C-Instructions
			if inst.typ == CInstruction {
				continue
			}
			// If it is a label add it.
			if inst.typ == Label {
				s.Add(inst.val, inst.addr)
				continue
			}
			// Don't process A-Instruction that is not a variable/lable
			if _, err := strconv.Atoi(inst.val); err == nil {
				// AInstruction is a value containing a number.
				continue
			}
			// Ignore existing variables
			if _, ok := s.get(inst.val); ok {
				continue
			}
			// Now this could be a variable or a label. Since we don't know,
			// store it in vars and process it later.
			vars[inst.val] = struct{}{}
		}
		// Process variables.
		var nextVar uint16 = 16 // Variables start at address 16
		for k, _ := range vars {
			if _, ok := s.get(k); !ok {
				s.Add(k, nextVar)
				nextVar++
			}
		}
		s.Complete()
	}(ch, &s)

	return &s
}
