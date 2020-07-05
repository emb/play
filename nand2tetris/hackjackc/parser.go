package main

import (
	"fmt"
)

type Parser struct {
	lex *Lexer

	// Since Jack can have LL(2) we need to peek for the next
	// token.
	current Token
	nextTok Token
}

func NewParser(l *Lexer) *Parser {
	return &Parser{lex: l, nextTok: l.Next()}
}

func (p *Parser) Parse() (prog *Program, err error) {
	/*
		defer func() {
			if r := recover(); r != nil {
				switch v := r.(type) {
				case error:
					err = fmt.Errorf("parse: %w", v)
				default:
					err = fmt.Errorf("parse: %s", v)
				}
			}	}()
	*/
	prog, err = p.parseProgram()
	return
}

func (p *Parser) next() Token {
	p.current = p.nextTok
	p.nextTok = p.lex.Next()
	return p.current
}

func (p *Parser) peek() Token { return p.nextTok }

func (p *Parser) parseProgram() (*Program, error) {
	c := p.next()
	if c.Type != ClassToken {
		return nil, fmt.Errorf("parse program: want class got: %s", c)
	}
	i := p.next()
	if i.Type != IdentifierToken {
		return nil, fmt.Errorf("parse program: want identifier got: %s", i)
	}
	l := p.next()
	if l.Type != LeftBraceToken {
		return nil, fmt.Errorf("parse program: want { got : %s", l)
	}
	varDecs := []ClassVarDecleration{}
	if err := p.parseClassVarDecs(&varDecs); err != nil {
		return nil, fmt.Errorf("parse program: %w", err)
	}
	subDecs := []SubroutineDecleration{}
	if err := p.parseSubroutineDecs(&subDecs); err != nil {
		return nil, fmt.Errorf("parse program: %w", err)
	}
	r := p.next()
	if r.Type != RightBraceToken {
		return nil, fmt.Errorf("parse program: want } got: %s", r)
	}
	return &Program{
		Class:  c,
		LeftB:  l,
		Name:   Identifier{i},
		Vars:   varDecs,
		Subs:   subDecs,
		RightB: r,
	}, nil
}

func (p *Parser) parseClassVarDecs(vardecs *[]ClassVarDecleration) error {
	n := p.peek()
	if n.Type != StaticToken && n.Type != FieldToken {
		return nil
	}
	dt := p.next()
	vt, err := p.parseType()
	if err != nil {
		return fmt.Errorf("class vars: %w", err)
	}
	names, err := p.parseIdentifiers()
	if err != nil {
		return fmt.Errorf("class vars: %w", err)
	}
	semi := p.next()
	if semi.Type != SemiColonToken {
		return fmt.Errorf("class var: want ; got: %s", semi)
	}
	*vardecs = append(*vardecs, ClassVarDecleration{
		DecType: dt,
		VarType: vt,
		Names:   names,
		Semi:    semi,
	})
	return p.parseClassVarDecs(vardecs)
}

func (p *Parser) parseIdentifiers() ([]Identifier, error) {
	names := []Identifier{}
	for {
		v := p.next()
		if v.Type != IdentifierToken {
			return nil, fmt.Errorf("want identifier got: %s", v)
		}
		names = append(names, Identifier{v})
		if p.peek().Type == CommaToken {
			// TODO: Hack to make xml printing work. Will have to remove this later.
			names = append(names, Identifier{p.next()})
			continue
		}
		break
	}
	return names, nil
}

func (p *Parser) parseType() (Token, error) {
	t := p.next()
	switch t.Type {
	case IntToken, CharToken, BooleanToken, IdentifierToken:
		return t, nil
	}
	return Token{}, fmt.Errorf("type: want int, char, boolean or identifier got: %s", t)
}

