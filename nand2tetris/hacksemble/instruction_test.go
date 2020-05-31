package main

import (
	"fmt"
	"testing"
)

func TestEncodeDest(t *testing.T) {
	tests := []struct {
		v    string
		want uint16
	}{
		{"", 0},
		{"M", 1 << 3},
		{"D", 2 << 3},
		{"MD", 3 << 3},
		{"DM", 3 << 3},
		{"A", 4 << 3},
		{"AM", 5 << 3},
		{"AD", 6 << 3},
		{"AMD", 7 << 3},
		{"MAD", 7 << 3},
	}
	for _, tc := range tests {
		t.Run(tc.v, func(t *testing.T) {
			bits, err := encodeDest(tc.v)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if bits != tc.want {
				t.Errorf("encodeDest(%s) want 0b%b got 0b%b", tc.v, tc.want, bits)
			}
		})
	}
	t.Run("InvalidDestination", func(t *testing.T) {
		_, err := encodeDest("FooB")
		msg := `unknown destination 'F' in "FooB"`
		if err.Error() != msg {
			t.Errorf("error mismatch want %q got %q", msg, err)
		}
	})
}

func TestEncodeComp(t *testing.T) {
	tests := map[string][]struct {
		v    string
		want uint16
	}{
		"A": {
			{"0", 0b101010_000000},
			{"1", 0b111111_000000},
			{"-1", 0b111010_000000},
			{"D", 0b001100_000000},
			{"A", 0b110000_000000},
			{"!D", 0b001101_000000},
			{"!A", 0b110001_000000},
			{"-D", 0b001111_000000},
			{"-A", 0b110011_000000},
			{"D+1", 0b011111_000000},
			// {"1+D", 0b011111_000000}, // not sure if this is valid
			{"A+1", 0b110111_000000},
			{"D-1", 0b001110_000000},
			{"A-1", 0b110010_000000},
			{"D+A", 0b000010_000000},
			// {"A+D", 0b000010_000000}, // should we make this vaild?
			{"D-A", 0b010011_000000},
			{"A-D", 0b000111_000000},
			{"D&A", 0b000000_000000},
			{"D|A", 0b010101_000000},
		},
		"M": {
			{"M", 0b1_110000_000000},
			{"!M", 0b1_110001_000000},
			{"-M", 0b1_110011_000000},
			{"M+1", 0b1_110111_000000},
			{"M-1", 0b1_110010_000000},
			{"D+M", 0b1_000010_000000},
			{"D-M", 0b1_010011_000000},
			{"M-D", 0b1_000111_000000},
			{"D&M", 0b1_000000_000000},
			{"D|M", 0b1_010101_000000},
		},
	}
	for class, subtests := range tests {
		for _, tc := range subtests {
			t.Run(fmt.Sprintf("%s/%s", class, tc.v), func(t *testing.T) {
				bits, err := encodeComp([]byte(tc.v))
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if bits != tc.want {
					t.Errorf("encodeComp(%s) want 0b%b got 0b%b", tc.v, tc.want, bits)
				}
			})
		}
	}
}

func TestEncodeJump(t *testing.T) {
	tests := []struct {
		v    string
		want uint16
	}{
		{"", 0},
		{"JGT", 1},
		{"JEQ", 2},
		{"JGE", 3},
		{"JLT", 4},
		{"JNE", 5},
		{"JLE", 6},
		{"JMP", 7},
	}
	for _, tc := range tests {
		t.Run(tc.v, func(t *testing.T) {
			bits, err := encodeJump(tc.v)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if bits != tc.want {
				t.Errorf("encodeJump(%s) want 0b%b got 0b%b", tc.v, tc.want, bits)
			}
		})
	}
	t.Run("InvalidJump", func(t *testing.T) {
		_, err := encodeJump("gyfUp")
		msg := `unknown jump mnemonic "gyfUp"`
		if err.Error() != msg {
			t.Errorf("error mismatch want %q got %q", msg, err)
		}
	})
}

func TestInstructionEncode(t *testing.T) {
	getSymbol := func(l string) (uint16, bool) {
		return 16, true
	}

	tests := []struct {
		name string
		inst *instruction
		want uint16
		err  string
	}{
		// A-Insturction
		{"Simple", newInstruction("@7298", 0, 0), 7298, ""},
		{"Symbol", newInstruction("@temp", 0, 0), 16, ""},
		{"LargeAddress", newInstruction("@65536", 0, 0), 32767,
			`encode instruction([A] line(0) 65536): strconv.ParseUint: parsing "65536": value out of range`},

		// C-Instruction
		{"SimpleA", newInstruction("D =  D + A ; JEQ", 0, 0), 0b1110000010010010, ""},
		{"SimpleM", newInstruction("D =  D + M ; JEQ", 0, 0), 0b1111000010010010, ""},
		{"NoEquals", newInstruction("D-A", 0, 0), 0b1110010011000000, ""},
		{"WithJump", newInstruction("D|A; JGE", 0, 0), 0b1110010101000011, ""},
		{"MulitDest", newInstruction("MD=D&A;JLT", 0, 0), 0b1110000000011100, ""},
		{"InvalidJump", newInstruction("MD=D&A;JLM", 0, 0), 0,
			`encode instruction([C] line(0) MD=D&A;JLM): unknown jump mnemonic "JLM"`},
		{"InvalidDest", newInstruction("XM=D&A;JGE", 0, 0), 0,
			`encode instruction([C] line(0) XM=D&A;JGE): unknown destination 'X' in "XM"`},
		{"InvaildOp", newInstruction("A=D*A;JGE", 0, 0), 0,
			`encode instruction([C] line(0) A=D*A;JGE): invalid operotor '*' in mnemonic "D*A"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.inst.Encode(getSymbol)
			if b != tc.want {
				t.Errorf("Encode want 0b%b got 0b%b", tc.want, b)
			}
			if err != nil && err.Error() != tc.err {
				t.Errorf("Encode error want %q got %q", tc.err, err)
			}
			if err == nil && tc.err != "" {
				t.Errorf(`Encode error want %q got ""`, tc.err)
			}
		})
	}
}
