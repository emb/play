package evaluator

import (
	"testing"

	"github.com/emb/play/monkey/lexer"
	"github.com/emb/play/monkey/object"
	"github.com/emb/play/monkey/parser"
)

func TestEvalIntegerExpressions(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"5", 5},
		{"-4", -4},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	for i, tc := range tests {
		t.Logf("test[%d] input %q", i, tc.input)
		testIntObj(t, testEval(tc.input), tc.want)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"false", false},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"!5", false},
		{"!!5", true},
	}
	for i, tc := range tests {
		t.Logf("test[%d] input %q", i, tc.input)
		testBoolObj(t, testEval(tc.input), tc.want)
	}
}

func testEval(input string) object.Object {
	parse := parser.New(lexer.New(input))
	return Eval(parse.Program())
}

func testIntObj(t *testing.T, obj object.Object, want int64) {
	val, ok := obj.(*object.Int)
	if !ok {
		t.Fatalf("obj is of type %T, want *object.Int", obj)
	}
	if int64(*val) != want {
		t.Errorf("obj has a value %d, want %d", *val, want)
	}
}

func testBoolObj(t *testing.T, obj object.Object, want bool) {
	val, ok := obj.(*object.Bool)
	if !ok {
		t.Fatalf("ob is of type %T, want *object.Bool", obj)
	}
	if bool(*val) != want {
		t.Errorf("ob has a value %t, want %t", *val, want)
	}
}
