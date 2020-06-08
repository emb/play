package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Hack VM specifies a list of commands. Refer to chapter 7 of the
// book.
type CommandType uint8

const (
	_ CommandType = iota
	CmdArithmetic
	CmdPush
	CmdPop
)

func (t CommandType) String() string {
	switch t {
	case CmdArithmetic:
		return "Arithmetic"
	case CmdPush:
		return "Push"
	case CmdPop:
		return "Pop"
	}
	return "Unknown"
}

// Command stores information about a command.
type Command struct {
	kind CommandType
	// arg stores the arithmetic operation for arithmetic commands
	// or the memory segment for memory access commands.
	arg string
	// an optional parameter (e.g. memory access index)
	param *int
}

func (c *Command) String() string {
	buf := bytes.NewBufferString(c.kind.String())
	buf.WriteString(" ")
	buf.WriteString(c.arg)
	if c.param != nil {
		buf.WriteString(" ")
		buf.WriteString(strconv.Itoa(*c.param))
	}
	return buf.String()
}

// Parser a helper struct to manage errors
type Parser struct {
	r   io.Reader
	err error
}

func NewParser(r io.Reader) *Parser {
	return &Parser{r: r}
}

// Parse starts a go routine that will parse incoming VM commands. It
// returns a channel that will be closed after finishing to
// scan. Ensure to check the parser error after the scan is finished.
func (p *Parser) Parse() <-chan *Command {
	ch := make(chan *Command)
	go func() {
		s := bufio.NewScanner(p.r) // By default we scan lines
		for s.Scan() {
			cmd, err := parse(s.Text())
			if err != nil {
				p.err = err
				break
			}
			// parse returns empty commands if it encounters comment lines.
			if cmd == nil {
				continue
			}
			ch <- cmd
		}
		if err := s.Err(); err != nil {
			p.err = err
		}
		close(ch)
	}()
	return ch
}

// Err returns parser errors if any.
func (p *Parser) Err() error {
	return p.err
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
		return &Command{kind: CmdArithmetic, arg: p0}, nil
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

	cmd := Command{arg: parts[1], param: &index}
	switch strings.ToLower(parts[0]) {
	case "push":
		cmd.kind = CmdPush
	case "pop":
		cmd.kind = CmdPop
	default:
		return nil, fmt.Errorf("parse: memory access command %q is invalid", parts[0])
	}
	return &cmd, nil
}
