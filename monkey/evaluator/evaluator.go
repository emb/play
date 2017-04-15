// Package evaluator implement an Eval for the Monkey interpreter
package evaluator

import (
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

// Eval evaluates the Monkey AST.
func Eval(node ast.Node) object.Object {
	switch n := node.(type) {
	// Statements
	case *ast.Program:
		// The result of a program is equivalent to the result
		// of the last statement in Monkey.
		return evalStmts(n.Statements)
	case *ast.ExpressionStmt:
		return Eval(n.Expression)
	case *ast.BlockStmt:
		return evalStmts(n.Statements)
	// Expressions
	case *ast.IntegerLiteral:
		i := object.Int(n.Value)
		return &i
	case *ast.Boolean:
		return objb(n.Value)
	case *ast.PrefixExpr:
		switch n.Operator {
		case token.BANG:
			return evalBang(Eval(n.Right))
		case token.MINUS:
			return evalMinus(Eval(n.Right))
		default:
			return &null
		}
	case *ast.InfixExpr:
		return evalInfix(n.Operator, Eval(n.Left), Eval(n.Right))
	case *ast.IfExpr:
		condition := Eval(n.Condition)
		if truthy(condition) {
			return Eval(n.Consequence)
		} else if n.Alternative != nil {
			return Eval(n.Alternative)
		} else {
			return &null
		}
	}
	return nil
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
func evalStmts(stmts []ast.Statement) object.Object {
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt)
	}
	return result
}

func evalBang(operand object.Object) object.Object {
	switch o := operand.(type) {
	case *object.Bool:
		if *o == yes {
			return &no
		}
		return &yes
	case *object.Nul:
		return &yes
	default:
		return &no
	}
}

func evalMinus(operand object.Object) object.Object {
	if i, ok := operand.(*object.Int); ok {
		negated := -(int64(*i))
		r := object.Int(negated)
		return &r
	}
	return &null
}

func evalInfix(op string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() == right.Type() && left.Type() == object.Integer:
		return evalInfixInts(op, left, right)

	// NOTE: the following comparisons are pointer
	// comparisons and since we deal with all other types
	// before we reach here we assume that what remains is
	// a comparison of the Boolean type.
	case op == token.EQ:
		return objb(left == right)
	case op == token.NEQ:
		return objb(left != right)

	default:
		return &null
	}
}

func evalInfixInts(op string, left, right object.Object) object.Object {
	l := int64(*left.(*object.Int))
	r := int64(*right.(*object.Int))
	switch op {
	case token.PLUS:
		return obji(l + r)
	case token.MINUS:
		return obji(l - r)
	case token.ASTERISK:
		return obji(l * r)
	case token.SLASH:
		return obji(l / r)
	case token.LT:
		return objb(l < r)
	case token.GT:
		return objb(l > r)
	case token.EQ:
		return objb(l == r)
	case token.NEQ:
		return objb(l != r)
	default:
		return &null
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
