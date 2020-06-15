package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Hack VM specifies a list of commands. Refer to chapter 7 of the
// book.
type CommandType uint8

const (
	_ CommandType = iota
	CmdArithmetic
	CmdPush
	CmdPop
	CmdFunction
	CmdCall
	CmdReturn
	CmdLabel
	CmdGoto
	CmdIfGoto
)

func (t CommandType) String() string {
	switch t {
	case CmdArithmetic:
		return "Arithmetic"
	case CmdPush:
		return "Push"
	case CmdPop:
		return "Pop"
	case CmdLabel:
		return "Label"
	case CmdGoto:
		return "Goto"
	case CmdIfGoto:
		return "If-Goto"
	case CmdFunction:
		return "Function"
	case CmdCall:
		return "Call"
	case CmdReturn:
		return "Return"
	}
	return "Unknown"
}

// Command stores information about a command.
type Command struct {
	// Name a name space for the command typically the file name.
	Namespace string
	// Scope defines if the command is in a function or global scope.
	Scope string
	Type  CommandType
	// Arg stores the arithmetic operation for arithmetic commands
	// or the memory segment for memory access commands.
	Arg string
	// an optional parameter (e.g. memory access index)
	Param *int
}

func (c *Command) String() string {
	buf := bytes.NewBufferString(c.Type.String())
	buf.WriteString(" ")
	buf.WriteString(c.Arg)
	if c.Param != nil {
		buf.WriteString(" ")
		buf.WriteString(strconv.Itoa(*c.Param))
	}
	return buf.String()
}

// Parse parses commands from r into ch. Commands will have a
// namespace of name and will immediately return an error if it
// encounters one.
func Parse(name string, r io.Reader, ch chan<- *Command) error {
	s := bufio.NewScanner(r) // By default we scan lines
	scope := "global"
	for s.Scan() {
		cmd, err := parse(s.Text())
		if err != nil {
			return err
		}
		// parse returns empty commands if it encounters comment lines.
		if cmd == nil {
			continue
		}
		cmd.Namespace = name
		cmd.Scope = scope
		ch <- cmd
		if cmd.Type == CmdFunction {
			scope = cmd.Arg
		}
		if cmd.Type == CmdReturn {
			scope = "global"
		}
	}
	if err := s.Err(); err != nil {
		return err
	}
	return nil
}

func parse(line string) (*Command, error) {
	// Remove comments
	if i := strings.Index(line, "//"); i >= 0 {
		line = line[:i]
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		// White space
		return nil, nil
	}

	p0 := strings.ToLower(parts[0])
	switch p0 {
	// Memory access
	case "push", "pop":
		return parseMemory(parts)
	// Arithmetic
	case "add", "sub", "neg",
		"eq", "gt", "lt",
		"and", "or", "not":
		if len(parts) > 1 {
			return nil, fmt.Errorf("parse: arithmetic commands should not have parts in line %q",
				line)
		}
		return &Command{Type: CmdArithmetic, Arg: p0}, nil
	// Function Calls
	case "function", "call":
		return parseFunction(parts)
	case "return":
		if len(parts) > 1 {
			return nil, fmt.Errorf("parse: invalid return with arguments/parameters in line %q",
				line)
		}
		return &Command{Type: CmdReturn}, nil
	// Program Flow
	case "label":
		return parseLabel(parts)
	case "goto":
		if len(parts) > 2 {
			return nil, fmt.Errorf("parse: invalid goto with more than 2 parameters in line %q",
				line)
		}
		return &Command{Type: CmdGoto, Arg: parts[1]}, nil
	case "if-goto":
		if len(parts) > 2 {
			return nil, fmt.Errorf("parse: invalid if-goto with more than 2 parameters in line %q",
				line)
		}
		return &Command{Type: CmdIfGoto, Arg: parts[1]}, nil
	}
	return nil, fmt.Errorf("parse: unknown command %q", parts[0])
}

// segments is a set of valid segments in hack memory access commands
var segments = map[string]struct{}{
	"argument": struct{}{},
	"local":    struct{}{},
	"static":   struct{}{},
	"constant": struct{}{},
	"this":     struct{}{},
	"that":     struct{}{},
	"pointer":  struct{}{},
	"temp":     struct{}{},
}

func parseMemory(parts []string) (*Command, error) {
	if len(parts) != 3 {
		return nil, fmt.Errorf("parse: memory access commands should have 3 parts got %d", len(parts))
	}

	index, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("parse: memory access command index parsing: %w", err)
	}

	if _, ok := segments[parts[1]]; !ok {
		return nil, fmt.Errorf("parse: memory access command segment %q is invalid", parts[1])
	}

	cmd := Command{Arg: parts[1], Param: &index}
	switch strings.ToLower(parts[0]) {
	case "push":
		cmd.Type = CmdPush
	case "pop":
		cmd.Type = CmdPop
	default:
		return nil, fmt.Errorf("parse: memory access command %q is invalid", parts[0])
	}
	return &cmd, nil
}

func parseFunction(parts []string) (*Command, error) {
	if len(parts) != 3 {
		return nil, fmt.Errorf("parse: %s should have 3 parts got %d", parts[0], len(parts))
	}

	n, err := strconv.Atoi(parts[2])
	if err != nil {
		m := map[string]string{
			"function": "locals",
			"call":     "arguments",
		}
		return nil, fmt.Errorf("parse: %s invalid %s: %w", parts[0], m[parts[0]], err)
	}

	cmd := Command{Arg: parts[1], Param: &n}
	switch strings.ToLower(parts[0]) {
	case "function":
		cmd.Type = CmdFunction
	case "call":
		cmd.Type = CmdCall
	default:
		return nil, fmt.Errorf("parse: invalid function parsing %s", parts[0])
	}
	return &cmd, nil
}

func parseLabel(parts []string) (*Command, error) {
	if len(parts) != 2 {
		return nil, fmt.Errorf("parse: invalid label command with more than 2 parameters got %d parameters",
			len(parts))
	}

	label := parts[1]
	if unicode.IsDigit(rune(label[0])) {
		return nil, fmt.Errorf("parse: invalid label %q that starts with a digit", label)
	}
	for _, r := range []rune(label) {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) &&
			r != '.' && r != '_' && r != ':' {
			return nil, fmt.Errorf("parse: invalid %c in label %q", r, label)
		}
	}
	return &Command{Type: CmdLabel, Arg: label}, nil
}
