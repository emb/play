package main

import (
	"bytes"
	"fmt"
	"io"
)

// Translate converts VM code to assembly code.
func Translate(ch <-chan *Command, w io.Writer) error {
	// channel gets closed when we finish reading.
	for cmd := range ch {
		translated, err := translate(cmd)
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
		_, err = w.Write(translated)
		if err != nil {
			return err
		}
	}
	return nil
}

func translate(c *Command) ([]byte, error) {
	switch c.kind {
	case CmdPush:
		return push(c.arg, *c.param)
	case CmdPop:
		return pop(c.arg, *c.param)
	case CmdArithmetic:
		return arith(c.arg)
	}
	return nil, fmt.Errorf("translate: unsupported command %s", c)
}

// push fragment is a hack assembly fragment that will push what is in
// D into the stack.
const pushFrag = `	@SP
	A=M
	M=D
	@SP
	M=M+1
`

func push(segment string, index int) ([]byte, error) {
	switch segment {
	case "constant":
		buf := bytes.NewBufferString(fmt.Sprintf("\t@%d\n", index))
		buf.WriteString("\tD=A\n")
		buf.WriteString(pushFrag)
		return buf.Bytes(), nil
	}
	return nil, fmt.Errorf("translate: push unsupported memory segment %q", segment)
}

// pop fragment is a hack assembly fragment that pops an item of the
// stack into M
const popFrag = `	@SP
	AM=M-1
`

func pop(segment string, index int) ([]byte, error) {
	return nil, fmt.Errorf("translate: pop unsupported memory segment %q", segment)
}

func arith(op string) ([]byte, error) {
	switch op {
	case "add":
		// Pop into D
		buf := bytes.NewBufferString(popFrag)
		buf.WriteString("\tD=M\n")
		// We don't have to pop here since we have a pointer
		// to the memory address we can manipulate it directly
		// rather than pop/push. We can add directly.
		buf.WriteString(`	@SP
	A=M-1
	M=D+M
`)
		return buf.Bytes(), nil
	}
	return nil, fmt.Errorf("translate: unsupported arithmetic command %q", op)
}
