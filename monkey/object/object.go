// Package object defines an internal representation of data in the
// Monkey language.
//
// Choices to represent these values as an object is to keep the
// implementation simple and extensible. And performance concerns are
// being dropped.
package object

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/emb/play/monkey/ast"
)

// Type describes the type of object being manipulated.
//go:generate stringer -type=Type
type Type int

// Enumerate different object types
const (
	Integer Type = iota
	Boolean
	Null
	Return
	Function
)

// Object is an internal representation of values in the monkey
// language.
type Object interface {
	Type() Type
	Inspect() string
}

// Int represents an integer value within monkey
type Int int64

// Type returns the object type
func (i *Int) Type() Type { return Integer }

// Inspect provides a string representation of an Int value.
func (i *Int) Inspect() string { return strconv.FormatInt(int64(*i), 10) }

// Bool represents a boolean value
type Bool bool

// Type returns the object type
func (*Bool) Type() Type { return Boolean }

// Inspect provides a string representation of boolean value.
func (b *Bool) Inspect() string { return fmt.Sprintf("%t", *b) }

// Nul represents an absence of a value
type Nul struct{}

// Type return the object type
func (*Nul) Type() Type { return Null }

// Inspect provides a string representation of null value.
func (*Nul) Inspect() string { return "null" }

// Ret wraps a return value
type Ret struct {
	Value Object
}

// Type returns the object type
func (*Ret) Type() Type { return Return }

// Inspect provides a string representation of the return value
func (r *Ret) Inspect() string {
	return fmt.Sprintf("return(%s)", r.Value)
}

// NewEnvironment creates an environment used while evaluating Monkey
// program.
func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Object)}
}

// Environment is where let statement binds values to identifiers
type Environment struct {
	store map[string]Object
	outer *Environment
}

// Extend extends the current environment with an outer scoped environment.
func (e *Environment) Extend(o *Environment) *Environment {
	e.outer = o
	return e
}

// Get returns an object bound to an identifier i in an environment
func (e *Environment) Get(i string) (Object, bool) {
	if e == nil {
		return nil, false
	}
	v, ok := e.store[i]
	if !ok {
		return e.outer.Get(i)
	}
	return v, ok
}

// Set stores a value v bound to identifier i in an environment
func (e *Environment) Set(i string, v Object) {
	e.store[i] = v
}

// Funct is an object that describes a function that can be evaluated.
type Funct struct {
	Env        *Environment
	Parameters []*ast.Identifier
	Body       *ast.BlockStmt
}

// Type returns the object type
func (*Funct) Type() Type { return Function }

// Inspect provides a string representation of a function
func (f *Funct) Inspect() string {
	var buf bytes.Buffer
	params := make([]string, len(f.Parameters))
	for i, p := range f.Parameters {
		params[i] = p.String()
	}
	buf.WriteString("fn (")
	buf.WriteString(strings.Join(params, ", "))
	buf.WriteByte(')')
	buf.WriteString(f.Body.String())
	return buf.String()
}
