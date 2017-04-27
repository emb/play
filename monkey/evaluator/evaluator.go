// Package evaluator implement an Eval for the Monkey interpreter
package evaluator

import (
	"errors"
	"fmt"

	"github.com/emb/play/monkey/ast"
	"github.com/emb/play/monkey/object"
	"github.com/emb/play/monkey/token"
)

// easy references that can be used by the evaluator.
var (
	null = object.Nul{}
	yes  = object.Bool(true)
	no   = object.Bool(false)
)

// OpTypeMismatch describes a type mismatch in an infix operator
type OpTypeMismatch struct {
	left  object.Type
	op    string
	right object.Type
}

// Error returns string representing an error
func (e OpTypeMismatch) Error() string {
	return fmt.Sprintf("type mismatch: %s %s %s", e.left, e.op, e.right)
}

// BadPrefixOp describes an unknown bad prefix operator
type BadPrefixOp struct {
	op    string
	right object.Type
}

// Error returns a string describing the error
func (e BadPrefixOp) Error() string {
	return fmt.Sprintf("bad operator: %s%s", e.op, e.right)
}

// BadInfixOp describes an invalid infix operation
type BadInfixOp struct {
	left  object.Type
	op    string
	right object.Type
}

// Error returns a string describing the error
func (e BadInfixOp) Error() string {
	return fmt.Sprintf("bad operation: %s %s %s", e.left, e.op, e.left)
}

// UnboundIdent is an error returned if an identifier is not
// bound/found.
type UnboundIdent struct {
	ident string
}

// Error returns a string describing the unbound identifier error.
func (e UnboundIdent) Error() string {
	return fmt.Sprintf("unbound identifier: %s", e.ident)
}

// BadFn is an error that happens when trying to evaluate a call
// expression on a non function object.
type BadFn struct {
	exp object.Type
}

// Error returns a string describing the error
func (b BadFn) Error() string {
	return fmt.Sprintf("bad fn call, %s is not a function", b.exp)
}

// BadBuiltinArg describes an unexpected builtin function call with
// unsupported argument type
type BadBuiltinArg struct {
	name    string
	argtype object.Type
}

// Error returns a string describing the error
func (b BadBuiltinArg) Error() string {
	return fmt.Sprintf("bad argument type %s for bultin in '%s'",
		b.argtype, b.name)
}

// BadBuiltinNArgs describes the wrong number of arguments to a builtin
// function.
type BadBuiltinNArgs struct {
	name  string
	nargs int
	got   int
}

// Error returns a string describing the error
func (b BadBuiltinNArgs) Error() string {
	return fmt.Sprintf("bad number of arguments %d to builtin '%s' which expects %d",
		b.got, b.name, b.nargs)
}

// ErrUnexpected is an unexpected error within the evaluator it should
// not happen
var ErrUnexpected = errors.New("unexpected error")

var builtins = map[string]*object.BuiltinFunct{
	"len": &object.BuiltinFunct{
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, BadBuiltinNArgs{
					name:  "len",
					nargs: 1,
					got:   len(args),
				}
			}
			switch arg := args[0].(type) {
			case *object.Str:
				result := object.Int(len(*arg))
				return &result, nil
			case object.Arr:
				result := object.Int(len(arg))
				return &result, nil
			default:
				return nil, BadBuiltinArg{
					name:    "len",
					argtype: args[0].Type(),
				}
			}

		},
	},
	"first": &object.BuiltinFunct{
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, BadBuiltinNArgs{
					name:  "first",
					nargs: 1,
					got:   len(args),
				}
			}
			arr, ok := args[0].(object.Arr)
			if !ok {
				return nil, BadBuiltinArg{
					name:    "first",
					argtype: args[0].Type(),
				}
			}
			if len(arr) > 0 {
				return arr[0], nil
			}
			return &null, nil
		},
	},
	"last": &object.BuiltinFunct{
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, BadBuiltinNArgs{
					name:  "last",
					nargs: 1,
					got:   len(args),
				}
			}
			arr, ok := args[0].(object.Arr)
			if !ok {
				return nil, BadBuiltinArg{
					name:    "last",
					argtype: args[0].Type(),
				}
			}
			if len(arr) > 0 {
				return arr[len(arr)-1], nil
			}
			return &null, nil
		},
	},
	"rest": &object.BuiltinFunct{
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, BadBuiltinNArgs{
					name:  "rest",
					nargs: 1,
					got:   len(args),
				}
			}
			arr, ok := args[0].(object.Arr)
			if !ok {
				return nil, BadBuiltinArg{
					name:    "rest",
					argtype: args[0].Type(),
				}
			}
			if len(arr) > 0 {
				ret := make([]object.Object, len(arr)-1)
				copy(ret, arr[1:])
				return object.Arr(ret), nil
			}
			return &null, nil
		},
	},
	"push": &object.BuiltinFunct{
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, BadBuiltinNArgs{
					name:  "push",
					nargs: 2,
					got:   len(args),
				}
			}
			arr, ok := args[0].(object.Arr)
			if !ok {
				return nil, BadBuiltinArg{
					name:    "push",
					argtype: args[0].Type(),
				}
			}
			ret := make([]object.Object, len(arr)+1)
			copy(ret, arr)
			ret[len(arr)] = args[1]
			return object.Arr(ret), nil
		},
	},
	"puts": &object.BuiltinFunct{
		Fn: func(args ...object.Object) (object.Object, error) {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return &null, nil
		},
	},
}

