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
		{"7", 7},
	}
	for _, tc := range tests {
		evaled := testEval(tc.input)
		testIntegerObj(t, evaled, tc.want)
	}
}

func testEval(input string) object.Object {
	parse := parser.New(lexer.New(input))
	return Eval(parse.Program())
}

func testIntegerObj(t *testing.T, obj object.Object, want int64) {
	val, ok := obj.(*object.Int)
	if !ok {
		t.Errorf("obj is of type %T, want *object.Int", obj)
	}
	if int64(*val) != want {
		t.Errorf("obj has a value %d, want %d", *val, want)
	}
}
