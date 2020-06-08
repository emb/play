package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	ptr := func(i int) *int { return &i }
	tests := []struct {
		line    string
		want    *Command
		wantErr string
	}{
		{"foo bar", nil, `parse: unknown command "foo"`},
		// Arithmetic
		{"Add // some comment", &Command{kind: CmdArithmetic, arg: "add"}, ""},
		{"Add 2 3", nil, "parse: arithmetic commands should not have parts in line \"Add 2 3\""},
		{"sub", &Command{kind: CmdArithmetic, arg: "sub"}, ""},
		{"neg", &Command{kind: CmdArithmetic, arg: "neg"}, ""},
		{"eq", &Command{kind: CmdArithmetic, arg: "eq"}, ""},
		{"gt", &Command{kind: CmdArithmetic, arg: "gt"}, ""},
		{"lt", &Command{kind: CmdArithmetic, arg: "lt"}, ""},
		{"and", &Command{kind: CmdArithmetic, arg: "and"}, ""},
		{"or", &Command{kind: CmdArithmetic, arg: "or"}, ""},
		{"not", &Command{kind: CmdArithmetic, arg: "not"}, ""},
		// Memory Access
		{"push // foo", nil, "parse: memory access commands should have 3 parts got 1"},
		{"pop foo bar // foo", nil, `parse: memory access command index parsing: strconv.Atoi: parsing "bar": invalid syntax`},
		{"push foo 3", nil, `parse: memory access command segment "foo" is invalid`},
		{"push local 3", &Command{kind: CmdPush, arg: "local", param: ptr(3)}, ""},
		{"pop temp 8", &Command{kind: CmdPop, arg: "temp", param: ptr(8)}, ""},
	}
	for tn, tc := range tests {
		t.Run(fmt.Sprintf("Test%02d", tn), func(t *testing.T) {
			c, err := parse(tc.line)
			if !reflect.DeepEqual(tc.want, c) {
				t.Errorf("parse want %v, got %v", tc.want, c)
			}
			if err != nil && err.Error() != tc.wantErr {
				t.Errorf("pares err want %q got %q", tc.wantErr, err)
			}
			if err == nil && tc.wantErr != "" {
				t.Errorf("parse err want %q got %q", tc.wantErr, err)
			}
		})
	}
}
