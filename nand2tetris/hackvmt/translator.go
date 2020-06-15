package main

import (
	"fmt"
	"io"
)

// Translate converts VM code to assembly code.
func Translate(ch <-chan *Command, w io.Writer) error {
	// unique is useful to add for any labels that maybe used.
	// there is no semantic
	unique := 0
	// channel gets closed when we finish reading.
	for cmd := range ch {
		translated, err := translate(cmd, unique)
		if err != nil {
			return err
		}
		// Add a comment
		comment := []byte(fmt.Sprintf("\t// %s\n", cmd))
		_, err = w.Write(comment)
		if err != nil {
			return err
		}
		// Write assembly
		_, err = w.Write([]byte(translated))
		if err != nil {
			return err
		}
		unique++
	}
	return nil
}

func translate(c *Command, unique int) (string, error) {
	switch c.Type {
	// Memory Access
	case CmdPush:
		return push(c.Namespace, c.Arg, *c.Param)
	// Arithmetic
	case CmdPop:
		return pop(c.Namespace, c.Arg, *c.Param)
	case CmdArithmetic:
		return arith(c.Arg, unique)
	// Control
	case CmdLabel:
		return label(c.Namespace, c.Scope, c.Arg), nil
	case CmdGoto:
		return gotolabel(c.Namespace, c.Scope, c.Arg), nil
	case CmdIfGoto:
		return ifgotolabel(c.Namespace, c.Scope, c.Arg), nil
	}
	return "", fmt.Errorf("translate: unsupported command %s", c)
}

func push(name, segment string, index int) (string, error) {
	switch segment {
	case "local", "argument", "this", "that":
		return pushSegment(segment, index)
	case "temp":
		return pushTemp(index)
	case "constant":
		return pushConstant(index), nil
	case "static":
		return pushStatic(name, index), nil
	case "pointer":
		return pushPointer(index)
	}
	return "", fmt.Errorf("translate: push unsupported memory segment %q", segment)
}

func pop(name, segment string, index int) (string, error) {
	switch segment {
	case "local", "argument", "this", "that":
		return popSegment(segment, index)
	case "temp":
		return popTemp(index)
	case "static":
		return popStatic(name, index), nil
	case "pointer":
		return popPointer(index)
	}
	return "", fmt.Errorf("translate: pop unsupported memory segment %q", segment)
}

func arith(op string, unique int) (string, error) {
	switch op {
	case "add":
		return fmt.Sprintf(addFrag, popFrag), nil
	case "sub":
		return fmt.Sprintf(subFrag, popFrag), nil
	case "neg":
		return negFrag, nil
	case "eq":
		return branch("EQ", unique)
	case "lt":
		return branch("LT", unique)
	case "gt":
		return branch("GT", unique)
	case "and":
		return fmt.Sprintf(andFrag, popFrag), nil
	case "or":
		return fmt.Sprintf(orFrag, popFrag), nil
	case "not":
		return notFrag, nil
	}
	return "", fmt.Errorf("translate: unsupported arithmetic command %q", op)
}

// genLabel generate a label.
func genLabel(name, scope, label string) string { return fmt.Sprintf("%s.%s$%s", name, scope, label) }

func label(name, scope, label string) string {
	return fmt.Sprintf("(%s)\n", genLabel(name, scope, label))
}

func gotolabel(name, scope, label string) string {
	return fmt.Sprintf(gotoFrag, genLabel(name, scope, label))
}

func ifgotolabel(name, scope, label string) string {
	return fmt.Sprintf(ifgotoFrag, popFrag, genLabel(name, scope, label))
}