func (p *Parser) parseSubroutineDecs(subdecs *[]SubroutineDecleration) error {
	n := p.peek()
	if n.Type != ConstructorToken && n.Type != FunctionToken && n.Type != MethodToken {
		return fmt.Errorf("subroutine: expected subroutine deceleration got: %s", n)
	}
	st := p.next()
	var rt Token
	if p.peek().Type == VoidToken {
		rt = p.next()
	} else {
		t, err := p.parseType()
		if err != nil {
			return fmt.Errorf("subroutine: return: %w", err)
		}
		rt = t
	}
	n = p.next()
	if n.Type != IdentifierToken {
		return fmt.Errorf("subroutine: want identifier got: %s", n)
	}
	lp := p.next()
	if lp.Type != LeftParenToken {
		return fmt.Errorf("subroutine: want ( got: %s", lp)
	}
	ps := Parameters{[]Token{}}
	for {
		if p.peek().Type == RightParenToken {
			break
		}
		t, err := p.parseType()
		if err != nil {
			return fmt.Errorf("subroutine: parameters: %w", err)
		}
		ps.Tokens = append(ps.Tokens, t)
		vn := p.next()
		if vn.Type != IdentifierToken {
			return fmt.Errorf("subroutine: parameters: want identifier got: %s", vn)
		}
		ps.Tokens = append(ps.Tokens, vn)
		if p.peek().Type == CommaToken {
			ps.Tokens = append(ps.Tokens, p.next())
			continue
		}
		break
	}
	rp := p.next()
	if rp.Type != RightParenToken {
		return fmt.Errorf("subroutine: want ) got: %s", rp)
	}
	body, err := p.parseSubroutineBody()
	if err != nil {
		return fmt.Errorf("subroutine: %w", err)
	}
	*subdecs = append(*subdecs, SubroutineDecleration{
		SubType: st,
		RetType: rt,
		Name:    Identifier{n},
		LeftP:   lp,
		Params:  ps,
		RightP:  rp,
		SubBody: body,
	})
	// Check if we are done.
	if p.peek().Type == RightBraceToken {
		return nil
	}
	return p.parseSubroutineDecs(subdecs)
}

func (p *Parser) parseSubroutineBody() (SubroutineBody, error) {
	lb := p.next()
	if lb.Type != LeftBraceToken {
		return SubroutineBody{}, fmt.Errorf("body: want { got: %s", lb)
	}
	varDecs := []VarDeclaration{}
	if err := p.parseVarDeclaration(&varDecs); err != nil {
		return SubroutineBody{}, fmt.Errorf("body: %w", err)
	}
	statements := Statements{[]interface{}{}}
	if err := p.parseStatements(&statements); err != nil {
		return SubroutineBody{}, fmt.Errorf("body: %w", err)
	}
	rb := p.next()
	if rb.Type != RightBraceToken {
		return SubroutineBody{}, fmt.Errorf("body: want } got: %s", rb)
	}
	return SubroutineBody{
		LeftB:      lb,
		VarDec:     varDecs,
		Statements: statements,
		RightB:     rb,
	}, nil
}

func (p *Parser) parseVarDeclaration(vardecs *[]VarDeclaration) error {
	if p.peek().Type != VarToken {
		return nil
	}
	v := p.next()
	t, err := p.parseType()
	if err != nil {
		return fmt.Errorf("var declaration: %w", err)
	}
	names, err := p.parseIdentifiers()
	if err != nil {
		return fmt.Errorf("var declaration: %w", err)
	}
	semi := p.next()
	if semi.Type != SemiColonToken {
		return fmt.Errorf("var declaration: want ; got: %s", semi)
	}
	*vardecs = append(*vardecs, VarDeclaration{
		Var:     v,
		VarType: t,
		Names:   names,
		Semi:    semi,
	})
	return p.parseVarDeclaration(vardecs)
}

