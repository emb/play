package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Hack assembly language specifies an A-Instruction and
// C-Instruction. The label would be useful when building the symbol
// table.
type instructionType uint8

const (
	_ instructionType = iota
	AInstruction
	CInstruction
	Label
)

type errEnc struct {
	inst *instruction
	err  error
}

func (e *errEnc) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func (e *errEnc) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("encode instruction(%s): %s", e.inst, e.err)
}

type instruction struct {
	typ  instructionType
	addr uint16
	val  string

	line uint // store line number for debugging
}

func (i *instruction) String() string {
	var t string
	switch i.typ {
	case AInstruction:
		t = "A"
	case CInstruction:
		t = "C"
	case Label:
		t = "L"
	}
	return fmt.Sprintf("[%s] line(%d) %s", t, i.line, i.val)
}


// encodeA encodes AInstruction in hack computer. An A inscruction
// contains a positive integer value or a symbol.
func encodeA(v string, getSymbol func(string)(uint16, bool)) (uint16, error) {
	b, err := strconv.ParseUint(v, 10, 15)
	if err == nil {
		return uint16(b), nil
	} else if errors.Is(err, strconv.ErrRange) {
		return uint16(b), err
	}
	addr, ok := getSymbol(v)
	if !ok {
		return 0, errors.New("symbol not found")
	}
	return addr, nil
}

// destMap encodes destinations as uint16. See figure 4.4 in the book.
var destMap = map[rune]uint16{
	'M': 0b0000_0000_0000_1000,
	'D': 0b0000_0000_0001_0000,
	'A': 0b0000_0000_0010_0000,
}

func encodeDest(v string) (uint16, error) {
	if len(v) == 0 {
		return 0, nil
	}
	var d uint16
	for _, c := range v {
		bits, ok := destMap[c]
		if !ok {
			return 0, fmt.Errorf("unknown destination %q in %q", c, v)
		}
		d |= bits
	}
	return d, nil
}

// comps a helper type describing computation specfication.
type comps struct {
	op      byte
	operand [2]byte
}

// compsMaps maps computation speficications to bits. Note the M
// operand is ignored here. The calling function will have to deal
// with M separately.
var compsMap = map[comps]uint16{
	comps{0, [...]byte{'0', 0}}:     0b0000_101010_000000,
	comps{0, [...]byte{'1', 0}}:     0b0000_111111_000000,
	comps{'-', [...]byte{'1', 0}}:   0b0000_111010_000000,
	comps{0, [...]byte{'D', 0}}:     0b0000_001100_000000,
	comps{0, [...]byte{'A', 0}}:     0b0000_110000_000000,
	comps{'!', [...]byte{'D', 0}}:   0b0000_001101_000000,
	comps{'!', [...]byte{'A', 0}}:   0b0000_110001_000000,
	comps{'-', [...]byte{'D', 0}}:   0b0000_001111_000000,
	comps{'-', [...]byte{'A', 0}}:   0b0000_110011_000000,
	comps{'+', [...]byte{'D', '1'}}: 0b0000_011111_000000,
	comps{'+', [...]byte{'A', '1'}}: 0b0000_110111_000000,
	comps{'-', [...]byte{'D', '1'}}: 0b0000_001110_000000,
	comps{'-', [...]byte{'A', '1'}}: 0b0000_110010_000000,
	comps{'+', [...]byte{'D', 'A'}}: 0b0000_000010_000000,
	comps{'-', [...]byte{'D', 'A'}}: 0b0000_010011_000000,
	comps{'-', [...]byte{'A', 'D'}}: 0b0000_000111_000000,
	comps{'&', [...]byte{'D', 'A'}}: 0b0000_000000_000000,
	comps{'|', [...]byte{'D', 'A'}}: 0b0000_010101_000000,
}

// operands is a a set of valid operands
var operands = map[byte]struct{}{
	0: struct{}{}, // zero operand is valid.

	'0': struct{}{},
	'1': struct{}{},

	'A': struct{}{},
	'D': struct{}{},
	'M': struct{}{},
}

var operators = map[byte]struct{}{
	0: struct{}{}, // zero operator is valid.

	'-': struct{}{},
	'+': struct{}{},

	'&': struct{}{},
	'|': struct{}{},
	'!': struct{}{},
}

