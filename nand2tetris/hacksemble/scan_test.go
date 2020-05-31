package main

import (
	"os"
	"strings"
	"testing"
)

func TestReadInstructions(t *testing.T) {
	prog := strings.NewReader(`// Simple program to add 2 numbers.
	@R0  // Memory[0]
	D=A
	@R1  // Memory[1]
	D=D+A
	@R2  // Memory[2]
	M=D
(END)
	@END
	0;JMP
`)
	ch, _ := readInstructions(nil, prog)
	insts := []*instruction{}
	for i := range ch {
		t.Logf("instruction: %s", i)
		insts = append(insts, i)
	}

	want := 9
	if len(insts) != want {
		t.Errorf("readInstructions: want %d insturctions got %d", want, len(insts))
	}

	// Check for various instruction types.
	tests := []struct {
		name string
		want instruction
		got  instruction
	}{
		{
			name: "inst0",
			want: instruction{typ: AInstruction, val: "R0", addr: 0, line: 1},
			got:  *insts[0],
		},
		{
			name: "inst3",
			want: instruction{typ: CInstruction, val: "D=D+A", addr: 3, line: 4},
			got:  *insts[3],
		},
		{
			name: "inst6",
			want: instruction{typ: Label, val: "END", addr: 6, line: 7},
			got:  *insts[6],
		},
		{
			name: "inst7",
			want: instruction{typ: AInstruction, val: "END", addr: 6, line: 8},
			got:  *insts[7],
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want != tc.got {
				t.Errorf("instruction: want %v, got %v", tc.want, tc.got)
			}
		})
	}
}

func TestReadInstructionStops(t *testing.T) {
	stop := make(chan os.Signal, 1)
	stop <- os.Kill
	r := strings.NewReader(`// some program
	@R0
	D=A
(END)
	@END
	0;JMP
`)
	ch, _ := readInstructions(stop, r)
	var count int
	for range ch {
		// drain
		count++
	}
	if count > 4 {
		// the channel has a buffer of 4 the 5 insturctions above should never finish.
		t.Errorf("did not stop processed %d instructions expecting less than 4", count)
	}
}