func (p *Parser) parseStatements(stmts *Statements) error {
	tok := p.peek()
	tt := tok.Type
	if tt != LetToken && tt != IfToken && tt != WhileToken && tt != DoToken && tt != ReturnToken {
		return fmt.Errorf("statements: want statement got: %s", tok)
	}
	var stmt interface{}
	var err error
	switch tok.Type {
	case LetToken:
		stmt, err = p.parseLetStatement()
	case IfToken:
		stmt, err = p.parseIfStatement()
	case WhileToken:
		stmt, err = p.parseWhileStatement()
	case DoToken:
		stmt, err = p.parseDoStatement()
	case ReturnToken:
		stmt, err = p.parseReturn()
	}
	if err != nil {
		return fmt.Errorf("statements: %w", err)
	}
	stmts.Statements = append(stmts.Statements, stmt)
	if p.peek().Type == RightBraceToken { // finished all statements
		return nil
	}
	return p.parseStatements(stmts)
}

func (p *Parser) parseLetStatement() (LetStatement, error) {
	var ls LetStatement
	ls.Let = p.next()
	if ls.Let.Type != LetToken {
		return LetStatement{}, fmt.Errorf("let statement: want let got: %s", ls.Let)
	}
	ls.Name = Identifier{p.next()}
	if ls.Name.Type != IdentifierToken {
		return LetStatement{}, fmt.Errorf("let statement: want identifier got: %s", ls.Name)
	}
	if p.peek().Type == LeftBraketToken {
		lb := p.next()
		ls.LeftB = &lb
		exp, err := p.parseExpression()
		if err != nil {
			return LetStatement{}, fmt.Errorf("let statement: %w", err)
		}
		ls.ArrExpression = &exp
		rb := p.next()
		if rb.Type != RightBraketToken {
			return LetStatement{}, fmt.Errorf("let statement: want ] got: %s", rb)
		}
		ls.RightB = &rb
	}
	ls.Equal = p.next()
	if ls.Equal.Type != EqualToken {
		return LetStatement{}, fmt.Errorf("let statement: want = got: %s", ls.Equal)
	}
	exp, err := p.parseExpression()
	if err != nil {
		return LetStatement{}, fmt.Errorf("let statement: %w", err)
	}
	ls.Expression = exp
	ls.Semi = p.next()
	if ls.Semi.Type != SemiColonToken {
		return LetStatement{}, fmt.Errorf("let statement: want ; got: %s", ls.Semi)
	}
	return ls, nil
}

func (p *Parser) parseIfStatement() (IfStatement, error) {
	var is IfStatement
	is.If = p.next()
	if is.If.Type != IfToken {
		return IfStatement{}, fmt.Errorf("if statement: want if got: %s", is.If)
	}
	is.LeftP = p.next()
	if is.LeftP.Type != LeftParenToken {
		return IfStatement{}, fmt.Errorf("if statement: want ( got: %s", is.LeftP)
	}
	exp, err := p.parseExpression()
	if err != nil {
		return IfStatement{}, fmt.Errorf("if statement: %w", err)
	}
	is.Expression = exp
	is.RightP = p.next()
	if is.RightP.Type != RightParenToken {
		return IfStatement{}, fmt.Errorf("if statement: want ) got: %s", is.RightP)
	}
	is.LeftB = p.next()
	if is.LeftB.Type != LeftBraceToken {
		return IfStatement{}, fmt.Errorf("if statement: want { got: %s", is.LeftB)
	}
	statements := Statements{[]interface{}{}}
	if err := p.parseStatements(&statements); err != nil {
		return IfStatement{}, fmt.Errorf("if statement: %w", err)
	}
	is.Statements = statements
	is.RightB = p.next()
	if is.RightB.Type != RightBraceToken {
		return IfStatement{}, fmt.Errorf("if statement: want } got: %s", is.RightB)
	}
	if p.peek().Type == ElseToken {
		e := p.next()
		is.Else = &e
		lb := p.next()
		if lb.Type != LeftBraceToken {
			return IfStatement{}, fmt.Errorf("if statement: else: want { got: %s", lb)
		}
		is.ElseLeftB = &lb
		statements := Statements{[]interface{}{}}
		if err := p.parseStatements(&statements); err != nil {
			return IfStatement{}, fmt.Errorf("if statement: else: %w", err)
		}
		is.ElseStatements = &statements
		rb := p.next()
		if rb.Type != RightBraceToken {
			return IfStatement{}, fmt.Errorf("if statement: else: want } got: %s", rb)
		}
		is.ElseRightB = &rb
	}
	return is, nil
}

