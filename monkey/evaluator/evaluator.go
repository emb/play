// Package evaluator implement an Eval for the Monkey interpreter
package evaluator

import (
	"github.com/emb/play/monkey/ast"
	"github.com/emb/play/monkey/object"
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
		i := object.Bool(n.Value)
		return &i
	}
	return nil
}
