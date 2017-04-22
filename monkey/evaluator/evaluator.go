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

// ErrUnexpected is an unexpected error within the evaluator it should
// not happen
var ErrUnexpected = errors.New("unexpected error")

// Eval evaluates the Monkey AST.
func Eval(node ast.Node) (object.Object, error) {
	switch n := node.(type) {
	// Statements
	case *ast.Program:
		// The result of a program is equivalent to the result
		// of the last statement in Monkey. Furthermore,
		// receiving a return object requires the result to be
		// unwrapped.
		result, err := evalStmts(n.Statements)
		if err != nil {
			return nil, err
		}
		if ret, ok := result.(*object.Ret); ok {
			return ret.Value, nil
		}
		return result, nil
	case *ast.ExpressionStmt:
		return Eval(n.Expression)
	case *ast.BlockStmt:
		return evalStmts(n.Statements)
	case *ast.ReturnStmt:
		v, err := Eval(n.Value)
		return &object.Ret{Value: v}, err
	// Expressions
	case *ast.IntegerLiteral:
		i := object.Int(n.Value)
		return &i, nil
	case *ast.Boolean:
		return objb(n.Value), nil
	case *ast.PrefixExpr:
		r, err := Eval(n.Right)
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
		l, err := Eval(n.Left)
		if err != nil {
			return nil, err
		}
		r, err := Eval(n.Right)
		if err != nil {
			return nil, err
		}
		return evalInfix(n.Operator, l, r)
	case *ast.IfExpr:
		condition, err := Eval(n.Condition)
		if err != nil {
			return nil, err
		}
		if truthy(condition) {
			return Eval(n.Consequence)
		} else if n.Alternative != nil {
			return Eval(n.Alternative)
		} else {
			return &null, nil
		}
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
func evalStmts(stmts []ast.Statement) (object.Object, error) {
	var (
		result object.Object
		err    error
	)
	for _, stmt := range stmts {
		result, err = Eval(stmt)
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