func (p *Parser) parseWhileStatement() (WhileStatement, error) {
	var ws WhileStatement
	ws.While = p.next()
	if ws.While.Type != WhileToken {
		return WhileStatement{}, fmt.Errorf("while statement: want while got: %s", ws.While)
	}
	ws.LeftP = p.next()
	if ws.LeftP.Type != LeftParenToken {
		return WhileStatement{}, fmt.Errorf("while statement: want ( got: %s", ws.LeftP)
	}
	exp, err := p.parseExpression()
	if err != nil {
		return WhileStatement{}, fmt.Errorf("while statement: %w", err)
	}
	ws.Expression = exp
	ws.RightP = p.next()
	if ws.RightP.Type != RightParenToken {
		return WhileStatement{}, fmt.Errorf("while statement: want ) got: %s", ws.RightP)
	}
	ws.LeftB = p.next()
	if ws.LeftB.Type != LeftBraceToken {
		return WhileStatement{}, fmt.Errorf("while statement: want { got: %s", ws.LeftB)
	}
	statements := Statements{[]interface{}{}}
	if err := p.parseStatements(&statements); err != nil {
		return WhileStatement{}, fmt.Errorf("while statement: %w", err)
	}
	ws.Statements = statements
	ws.RightB = p.next()
	if ws.RightB.Type != RightBraceToken {
		return WhileStatement{}, fmt.Errorf("if statement: want } got: %s", ws.RightB)
	}
	return ws, nil
}

func (p *Parser) parseDoStatement() (DoStatement, error) {
	var ds DoStatement
	ds.Do = p.next()
	if ds.Do.Type != DoToken {
		return DoStatement{}, fmt.Errorf("do statement: want do got: %s", ds.Do)
	}
	sc, err := p.parseSubroutineCall(p.next())
	if err != nil {
		return DoStatement{}, fmt.Errorf("do statement: %w", err)
	}
	ds.SubroutineCall = sc
	ds.Semi = p.next()
	if ds.Semi.Type != SemiColonToken {
		return DoStatement{}, fmt.Errorf("do statement: want ; got: %s", ds.Semi)
	}
	return ds, nil
}

func (p *Parser) parseSubroutineCall(tok Token) (SubroutineCall, error) {
	var sc SubroutineCall
	sc.Name = Identifier{tok}
	if sc.Name.Type != IdentifierToken {
		return SubroutineCall{}, fmt.Errorf("subroutine call: want identifier got: %s", sc.Name)
	}
	if p.peek().Type == DotToken {
		d := p.next()
		sc.Dot = &d
		cn := sc.Name
		sc.Dest = &cn
		sc.Name = Identifier{p.next()}
		if sc.Name.Type != IdentifierToken {
			return SubroutineCall{}, fmt.Errorf("subroutine call: want identifier got: %s", sc.Name)
		}

	}
	sc.LeftP = p.next()
	if sc.LeftP.Type != LeftParenToken {
		return SubroutineCall{}, fmt.Errorf("subroutine call: want ( got: %s", sc.LeftP)
	}
	el, err := p.parseExpressionList()
	if err != nil {
		return SubroutineCall{}, fmt.Errorf("subroutine call: %w", err)
	}
	sc.Expressions = &el
	sc.RightP = p.next()
	if sc.RightP.Type != RightParenToken {
		return SubroutineCall{}, fmt.Errorf("subroutine call: want ) got: %s", sc.RightP)
	}
	return sc, nil
}

func (p *Parser) parseExpressionList() (ExpressionList, error) {
	if p.peek().Type == RightParenToken {
		return ExpressionList{}, nil
	}
	el := ExpressionList{Expressions: []interface{}{}}
	for {
		e, err := p.parseExpression()
		if err != nil {
			return ExpressionList{}, fmt.Errorf("expression list: %w", err)
		}
		el.Expressions = append(el.Expressions, e)
		if p.peek().Type == CommaToken {
			el.Expressions = append(el.Expressions, p.next())
			continue
		}
		break
	}
	return el, nil
}

