// Package object defines an internal representation of data in the
// Monkey language.
//
// Choices to represent these values as an object is to keep the
// implementation simple and extensible. And performance concerns are
// being dropped.
package object

import (
	"fmt"
	"strconv"
)

// Type describes the type of object being manipulated.
//go:generate stringer -type=Type
type Type int

// Enumerate different object types
const (
	Integer Type = iota
	Boolean
	Null
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
