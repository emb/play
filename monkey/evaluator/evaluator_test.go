package evaluator

import (
	"strconv"
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
		i, _ := testEval(tc.input)
		testIntObj(t, i, tc.want)
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
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}
	for i, tc := range tests {
		t.Logf("test[%d] input %q", i, tc.input)
		b, _ := testEval(tc.input)
		testBoolObj(t, b, tc.want)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (3 == 3) { true } else { false}", true},
		{"if (3 == 4) { true } else { false}", false},
	}
	for i, tc := range tests {
		t.Logf("test[%d] input %q", i, tc.input)
		result, _ := testEval(tc.input)
		switch w := tc.want.(type) {
		case int:
			testIntObj(t, result, int64(w))
		case bool:
			testBoolObj(t, result, w)
		default:
			testIsNull(t, result)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`
if (10 > 1) {
  if (10 > 1) {
    return 10;
  }

  return 1;
}
`,
			10,
		},
	}
	for i, tc := range tests {
		t.Logf("test[%d] input %q", i, tc.input)
		evaled, _ := testEval(tc.input)
		testIntObj(t, evaled, tc.want)
	}
}

func TestErrorHandling(t *testing.T) {
	typeErr := func(l object.Type, op string, r object.Type) error {
		return OpTypeMismatch{left: l, op: op, right: r}
	}
	prefixErr := func(op string, r object.Type) error {
		return BadPrefixOp{op: op, right: r}
	}
	infixErr := func(l object.Type, op string, r object.Type) error {
		return BadInfixOp{left: l, op: op, right: r}
	}
	tests := []struct {
		input string
		err   error
	}{
		{"5 + true;", typeErr(object.Integer, "+", object.Boolean)},
		{"5 + true; 5;", typeErr(object.Integer, "+", object.Boolean)},
		{"-true;", prefixErr("-", object.Boolean)},
		{
			"true + false;",
			infixErr(object.Boolean, "+", object.Boolean)},
		{
			"5; true + false; 4;",
			infixErr(object.Boolean, "+", object.Boolean),
		},
		{
			"if (10 > 1) { true + false; }",
			infixErr(object.Boolean, "+", object.Boolean),
		},
		{
			`if (10 > 1) {
  if (10 > 1) {
    return true + false;
  }

  return 1;
}
`,
			infixErr(object.Boolean, "+", object.Boolean),
		},
		{"foobar", UnboundIdent{ident: "foobar"}},
		{`"hi" - "ho"`, infixErr(object.String, "-", object.String)},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := testEval(tc.input)
			if err != tc.err {
				t.Errorf("error is %q, want %q", err, tc.err)
			}
		})
	}
}

func TestLetStatement(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"let a = 3; a;", 3},
		{"let b = 3 * 5; b;", 15},
		{"let c = 7; let d = c; d;", 7},
		{"let a = 3; let b = 4; let c = a + b + 5; c;", 12},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r, _ := testEval(tc.input)
			testIntObj(t, r, tc.want)
		})
	}
}

func TestFunctObject(t *testing.T) {
	input := "fn(x) { x +2; }"
	result, err := testEval(input)
	if err != nil {
		t.Fatalf("eval failed: %s", err)
	}

	fn, ok := result.(*object.Funct)
	if !ok {
		t.Fatalf("fn is of type %T, want *object.Funct", fn)
	}
	if len(fn.Parameters) != 1 {
		t.Errorf("fn has %d parameter, want 1", len(fn.Parameters))
	}
	if fn.Parameters[0].String() != "x" {
		t.Errorf("first parameter is %s, want x", fn.Parameters[0])
	}
	wantBody := "{(x + 2)}"
	if fn.Body.String() != wantBody {
		t.Errorf("fn.Body is %q, want %q", fn.Body, wantBody)
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"let identity = fn(x) { x; }; identity(4);", 4},
		{"let identity = fn(x) { return x;}; identity(3);", 3},
		{"let double = fn(x) { x * 2; }; double(7);", 14},
		{"let add = fn(a, b) { a + b; }; add(3, 4);", 7},
		{"let add = fn(a, b) { a + b; }; add(3 + 3, add(4, 4));", 14},
		{"fn(x) { x ;}(2)", 2},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, _ := testEval(tc.input)
			testIntObj(t, result, tc.want)
		})
	}
}