func (p *Parser) parseReturn() (ReturnStatement, error) {
	var rs ReturnStatement
	rs.Return = p.next()
	if rs.Return.Type != ReturnToken {
		return ReturnStatement{}, fmt.Errorf("return statement: want return got: %s", rs.Return)
	}
	if p.peek().Type != SemiColonToken {
		e, err := p.parseExpression()
		if err != nil {
			return ReturnStatement{}, fmt.Errorf("return statement: %w", err)
		}
		rs.Expression = &e
	}
	rs.Semi = p.next()
	if rs.Semi.Type != SemiColonToken {
		return ReturnStatement{}, fmt.Errorf("return statement: want ; got: %s", rs.Semi)
	}
	return rs, nil
}

var (
	ops = []TokenType{
		PlusToken, MinusToken, MultiplyToken, DivideToken,
		AmpersandToken, PipeToken,
		GreaterThanToken, LessThanToken, EqualToken,
	}
	unaryOps         = []TokenType{MinusToken, TildeToken}
	keywordConstants = []TokenType{
		TrueToken, FalseToken,
		ThisToken, NullToken,
	}
)

func tokenIn(t Token, ts []TokenType) bool {
	for _, tt := range ts {
		if t.Type == tt {
			return true
		}
	}
	return false
}

func (p *Parser) parseExpression() (Expression, error) {
	t, err := p.parseTerm()
	if err != nil {
		return Expression{}, fmt.Errorf("expression: %w", err)
	}
	expr := Expression{Term1: t}
	if tokenIn(p.peek(), ops) {
		op := p.next()
		expr.Op = &op
		t1, err := p.parseTerm()
		if err != nil {
			return Expression{}, fmt.Errorf("expression: %w", err)
		}
		expr.Term2 = &t1
	}
	return expr, nil
}

func (p *Parser) parseTerm() (Term, error) {
	tok := p.peek()
	if tok.Type == IntegerConstant ||
		tok.Type == StringConstant ||
		tokenIn(tok, keywordConstants) {
		p.next() // ignore but make sure we advance
		return Term{Token: &tok}, nil
	}
	if tok.Type == LeftParenToken {
		lp := p.next()
		t := Term{}
		t.ParenLeftP = &lp
		e, err := p.parseExpression()
		if err != nil {
			return Term{}, fmt.Errorf("term: %w", err)
		}
		t.ParenExpression = &e
		rp := p.next()
		if rp.Type != RightParenToken {
			return Term{}, fmt.Errorf("term: want ) got: %s", rp)
		}
		t.ParenRightP = &rp
		return t, nil
	}
	if tokenIn(tok, unaryOps) {
		op := p.next()
		t, err := p.parseTerm()
		if err != nil {
			return Term{}, fmt.Errorf("term: %w", err)
		}
		return Term{UnaryOp: &op, UnaryOperand: &t}, nil
	}
	if tok.Type == IdentifierToken {
		ident := p.next()
		if p.peek().Type == LeftBraketToken {
			lb := p.next()
			e, err := p.parseExpression()
			if err != nil {
				return Term{}, fmt.Errorf("term: array expression: %w", err)
			}
			rb := p.next()
			if rb.Type != RightBraketToken {
				return Term{}, fmt.Errorf("term: array expression want ] got: %s", rb)
			}
			return Term{
				ArrayIdent:  &ident,
				ArrayLeftB:  &lb,
				ArrayExpr:   &e,
				ArrayRightB: &rb,
			}, nil
		}
		if p.peek().Type == DotToken ||
			p.peek().Type == LeftParenToken {
			sc, err := p.parseSubroutineCall(ident)
			if err != nil {
				return Term{}, fmt.Errorf("term: %s", err)
			}
			return Term{SubroutineCall: sc}, nil
		}
		return Term{Token: &ident}, nil
	}
	return Term{}, fmt.Errorf("term: unknown term starting with token: %s", tok)
}
