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
		var result object.Object
		for _, stmt := range n.Statements {
			result = Eval(stmt)
		}
		return result
	case *ast.ExpressionStmt:
		return Eval(n.Expression)
	// Expressions
	case *ast.IntegerLiteral:
		i := object.Int(n.Value)
		return &i
	case *ast.Boolean:
		if n.Value {
			return &yes
		}
		return &no
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
	}
	return nil
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
	default:
		return &null
	}
}

func evalInfixInts(op string, left, right object.Object) object.Object {
	l := int64(*left.(*object.Int))
	r := int64(*right.(*object.Int))
	switch op {
	case token.PLUS:
		ret := object.Int(l + r)
		return &ret
	case token.MINUS:
		ret := object.Int(l - r)
		return &ret
	case token.ASTERISK:
		ret := object.Int(l * r)
		return &ret
	case token.SLASH:
		ret := object.Int(l / r)
		return &ret

	default:
		return &null
	}
}
