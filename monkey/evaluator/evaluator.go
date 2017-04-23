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
	return fmt.Sprintf("bad fn call, %T is not a function", b.exp)
}

// ErrUnexpected is an unexpected error within the evaluator it should
// not happen
var ErrUnexpected = errors.New("unexpected error")

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
		v, ok := env.Get(n.Value)
		if !ok {
			return nil, UnboundIdent{ident: n.Value}
		}
		return v, nil
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

func apply(fn object.Object, args []object.Object) (object.Object, error) {
	funct, ok := fn.(*object.Funct)
	if !ok {
		return nil, BadFn{exp: fn.Type()}
	}
	result, err := Eval(funct.Body, makeFnEnv(funct, args))
	if err != nil {
		return nil, err
	}
	return result, nil
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
