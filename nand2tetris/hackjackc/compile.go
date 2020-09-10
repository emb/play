package main

import (
	"io"
	"strconv"
	"fmt"
)

// Use AST structs here to generate VM code.

// Context stores some useful context while compiling.
type Context struct {
	Name string // Class name being compiled
	
	sym *Symbols
	n int
}

// Unique returns a unique string suffix through out the lifetime of
// the context object. It uses a counter that increments every call.
func (c*Context) Unique() string {
	next := strconv.Itoa(c.n)
	c.n++
	return next
}


// writer a simple wrapper that stores previous errors and refuse to
// write new writes if a previous error exist. This is a hack to allow
// functions to make multiple writes and only check for error in the
// end.
type writer struct {
	io.Writer
	errored error
}

func (w *writer) WriteLine(l string) error {
	if w.errored != nil {
		return w.errored
	}
	line := []byte(l)
	line = append(line, '\n')
	_, w.errored = w.Write(line)
	return w.errored
}

func (w *writer) Writef(f string, args ...interface{}) error {
	if w.errored != nil {
		return w.errored
	}
	_, w.errored = w.Write([]byte(fmt.Sprintf(f, args...)))
	return w.errored
}


// Compile generates VM code and writes to w
func (p*Program) Compile(ctx *Context, w *writer) error {
	ctx.Name = p.Name.Literal
	w.Writef("// Class %s\n", ctx.Name)
	// compile variable declarations first
	var n int // number of variables
	for _, v := range p.Vars {
		for _, name := range v.Names {
			ctx.sym.Global(name.Literal, v.VarType.Literal, v.DecType.Literal, n)
			n++
		}
	}
	// followed by subroutine
	for _, s := range p.Subs {
		if err := s.Compile(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

func (s *SubroutineDecleration) Compile(ctx *Context, w *writer) error {
	ctx.sym.InitLocal()
	var arg0 int // during a method declaration the first argument would be this we should store symbols in the table accordingly.
	if s.SubType.Type == MethodToken {
		arg0++
	}
	// Parameters
	for i, p := range s.Params {
		ctx.sym.Local(p.Name.Literal, p.Type.Literal, "argument", i+arg0)
	}
	w.Writef("// declare %s\n", s.SubType.Literal)
	w.Writef("function %s.%s %d\n", ctx.Name, s.Name.Literal, s.SubBody.NVars())
	// If constructor allocate memory
	if s.SubType.Type == ConstructorToken {
		// Hack use the symbol table to count number of fields.
		n := ctx.sym.Count("field")
		// Allocate memory and anchor this pointer.
		w.Writef("push constant %d\n", n)
		w.WriteLine("call Memory.alloc 1") // pushed on 1 argument
		w.WriteLine("pop pointer 0")
	}
	// If this is a method set this pointer.
	if s.SubType.Type == MethodToken {
		w.WriteLine("push argument 0")
		w.WriteLine("pop pointer 0")
	}
	// Body
	if err := s.SubBody.Compile(ctx, w); err != nil {
		return fmt.Errorf("compile subroutine: %w", err)
	}
	err := w.Writef("// end %s\n", s.SubType.Literal)
	if err != nil { // check the last write error attempt
		return fmt.Errorf("compile subroutine: %w", err)
	}
	return nil
}

func (s *SubroutineBody) Compile(ctx *Context, w *writer) error {
	// Var Declaration
	var count int
	for _, vd := range s.VarDec {
		for _, n := range vd.Names {
			ctx.sym.Local(n.Literal, vd.VarType.Literal, "local", count)
			count++
		}
	}
	// Statements
	return s.Statements.Compile(ctx, w)
}

func (s *Statements) Compile(ctx *Context, w *writer) error {
	for _, stmt := range s.Statements {
		if err := stmt.Compile(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

func (ls *LetStatement) Compile(ctx *Context, w *writer) error {
	// TODO arrays
	if err := ls.Expression.Compile(ctx, w); err != nil {
		return fmt.Errorf("compile let statement: %w", err)
	}
	segment, err := ls.Name.Segment(ctx)
	if err != nil {
		return fmt.Errorf("compile let statement: %w", err)
	}

	// Set segment
	if err := w.Writef("pop %s\n", segment); err != nil {
		return fmt.Errorf("compile let statement: %w", err)
	}
	return nil;
}

func (is *IfStatement) Compile(ctx *Context, w *writer) error {
	// generate exit label.
	ifLabel := fmt.Sprintf("IF_FALSE_%s", ctx.Unique())
	var exitLabel string
	if is.ElseStatements != nil {
		exitLabel = fmt.Sprintf("IF_EXIT_%s", ctx.Unique())
	}
	w.WriteLine("// if statement")
	// first ensure to compile the expression pushing to the stack.
	if err := is.Expression.Compile(ctx, w); err != nil {
		return fmt.Errorf("compile if statement: %w", err)
	}
	// Then invert the expression to make label jumping easier.
	w.WriteLine("not")
	// Check if expression was not true and jump.
	w.Writef("if-goto %s\n", ifLabel)
	// Compile if true statements
	if err := is.Statements.Compile(ctx, w); err != nil {
		return fmt.Errorf("compile if statement: %w", err)		
	}
	// Go to if/else exit
	if is.ElseStatements != nil {
		w.Writef("goto %s\n", exitLabel)
	}
	w.Writef("label %s\n", ifLabel)
	if is.ElseStatements != nil {
		if err := is.ElseStatements.Compile(ctx, w); err != nil {
			return fmt.Errorf("compile if statement: %w", err)
		}
		w.Writef("label %s\n", exitLabel)
	}
	// check the last error
	if err := w.WriteLine("// end if"); err != nil {
		return fmt.Errorf("compile if statement: %w", err)
	}
	return nil;
}

func (ds *DoStatement) Compile(ctx *Context, w *writer) error {
	if err := ds.SubCall.Compile(ctx, w); err != nil {
		return fmt.Errorf("compile do statement: %w", err)
	}
	if err := w.WriteLine("pop temp 0"); err != nil {
		return fmt.Errorf("compile do statement: %w", err)
	}
	return nil
}

func (sc *SubroutineCall) Compile(ctx *Context, w*writer) error {
	nargs := len(sc.Expressions)
	var dest string
	// Handle calling methods
	if sc.Dest == nil || (sc.Dest != nil && sc.Dest.Type == ThisToken) {
		nargs++
		dest = ctx.Name
		w.WriteLine("push pointer 0")
	} else {
		dest = sc.Dest.Literal // default to destination unless it is a var
		v, isVar := ctx.sym.Lookup(sc.Dest.Literal)
		if isVar {
			nargs++
			dest = v.Type
			segment, err := sc.Dest.Segment(ctx)
			if err != nil {
				return fmt.Errorf("compile subroutine call: %w", err)
			}
			w.Writef("push %s\n", segment)
		}
	}
	// Push the rest of the arguments
	for _, e := range sc.Expressions {
		if err := e.Compile(ctx, w); err != nil {
			return fmt.Errorf("compile subroutine call: %w", err)
		}
	}
	// make the call
	err := w.Writef("call %s.%s %d\n", dest, sc.Name.Literal, nargs)
	if err != nil {
		return fmt.Errorf("compile subroutine call: %w", err)
	}
	return nil
}

func (rs *ReturnStatement) Compile(ctx *Context, w *writer) error {
	if rs.Expression == nil {
		w.WriteLine("push constant 0")
	} else {
		if err := rs.Expression.Compile(ctx, w); err != nil {
			return fmt.Errorf("compile return: %w", err)
		}
	}
	if err := w.WriteLine("return"); err != nil {
		return fmt.Errorf("compile return: %w", err)
	}
	return nil
}

func (ws *WhileStatement) Compile(ctx *Context, w *writer) error {
	wstart := fmt.Sprintf("WHILE_START_%s", ctx.Unique())
	wend := fmt.Sprintf("WHILE_END_%s", ctx.Unique())
	w.Writef("label %s\n", wstart)
	if err := ws.Expression.Compile(ctx, w); err != nil {
		return fmt.Errorf("compile while statement: %w", err)
	}
	w.WriteLine("not")
	w.Writef("if-goto %s\n", wend)
	if err := ws.Statements.Compile(ctx, w); err != nil {
		return fmt.Errorf("compile while statement: %w", err)
	}
	w.Writef("goto %s\n", wstart)
	if err := w.Writef("label %s\n", wend); err != nil {
		return fmt.Errorf("compile while statement: %w", err)
	}
	return nil;
}

func (e *Expression) Compile(ctx *Context, w *writer) error {
	if err := e.Term1.Compile(ctx, w); err != nil {
		return fmt.Errorf("compile expression: %w", err)
	}
	if e.Op != nil {
		// Stack machine first compile the second operation
		// then the operator.
		if err := e.Term2.Compile(ctx, w); err != nil {
			return fmt.Errorf("compile expression: %w", err)
		}
		if err := compileOp(ctx, w, e.Op); err != nil {
			return fmt.Errorf("compile expression: %w", err)	
		}
	}
	return nil
}

func (t *Term) Compile(ctx *Context, w *writer) error {
	// Compile Integer/String constants, keywords & identifiers
	if t.Token != nil {
		var err error
		switch t.Token.Type {
		case IntegerConstant:
			err = w.Writef("push constant %s\n", t.Token.Literal)
		case IdentifierToken:
			i := Identifier{*t.Token}
			segment, errs := i.Segment(ctx)
			if errs != nil {
				return fmt.Errorf("compile term: %w", err)
			}
			err = w.Writef("push %s\n", segment)
		case TrueToken:
			w.WriteLine("push constant 0")
			err = w.WriteLine("not")
		case NullToken, FalseToken:
			err = w.WriteLine("push constant 0")
		case ThisToken:
			err = w.WriteLine("push pointer 0")
		default:
			return fmt.Errorf("compile term: unknown token %s", t)
		}
		if err != nil {
			return fmt.Errorf("compile term token %s: %w", err)

		}
		return nil
	}
	// Parenthesized expression
	if t.ParenExpression != nil {
		return t.ParenExpression.Compile(ctx, w)
	}
	// Unary operators
	if t.UnaryOp != nil {
		if err := t.UnaryOperand.Compile(ctx, w); err != nil {
			return err
		}
		// unary operator uses `neg` instead of `sub`
		var err error
		switch t.UnaryOp.Type {
		case MinusToken:
			err = w.WriteLine("neg")
		case TildeToken:
			err = w.WriteLine("not")
		default:
			err = fmt.Errorf("unknown operator %s", t.UnaryOp)
		}
		return err
	}
	// Subroutine calls
	if t.SubCall != nil {
		if err := t.SubCall.Compile(ctx, w); err != nil {
			return fmt.Errorf("compile term: %w",err)
		}
		return nil
	}
	return fmt.Errorf("unsupported term compile: %#v", t)
}

func (i *Identifier) Segment(ctx *Context) (string,error) {
	v, ok := ctx.sym.Lookup(i.Token.Literal)
	if !ok {
		return "", fmt.Errorf("identifier: symbol %q not found",
			i.Token.Literal)
	}
	segment := v.Kind
	if segment == "field" {
		segment = "this"
	}
	return fmt.Sprintf("%s %d", segment, v.Index), nil
}

func compileOp(ctx *Context, w*writer, op *Token) error {
	var err error 
	switch op.Type {
	case PlusToken:
		err = w.WriteLine("add");
	case MinusToken:
		err = w.WriteLine("sub")
	case MultiplyToken:
		err = w.WriteLine("call Math.multiply 2")
	case DivideToken:
		err = w.WriteLine("call Math.divide 2")
	case AmpersandToken:
		err = w.WriteLine("and")
	case PipeToken:
		err = w.WriteLine("or")
	case GreaterThanToken:
		err = w.WriteLine("gt")
	case LessThanToken:
		err = w.WriteLine("lt")
	case EqualToken:
		err = w.WriteLine("eq")
	case TildeToken:
		err = w.WriteLine("not")
	default:
		err = fmt.Errorf("unknown operator %s", op)
	}
	if err != nil {
		return fmt.Errorf("compile term: %w", err)
	}
	return nil
}

