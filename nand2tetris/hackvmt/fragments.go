package main

import "fmt"

// pop fragment is a hack assembly fragment that pops an item of the
// stack into D
const popFrag = `	@SP
	AM=M-1
	D=M
`

// push fragment is a hack assembly fragment that will push what is in
// D into the stack.
const pushFrag = `	@SP
	A=M
	M=D
	@SP
	M=M+1
`

// push constant pushes a constant value into the stack
const pushConstantFrag = `	@%d
	D=A
%s`

func pushConstant(c int) string {
	return fmt.Sprintf(pushConstantFrag, c, pushFrag)
}

// push static fragment pushes a static variable in a global
// namespace into the stack
const pushStaticFrag = `	@%s.%d
	D=M
%s`

func pushStatic(name string, index int) string {
	return fmt.Sprintf(pushStaticFrag, name, index, pushFrag)
}

// pop static fragment pops a value from the stack into a global
// static variable.
const popStaticFrag = `%s	@%s.%d
	M=D
`

func popStatic(name string, index int) string {
	return fmt.Sprintf(popStaticFrag, popFrag, name, index)
}

// push segment pushes a stack value from one of the following
// segments; local, argument, this, that.
const pushSegmentFrag = `	@%d
	D=A
	@%s
	A=M+D
	D=M
%s`

var segMap = map[string]string{
	"local":    "LCL",
	"argument": "ARG",
	"this":     "THIS",
	"that":     "THAT",
}

func pushSegment(segment string, index int) (string, error) {
	addr, ok := segMap[segment]
	if !ok {
		return "", fmt.Errorf("translate: push unsupported memory segment %q", segment)
	}
	return fmt.Sprintf(pushSegmentFrag, index, addr, pushFrag), nil
}

// pop segment popes a value from the stack into the given segment
// index. Supported segments are: local, argument, this, that.
//
// NOTE: we use R13 as a temp storage of destination address
const popSegmentFrag = `	@%d
	D=A
	@%s
	D=M+D
	@R13
	M=D
%s	@R13
	A=M
	M=D
`

func popSegment(segment string, index int) (string, error) {
	addr, ok := segMap[segment]
	if !ok {
		return "", fmt.Errorf("translate: pop unsupported memory segment %q", segment)
	}
	return fmt.Sprintf(popSegmentFrag, index, addr, popFrag), nil
}

// push temp pushes a value from temp allocated registers into the stack.
const pushTempFrag = `	@%d
	D=A
	@5
	A=D+A
	D=M
%s`

func pushTemp(index int) (string, error) {
	if index > 8 {
		return "", fmt.Errorf("translate: index % out of bound for temp memory segment", index)
	}
	return fmt.Sprintf(pushTempFrag, index, pushFrag), nil
}

// pop temp pops a value from the stack into the temp allocated register
//
// NOTE: we store destination address in R14
const popTempFrag = `	@%d
	D=A
	@5
	D=D+A
	@R14
	M=D
%s	@R14
	A=M
	M=D
`

func popTemp(index int) (string, error) {
	if index > 8 {
		return "", fmt.Errorf("translate: index % out of bound for temp memory segment", index)
	}
	return fmt.Sprintf(popTempFrag, index, popFrag), nil
}

// push pointer pushes the base address of this/that into the stack.
const pushPointerFrag = `	@%s
	D=M
%s`

var pointerMap = [...]string{"THIS", "THAT"}

func pushPointer(i int) (string, error) {
	if i != 0 && i != 1 {
		return "", fmt.Errorf("translate: %d not a valid index for push pointer", i)
	}
	return fmt.Sprintf(pushPointerFrag, pointerMap[i], pushFrag), nil
}

const popPointerFrag = `%s	@%s
	M=D
`

func popPointer(i int) (string, error) {
	if i != 0 && i != 1 {
		return "", fmt.Errorf("translate: %d not a valid index for pop pointer", i)
	}
	return fmt.Sprintf(popPointerFrag, popFrag, pointerMap[i]), nil
}

// add pops a value from a stack and does an in-place modification of
// the result. Since there is no point popping to values adding and
// pushing the value back. Note addition is associative.
const addFrag = `%s	@SP
	A=M-1
	M=M+D
`

// subFrag pop a value from a stack and modifies in place the
// subtraction of the popped value from what is in the stack. Order
// here matters.
const subFrag = `%s	@SP
	A=M-1
	M=M-D
`

// neg negates the top of the stack.
const negFrag = `	@SP
	A=M-1
	M=-M
`

// branch fragments creates assembly instructions that allow for 3
// types of branching (EQ, GT, LT).
const branchFrag = `%s	@SP
	A=M-1
	D=M-D
	@%s
	D;J%s
	@SP
	A=M-1
	M=0	// false
	@%s
	0;JMP
(%s)
	@SP
	A=M-1
	M=-1	// true
(%s)
`

// branch creates a branching fragment based on t, t must be one of EQ/LT/GT.
func branch(name, t string, unique int) (string, error) {
	if !(t == "EQ" || t == "LT" || t == "GT") {
		return "", fmt.Errorf("translate: cannot branch with %q expecting one of EQ,LT,GT", t)
	}
	branch := fmt.Sprintf("%s.if_%s.%d", name, t, unique)
	end := fmt.Sprintf("%s.if_NOT.%d", name, unique)
	return fmt.Sprintf(branchFrag, popFrag, branch, t, end, branch, end), nil
}

// and fragment pops a value of the stack and perform logical and in
// place on top of the stack.
const andFrag = `%s	@SP
	A=M-1
	M=D&M
`

// or fragment pops a value from the top of the stack and performs
// logical or in-place on top of the stack.
const orFrag = `%s	@SP
	A=M-1
	M=D|M
`

// not fragment perform not logical operation on top the stack.
const notFrag = `	@SP
	A=M-1
	M=!M
`

// goto fragment unconditionally jump to label
const gotoFrag = `	@%s
	0;JMP
`

// if goto fragment pops a value of the stack and conditionally jumps.
const ifgotoFrag = `%s	@%s
	D;JNE
`

// return fragment. Stores frame into R13 & return-address in R14
const retFrag = `	@LCL
	D=M
	@R13	// FRAME
	M=D
	@5
	A=D-A	// FRAME-5 (return-address)
	D=M
	@R14	// (return-address)
	M=D
	@SP
	AM=M-1
	D=M
	@ARG	// *ARG = pop()
	A=M
	M=D
	@ARG
	D=M+1
	@SP	// SP = ARG+1
	M=D
	@R13
	AM=M-1	// FRAME-1
	D=M
	@THAT
	M=D
	@R13
	AM=M-1	// FRAME-2
	D=M
	@THIS
	M=D
	@R13
	AM=M-1	// FRAME-3
	D=M
	@ARG
	M=D
	@R13
	AM=M-1	// Frame-4
	D=M
	@LCL
	M=D
	@R14
	A=M
	0;JMP
`

// call fragment which prepares the current frame before transferring control.
const callFrag = `	@%s	// push return addr
	D=A
%s	@LCL	// push LCL
	D=M
%s	@ARG	// push ARG
	D=M
%s	@THIS	// push THIS
	D=M
%s	@THAT	// push THAT
	D=M
%s	@SP	// ARG = SP - n - 5
	D=M
	@%d
	D=D-A
	@5
	D=D-A
	@ARG
	M=D
	@SP
	D=M
	@LCL
	M=D
	@%s
	0;JMP
(%s)
`