// Eval evaluates the Monkey AST.
func Eval(node ast.Node, env *object.Environment) (object.Object, error) {
	switch n := node.(type) {
	// Statements
	case *ast.Program:
		// The result of a program is equivalent to the result
		// of the last statement in Monkey. Furthermore,
		// receiving a return object requires the result to be
		// unwrapped.
		result, err := evalStmts(n.Statements, env)
		if err != nil {
			return nil, err
		}
		return unwrap(result), nil
	case *ast.ExpressionStmt:
		return Eval(n.Expression, env)
	case *ast.BlockStmt:
		return evalStmts(n.Statements, env)
	case *ast.ReturnStmt:
		v, err := Eval(n.Value, env)
		return &object.Ret{Value: v}, err
	case *ast.LetStmt:
		result, err := Eval(n.Value, env)
		if err != nil {
			return nil, err
		}
		env.Set(n.Name.Value, result)
		return nil, nil
	// Expressions
	case *ast.IntegerLiteral:
		i := object.Int(n.Value)
		return &i, nil
	case *ast.StringLiteral:
		s := object.Str(n.Value)
		return &s, nil
	case *ast.Boolean:
		return objb(n.Value), nil
	case *ast.ArrayLiteral:
		result, err := evalExprs(n.Elements, env)
		if err != nil {
			return nil, err
		}
		return object.Arr(result), nil
	case *ast.HashLiteral:
		return evalHash(n, env)
	case *ast.PrefixExpr:
		r, err := Eval(n.Right, env)
		if err != nil {
			return nil, err
		}

		switch n.Operator {
		case token.BANG:
			return evalBang(r)
		case token.MINUS:
			return evalMinus(r)
		default:
			return nil, BadPrefixOp{op: n.Operator, right: r.Type()}
		}
	case *ast.InfixExpr:
		l, err := Eval(n.Left, env)
		if err != nil {
			return nil, err
		}
		r, err := Eval(n.Right, env)
		if err != nil {
			return nil, err
		}
		return evalInfix(n.Operator, l, r)
	case *ast.IfExpr:
		condition, err := Eval(n.Condition, env)
		if err != nil {
			return nil, err
		}
		if truthy(condition) {
			return Eval(n.Consequence, env)
		} else if n.Alternative != nil {
			return Eval(n.Alternative, env)
		} else {
			return &null, nil
		}
	case *ast.Identifier:
		if v, ok := env.Get(n.Value); ok {
			return v, nil
		}
		if b, ok := builtins[n.Value]; ok {
			return b, nil
		}
		return nil, UnboundIdent{ident: n.Value}
	case *ast.FunctionLiteral:
		return &object.Funct{
			Env:        env,
			Parameters: n.Parameters,
			Body:       n.Body,
		}, nil
	case *ast.CallExpr:
		fn, err := Eval(n.Function, env)
		if err != nil {
			return nil, err
		}
		args, err := evalExprs(n.Arguments, env)
		if err != nil {
			return nil, err
		}
		return apply(fn, args)
	case *ast.IndexExpr:
		left, err := Eval(n.Left, env)
		if err != nil {
			return nil, err
		}
		index, err := Eval(n.Index, env)
		if err != nil {
			return nil, err
		}
		return evalIndex(left, index)
	}

	return nil, ErrUnexpected
}

func objb(b bool) *object.Bool {
	if b {
		return &yes
	}
	return &no
}

func obji(i int64) *object.Int {
	r := object.Int(i)
	return &r
}

// evalStmts evaluate each statement and returns the result of the
// last one.
func evalStmts(stmts []ast.Statement, env *object.Environment) (object.Object, error) {
	var (
		result object.Object
		err    error
	)
	for _, stmt := range stmts {
		result, err = Eval(stmt, env)
		if err != nil {
			return nil, err
		}
		// Early exit during a return statement. Note the
		// return object is not being unwrapped here because
		// doing so will prevent nested block statements from
		// working as expected. We bubble the return object
		// and it becomes the root node responsibility to
		// unwrap the result.
		if result != nil && result.Type() == object.Return {
			return result, nil
		}
	}
	return result, nil
}

func evalBang(operand object.Object) (object.Object, error) {
	switch o := operand.(type) {
	case *object.Bool:
		if *o == yes {
			return &no, nil
		}
		return &yes, nil
	case *object.Nul:
		return &yes, nil
	default:
		return &no, nil
	}
}

