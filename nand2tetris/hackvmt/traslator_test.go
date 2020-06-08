package main

import (
	"testing"
)

func TestPushTranslation(t *testing.T) {
	n := 3
	cmd := &Command{kind: CmdPush, arg: "constant", param: &n}
	got, err := translate(cmd)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	want := `	@3
	D=A
	@SP
	A=M
	M=D
	@SP
	M=M+1
`
	if string(got) != want {
		t.Errorf("translate: want [%s] got [%s]", want, got)
	}
}

func TestArithmeticTranslation(t *testing.T) {
	cmd := &Command{kind: CmdArithmetic, arg: "add"}
	got, err := translate(cmd)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	want := `	@SP
	AM=M-1
	D=M
	@SP
	A=M-1
	M=D+M
`
	if string(got) != want {
		t.Errorf("translate: want [%s] got [%s]", want, got)
	}
}
