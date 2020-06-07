package main

import (
	"strconv"
)

// Symbols stores hack computer symbols.
type Symbols struct {
	table map[string]uint16
}

// Add adds a symbol to the smybol table.
func (s *Symbols) Add(label string, addr uint16) {
	if s.table == nil {
		s.table = make(map[string]uint16)
	}
	s.table[label] = addr
}

// Get gets an addres from the symbol table.
func (s *Symbols) Get(label string) (addr uint16, ok bool) {
	if s == nil {
		return
	}
	addr, ok = s.table[label]
	return
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

// readSymbols reads A-Instructions/Labels from input slice and builds
// a symbol table.
func readSymbols(insts []*instruction) *Symbols {
	// Load predefined symbols
	var s Symbols
	for k, v := range predefinedSymbols {
		s.Add(k, v)
	}
	// process variables later
	vars := []string{}
	for _, inst := range insts {
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
		// Now this could be a variable or a label. Since we don't know,
		// store it in vars and process it later.
		vars = append(vars, inst.val)
	}
	// Process variables.
	var nextVar uint16 = 16 // Variables start at address 16
	for _, k := range vars {
		if _, ok := s.Get(k); !ok {
			s.Add(k, nextVar)
			nextVar++
		}
	}
	return &s
}