func encodeComp(mnemonic []byte) (uint16, error) {
	v := make([]byte, len(mnemonic))
	copy(v, mnemonic)
	if len(v) == 0 {
		return 0, errors.New("computation is required in a C-Instruction")
	}
	var a, op, b byte
	var isM bool
	if i := bytes.IndexByte(v, 'M'); i >= 0 {
		isM = true
		v[i] = 'A'
	}
	switch len(v) {
	case 1:
		a = v[0]
	case 2:
		op = v[0]
		a = v[1]
	case 3:
		a = v[0]
		op = v[1]
		b = v[2]
	default:
		return 0, fmt.Errorf("invalide mnemonic %q expecting an operator with maximum 2 operands",
			mnemonic)
	}
	// Validate before we continue.
	if _, ok := operators[op]; !ok {
		return 0, fmt.Errorf("invalid operotor %q in mnemonic %q", op, mnemonic)
	}
	if _, ok := operands[a]; !ok {
		return 0, fmt.Errorf("invalid operand %q in mnemonic %q", a, mnemonic)
	}
	if _, ok := operands[b]; !ok {
		return 0, fmt.Errorf("invalid operand %q in mnemonic %q", b, mnemonic)
	}

	bits, ok := compsMap[comps{op, [...]byte{a, b}}]
	if !ok {
		return 0, fmt.Errorf("computation not implementet for mnemonic %q", mnemonic)
	}

	var mprefix uint16 = 0
	if isM {
		mprefix = 0b0001_0000_0000_0000
	}
	return bits | mprefix, nil
}

// jumpMap encodes the jump bits see figure 4.5 in the book.
var jumpMap = map[string]uint16{
	"JGT": 0b0000_0000_0000_0001,
	"JEQ": 0b0000_0000_0000_0010,
	"JGE": 0b0000_0000_0000_0011,
	"JLT": 0b0000_0000_0000_0100,
	"JNE": 0b0000_0000_0000_0101,
	"JLE": 0b0000_0000_0000_0110,
	"JMP": 0b0000_0000_0000_0111,
}

func encodeJump(v string) (uint16, error) {
	if len(v) == 0 {
		return 0, nil
	}
	bits, ok := jumpMap[v]
	if !ok {
		return 0, fmt.Errorf("unknown jump mnemonic %q", v)
	}
	return bits, nil
}

// encodeC encodes a CInstruction according to hack computer
// specification. The instruction is in the form of
//
// 	Destination=Computation;Jump
//
// Where destination can be can be any of A/D registers or the special
// M memory register. Computation can be one of the ALU function and
// Jump is one of the jum mnemonics. Note: both = ; can be omitted.
func encodeC(v string) (uint16, error) {
	var dest, comp, jump bytes.Buffer
	iscomp := true // by default append to the comp mnemonic.
	for _, c := range v {
		if c == ' ' || c == '\t' { // I can't think of other whitespace in this code path.
			continue
		}
		if c == '=' {
			// whatever we have consumed so far should be moved to dest
			if _, err := comp.WriteTo(&dest); err != nil {
				return 0, err
			}
			comp.Reset()
			continue
		}
		if c == ';' {
			iscomp = false // next characters should be stored in jump.
			continue
		}
		if iscomp {
			if _, err := comp.WriteRune(c); err != nil {
				return 0, err
			}
			continue
		}
		if _, err := jump.WriteRune(c); err != nil {
			return 0, err
		}
	}
	d, err := encodeDest(dest.String())
	if err != nil {
		return 0, err
	}
	c, err := encodeComp(comp.Bytes())
	if err != nil {
		return 0, err
	}
	j, err := encodeJump(jump.String())
	if err != nil {
		return 0, err
	}
	var prefix uint16 = 0b_1110_0000_0000_0000 // see spec
	return prefix | d | c | j, nil
}

// Encode an instruction to its binary form based on hack computer.
func (i *instruction) Encode(getSymbol func(string) (uint16, bool)) (uint16, error) {
	var b uint16
	var err error
	switch i.typ {
	case AInstruction:
		b, err = encodeA(i.val, getSymbol)
	case CInstruction:
		b, err = encodeC(i.val)
	case Label:
		err = errors.New("labels cannot be encoded")
	default:
		err = errors.New("unknown instruction")
	}
	if err != nil {
		err = &errEnc{i, err}
	}
	return b, err
}

// extract trims the input line and ignores any comments.
func extract(line string) string {
	// find the index of the comment if any.
	if i := strings.Index(line, "//"); i >= 0 {
		return strings.TrimSpace(line[:i])
	}
	return strings.TrimSpace(line)
}

// newInstruction creates an instruction from an input
// line. Insturctions can be of any instructionType white space and
// comments will return a nil instruction without an error.
func newInstruction(line string, ln uint, addr uint16) *instruction {
	l := extract(line)
	switch {
	case l == "": // a comment line
		return nil
	case strings.HasPrefix(l, "@"):
		return &instruction{
			typ:  AInstruction,
			addr: addr,
			val:  l[1:], // remove prefix
			line: ln,
		}
	case strings.HasPrefix(l, "(") && strings.HasSuffix(l, ")"):
		return &instruction{
			typ:  Label,
			addr: addr,
			val:  l[1 : len(l)-1],
			line: ln,
		}
	default:
		return &instruction{
			typ:  CInstruction,
			addr: addr,
			val:  l,
			line: ln,
		}
	}
}
