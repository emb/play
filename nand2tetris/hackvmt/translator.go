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
		_, err = w.Write(translated)
		if err != nil {
			return err
		}
		unique++
	}
	return nil
}

func translate(c *Command, unique int) ([]byte, error) {
	switch c.kind {
	case CmdPush:
		return push(c.arg, *c.param)
	case CmdPop:
		return pop(c.arg, *c.param)
	case CmdArithmetic:
		return arith(c.arg, unique)
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

var addrMap = map[string]string{
	"local":    "LCL",
	"argument": "ARG",
	"this":     "THIS",
	"that":     "THAT",
	"temp":     "R5", // Address Number 5
}

func push(segment string, index int) ([]byte, error) {
	switch segment {
	case "local", "argument", "this", "that":
		addr := addrMap[segment]
		buf := bytes.NewBufferString(fmt.Sprintf("\t@%d\n", index))
		buf.WriteString("\tD=A\n")
		buf.WriteString(fmt.Sprintf(`	@%s
	A=M+D
	D=M
`, addr))
		buf.WriteString(pushFrag)
		return buf.Bytes(), nil
	case "temp":
		buf := bytes.NewBufferString(fmt.Sprintf("\t@%d\n", index))
		buf.WriteString("\tD=A\n")
		buf.WriteString(`	@5
	A=D+A
	D=M
`)
		buf.WriteString(pushFrag)
		return buf.Bytes(), nil
	case "constant":
		buf := bytes.NewBufferString(fmt.Sprintf("\t@%d\n", index))
		buf.WriteString("\tD=A\n")
		buf.WriteString(pushFrag)
		return buf.Bytes(), nil
	}
	return nil, fmt.Errorf("translate: push unsupported memory segment %q", segment)
}

// pop fragment is a hack assembly fragment that pops an item of the
// stack into D
const popFrag = `	@SP
	AM=M-1
	D=M
`

func pop(segment string, index int) ([]byte, error) {
	switch segment {
	case "local", "argument", "this", "that":
		addr := addrMap[segment]
		buf := bytes.NewBufferString(fmt.Sprintf("\t@%d\n", index))
		buf.WriteString("\tD=A\n")
		// Store address in R13
		buf.WriteString(fmt.Sprintf(`	@%s
	D=M+D
	@R13
	M=D
`, addr))
		buf.WriteString(popFrag)
		buf.WriteString(`	@R13
	A=M
	M=D
`)
		return buf.Bytes(), nil
	case "temp":
		if index > 8 {
			return nil, fmt.Errorf("translate: index % out of bound for temp memory segment", index)
		}
		buf := bytes.NewBufferString(fmt.Sprintf("\t@%d\n", index))
		buf.WriteString(`	D=A
	@5
	D=D+A
	@R14
	M=D
`)
		buf.WriteString(popFrag)
		buf.WriteString(`	@R14
	A=M
	M=D`)
		return buf.Bytes(), nil
	}
	return nil, fmt.Errorf("translate: pop unsupported memory segment %q", segment)
}

func arith(op string, unique int) ([]byte, error) {
	switch op {
	case "add":
		// Pop into D
		buf := bytes.NewBufferString(popFrag)
		// We don't have to pop here since we have a pointer
		// to the memory address we can manipulate it directly
		// rather than pop/push. We can add directly.
		buf.WriteString(`	@SP
	A=M-1
	M=M+D
`)
		return buf.Bytes(), nil
	case "sub":
		// Pop into D
		buf := bytes.NewBufferString(popFrag)
		buf.WriteString(`	@SP
	A=M-1
	M=M-D
`)
		return buf.Bytes(), nil
	case "neg":
		// Modify top of the stack with
		return []byte(`	@SP
	A=M-1
	M=-M
`), nil
	case "eq":
		return branch("EQ", unique), nil
	case "lt":
		return branch("LT", unique), nil
	case "gt":
		return branch("GT", unique), nil
	case "and":
		buf := bytes.NewBufferString(popFrag)
		buf.WriteString(`	@SP
	A=M-1
	M=D&M
`)
		return buf.Bytes(), nil
	case "or":
		buf := bytes.NewBufferString(popFrag)
		buf.WriteString(`	@SP
	A=M-1
	M=D|M
`)
		return buf.Bytes(), nil
	case "not":
		return []byte(`	@SP
	A=M-1
	M=!M
`), nil
	}
	return nil, fmt.Errorf("translate: unsupported arithmetic command %q", op)
}

// branch creates a branching fragment based on t, t must be one of EQ/LT/GT.
func branch(t string, unique int) []byte {
	branch := fmt.Sprintf("IF.%s.%d", t, unique) // e.g. IF.LT.1
	branchEnd := fmt.Sprintf("IF.END.%d", unique)
	buf := bytes.NewBufferString(popFrag)
	buf.WriteString(fmt.Sprintf(`	@SP
	A=M-1
	D=M-D
	@%s
	D;J%s
	@SP
	A=M-1
	M=0    // false
	@%s
	0;JMP
(%s)
	@SP
	A=M-1
	M=-1   // true
(%s)
`, branch, t, branchEnd, branch, branchEnd))
	return buf.Bytes()
}