func evalMinus(operand object.Object) (object.Object, error) {
	if i, ok := operand.(*object.Int); ok {
		negated := -(int64(*i))
		r := object.Int(negated)
		return &r, nil
	}
	return nil, BadPrefixOp{op: "-", right: operand.Type()}
}

func evalInfix(op string, left object.Object, right object.Object) (object.Object, error) {
	switch {
	case left.Type() != right.Type():
		return nil, OpTypeMismatch{
			left:  left.Type(),
			op:    op,
			right: right.Type(),
		}
	case left.Type() == object.Integer:
		return evalInfixInts(op, left, right)
	case left.Type() == object.String:
		return evalInfixStrs(op, left, right)

	// NOTE: the following comparisons are pointer
	// comparisons and since we deal with all other types
	// before we reach here we assume that what remains is
	// a comparison of the Boolean type.
	case op == token.EQ:
		return objb(left == right), nil
	case op == token.NEQ:
		return objb(left != right), nil

	default:
		return nil, BadInfixOp{
			left:  left.Type(),
			op:    op,
			right: right.Type(),
		}
	}
}

func evalInfixInts(op string, left, right object.Object) (object.Object, error) {
	l := int64(*left.(*object.Int))
	r := int64(*right.(*object.Int))
	switch op {
	case token.PLUS:
		return obji(l + r), nil
	case token.MINUS:
		return obji(l - r), nil
	case token.ASTERISK:
		return obji(l * r), nil
	case token.SLASH:
		return obji(l / r), nil
	case token.LT:
		return objb(l < r), nil
	case token.GT:
		return objb(l > r), nil
	case token.EQ:
		return objb(l == r), nil
	case token.NEQ:
		return objb(l != r), nil
	default:
		return nil, ErrUnexpected
	}
}

func evalInfixStrs(op string, l, r object.Object) (object.Object, error) {
	if op != token.PLUS {
		return nil, BadInfixOp{left: l.Type(), op: op, right: r.Type()}
	}
	lv := l.(*object.Str)
	rv := r.(*object.Str)
	result := object.Str(string(*lv) + string(*rv))
	return &result, nil
}

func truthy(obj object.Object) bool {
	switch o := obj.(type) {
	case *object.Bool:
		if *o == yes {
			return true
		}
		return false
	case *object.Nul:
		return false
	default:
		return true
	}
}

func evalExprs(exps []ast.Expression, env *object.Environment) ([]object.Object, error) {
	result := make([]object.Object, len(exps))
	for i, exp := range exps {
		r, err := Eval(exp, env)
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}

func evalIndex(left, index object.Object) (object.Object, error) {
	switch {
	case left.Type() == object.Array && index.Type() == object.Integer:
		arr := left.(object.Arr)
		i := int64(*index.(*object.Int))
		max := int64(len(arr) - 1)
		if i < 0 || i > max {
			return &null, nil
		}
		return arr[i], nil
	case left.Type() == object.Hash:
		hash := left.(*object.HashMap)
		key, ok := index.(object.Hashable)
		if !ok {
			return nil, badkey(index.Type())
		}
		pair, ok := hash.Pairs[key.HashKey()]
		if !ok {
			return &null, nil
		}
		return pair.Value, nil
	default:
		return nil, fmt.Errorf("bad index operator on type %s",
			left.Type())
	}
}

func badkey(t object.Type) error {
	return fmt.Errorf("bad key %s for a hash", t)
}

func evalHash(n *ast.HashLiteral, env *object.Environment) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair)
	for kn, vn := range n.Pairs {
		k, err := Eval(kn, env)
		if err != nil {
			return nil, err
		}
		hk, ok := k.(object.Hashable)
		if !ok {
			return nil, badkey(k.Type())
		}
		v, err := Eval(vn, env)
		if err != nil {
			return nil, err
		}
		pairs[hk.HashKey()] = object.HashPair{Key: k, Value: v}
	}
	return &object.HashMap{Pairs: pairs}, nil
}

func apply(fn object.Object, args []object.Object) (object.Object, error) {
	switch fn := fn.(type) {
	case *object.Funct:
		return Eval(fn.Body, makeFnEnv(fn, args))
	case *object.BuiltinFunct:
		return fn.Fn(args...)
	default:
		return nil, BadFn{exp: fn.Type()}
	}
}

func makeFnEnv(fn *object.Funct, args []object.Object) *object.Environment {
	env := object.NewEnvironment().Extend(fn.Env)
	for i, p := range fn.Parameters {
		// NOTE: assuming parameter evaluation order. Args are
		// the result of evaluating the arguments of a
		// function call, the order of the parameters and
		// their results should match.
		env.Set(p.Value, args[i])
	}
	return env
}

// nnwrap a return value if o is a object.ReturnValue
func unwrap(o object.Object) object.Object {
	if ret, ok := o.(*object.Ret); ok {
		return ret.Value
	}
	return o

}
