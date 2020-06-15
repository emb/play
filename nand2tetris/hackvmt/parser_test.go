package main

import (
	"fmt"
	"reflect"
	"strings"
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
		{"Add // some comment", &Command{Type: CmdArithmetic, Arg: "add"}, ""},
		{"Add 2 3", nil, "parse: arithmetic commands should not have parts in line \"Add 2 3\""},
		{"sub", &Command{Type: CmdArithmetic, Arg: "sub"}, ""},
		{"neg", &Command{Type: CmdArithmetic, Arg: "neg"}, ""},
		{"eq", &Command{Type: CmdArithmetic, Arg: "eq"}, ""},
		{"gt", &Command{Type: CmdArithmetic, Arg: "gt"}, ""},
		{"lt", &Command{Type: CmdArithmetic, Arg: "lt"}, ""},
		{"and", &Command{Type: CmdArithmetic, Arg: "and"}, ""},
		{"or", &Command{Type: CmdArithmetic, Arg: "or"}, ""},
		{"not", &Command{Type: CmdArithmetic, Arg: "not"}, ""},
		// Memory Access
		{"push // foo", nil, "parse: memory access commands should have 3 parts got 1"},
		{"pop foo bar // foo", nil, `parse: memory access command index parsing: strconv.Atoi: parsing "bar": invalid syntax`},
		{"push foo 3", nil, `parse: memory access command segment "foo" is invalid`},
		{"push local 3", &Command{Type: CmdPush, Arg: "local", Param: ptr(3)}, ""},
		{"pop temp 8", &Command{Type: CmdPop, Arg: "temp", Param: ptr(8)}, ""},
		// Function calls
		{"function foo 2", &Command{Type: CmdFunction, Arg: "foo", Param: ptr(2)}, ""},
		{"function foo ab", nil, `parse: function invalid locals: strconv.Atoi: parsing "ab": invalid syntax`},
		{"call foo 3", &Command{Type: CmdCall, Arg: "foo", Param: ptr(3)}, ""},
		{"call foo cd", nil, `parse: call invalid arguments: strconv.Atoi: parsing "cd": invalid syntax`},
		{"return", &Command{Type: CmdReturn}, ""},
		// Program control
		{"label foo", &Command{Type: CmdLabel, Arg: "foo"}, ""},
		{"label 1foo", nil, `parse: invalid label "1foo" that starts with a digit`},
		{"label f?oo", nil, `parse: invalid ? in label "f?oo"`},
		{"goto foo", &Command{Type: CmdGoto, Arg: "foo"}, ""},
		{"if-goto foo", &Command{Type: CmdIfGoto, Arg: "foo"}, ""},
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

func TestParseScope(t *testing.T) {
	prog := strings.NewReader(`function foo 1
	push constant 1
	return
label bar
`)
	ch := make(chan *Command, 4) // 4 lines above
	err := Parse("Ruin", prog, ch)
	close(ch)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	got := []string{}
	for c := range ch {
		if c.Namespace != "Ruin" {
			t.Errorf("command: %s does not have expected namespace %q got %q", c, "Ruin", c.Namespace)
		}
		got = append(got, c.Scope)
	}

	want := []string{"global", "foo", "foo", "global"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parse: want scope %s but got %s", want, got)
	}
}
