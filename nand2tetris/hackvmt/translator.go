package main

import (
	"bytes"
	"fmt"
	"io"
)

// Translate converts VM code to assembly code.
func Translate(ch <-chan *Command, w io.Writer) error {
	// unique is useful to add for any labels that maybe used.
	// there is no semantic
	unique := 0
	// Write initialization
	_, err := w.Write([]byte(bootstrap(unique)))
	if err != nil {
		return err
	}
	unique++
	// channel gets closed when we finish reading.
	for cmd := range ch {
		translated, err := translate(cmd, unique)
		if err != nil {
			return err
		}
		// Add a comment
		comment := []byte(fmt.Sprintf("// %s\n", cmd))
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

func bootstrap(unique int) string {
	// Set stack pointer
	sp := `// Bootstrap
	@256
	D=A
	@SP
	M=D
%s`
	return fmt.Sprintf(sp, call("Sys", "Sys.init", 0, unique))
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
		return arith(c.Namespace, c.Arg, unique)
	// Control
	case CmdLabel:
		return label(c.Namespace, c.Arg), nil
	case CmdGoto:
		return gotolabel(c.Namespace, c.Arg), nil
	case CmdIfGoto:
		return ifgotolabel(c.Namespace, c.Arg), nil
	// Functions & function call
	case CmdFunction:
		return function(c.Arg, *c.Param), nil
	case CmdReturn:
		return retFrag, nil
	case CmdCall:
		return call(c.Namespace, c.Arg, *c.Param, unique), nil
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

func arith(name, op string, unique int) (string, error) {
	switch op {
	case "add":
		return fmt.Sprintf(addFrag, popFrag), nil
	case "sub":
		return fmt.Sprintf(subFrag, popFrag), nil
	case "neg":
		return negFrag, nil
	case "eq":
		return branch(name, "EQ", unique)
	case "lt":
		return branch(name, "LT", unique)
	case "gt":
		return branch(name, "GT", unique)
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
func genLabel(name, label string) string { return fmt.Sprintf("%s$%s", name, label) }

func label(name, label string) string {
	return fmt.Sprintf("(%s)\n", genLabel(name, label))
}

func gotolabel(name, label string) string {
	return fmt.Sprintf(gotoFrag, genLabel(name, label))
}

func ifgotolabel(name, label string) string {
	return fmt.Sprintf(ifgotoFrag, popFrag, genLabel(name, label))
}

func function(name string, locals int) string {
	buf := bytes.NewBufferString(fmt.Sprintf("(%s)\n", name))
	// Push a single constant 0 which should set D=0
	if locals > 0 {
		buf.WriteString(pushConstant(0))
	}
	// Since D=0 continue to push the stack for as long as we have locals.
	for i := 1; i < locals; i++ {
		buf.WriteString(`	A=M
	M=D
	@SP
	M=M+1
`)
	}
	return buf.String()
}

func call(namespace, function string, args, unique int) string {
	ret := fmt.Sprintf("%s.ret.%d", namespace, unique)
	return fmt.Sprintf(callFrag, ret, pushFrag, pushFrag, pushFrag, pushFrag, pushFrag, args, function, ret)
}