func TestClosures(t *testing.T) {
	input := `
let newAdder = fn(x) {
  fn(y) { x + y }
};

let addTwo = newAdder(2);
addTwo(3);
`
	r, _ := testEval(input)
	testIntObj(t, r, 5)
}

func TestStringLiteral(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"Hello String!"`, "Hello String!"},
		{`"Concatenate" + " " + "Me" + "!"`, "Concatenate Me!"},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, _ := testEval(tc.input)
			str, ok := result.(*object.Str)
			if !ok {
				t.Fatalf("result is of type %T, want *object.Str",
					result)
			}
			if string(*str) != tc.want {
				t.Errorf("str has %q, want %q", *str, tc.want)
			}

		})
	}

}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{"len(1)", BadBuiltinArg{name: "len", argtype: object.Integer}},
		{
			`len("one", "two")`,
			BadBuiltinNArgs{name: "len", nargs: 1, got: 2},
		},
		{"len([])", 0},
		{"len([1, 2])", 2},
		{"first([8,7,6])", 8},
		{
			"first(4)",
			BadBuiltinArg{name: "first", argtype: object.Integer},
		},
		{
			"first([1,2], [3,4])",
			BadBuiltinNArgs{name: "first", nargs: 1, got: 2},
		},
		{"last([8,7,6])", 6},
		{
			"last(4)",
			BadBuiltinArg{name: "last", argtype: object.Integer},
		},
		{
			"last([1,2], [3,4])",
			BadBuiltinNArgs{name: "last", nargs: 1, got: 2},
		},
		{"rest([1,2,3,4])", []int{2, 3, 4}},
		{"push([], 1)", []int{1}},
		{"push([8,7,6], 5)", []int{8, 7, 6, 5}},
		{
			"push(4, 4)",
			BadBuiltinArg{name: "push", argtype: object.Integer},
		},
		{
			"push([1,2])",
			BadBuiltinNArgs{name: "push", nargs: 2, got: 1},
		},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, err := testEval(tc.input)
			switch want := tc.want.(type) {
			case int:
				testIntObj(t, result, int64(want))
			case []int:
				arr, ok := result.(object.Arr)
				if !ok {
					t.Errorf("result is of type %T, want object.Arr",
						result)
					return
				}
				if len(arr) != len(want) {
					t.Errorf("unexpected len %d, want %d",
						len(arr), len(want))
				}
				for i, v := range want {
					testIntObj(t, arr[i], int64(v))
				}
			case error:
				if tc.want != err {
					t.Errorf("error is %q, want %q",
						err, tc.want)
				}
			default:
				t.Errorf("unexpected want value")
			}
		})
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"
	result, _ := testEval(input)
	arr, ok := result.(object.Arr)
	if !ok {
		t.Fatalf("result is of type %T, want object.Arr", result)
	}
	if len(arr) != 3 {
		t.Fatalf("array has %d elements, want 3", len(arr))
	}
	testIntObj(t, arr[0], 1)
	testIntObj(t, arr[1], 4)
	testIntObj(t, arr[2], 6)
}

func TestArrayIndexExpr(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
	}{
		{"[1,2,3][-1]", nil},
		{"[1,2,3][0]", 1},
		{"[1,2,3][1]", 2},
		{"[1,2,3][2]", 3},
		{"[1,2,3][3]", nil},
		{"[1, 2, 3][1 + 1];", 3},
		{"let i = 0; [1][i];", 1},
		{"let myArray = [1, 2, 3]; myArray[2];", 3},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, _ := testEval(tc.input)
			if i, ok := tc.want.(int); ok {
				testIntObj(t, result, int64(i))
				return
			}
			testIsNull(t, result)
		})
	}
}

func testEval(input string) (object.Object, error) {
	parse := parser.New(lexer.New(input))
	return Eval(parse.Program(), object.NewEnvironment())
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

func testIsNull(t *testing.T, obj object.Object) {
	if obj != &null {
		t.Errorf("obj is %+v, want object.Null", obj)
	}
}
