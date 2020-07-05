package main

import (
	"fmt"

	"hackjackc/token"
)

type Parser struct {
	lex *Lexer

	// Since Jack can have LL(2) we need to peek for the next
	// token.
	current token.Token
	nextTok token.Token
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

func (p *Parser) next() token.Token {
	p.current = p.nextTok
	p.nextTok = p.lex.Next()
	return p.current
}

func (p *Parser) peek() token.Token { return p.nextTok }

func (p *Parser) parseProgram() (*Program, error) {
	c := p.next()
	if c.Type != token.Class {
		return nil, fmt.Errorf("parse program: want class got: %s", c)
	}
	i := p.next()
	if i.Type != token.Identifier {
		return nil, fmt.Errorf("parse program: want identifier got: %s", i)
	}
	l := p.next()
	if l.Type != token.LeftBrace {
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
	if r.Type != token.RightBrace {
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
	if n.Type != token.Static && n.Type != token.Field {
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
	if semi.Type != token.SemiColon {
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
		if v.Type != token.Identifier {
			return nil, fmt.Errorf("want identifier got: %s", v)
		}
		names = append(names, Identifier{v})
		if p.peek().Type == token.Comma {
			// TODO: Hack to make xml printing work. Will have to remove this later.
			names = append(names, Identifier{p.next()})
			continue
		}
		break
	}
	return names, nil
}

func (p *Parser) parseType() (token.Token, error) {
	t := p.next()
	switch t.Type {
	case token.Int, token.Char, token.Boolean, token.Identifier:
		return t, nil
	}
	return token.Token{}, fmt.Errorf("type: want int, char, boolean or identifier got: %s", t)
}

func (p *Parser) parseSubroutineDecs(subdecs *[]SubroutineDecleration) error {
	n := p.peek()
	if n.Type != token.Constructor && n.Type != token.Function && n.Type != token.Method {
		return fmt.Errorf("subroutine: expected subroutine deceleration got: %s", n)
	}
	st := p.next()
	var rt token.Token
	if p.peek().Type == token.Void {
		rt = p.next()
	} else {
		t, err := p.parseType()
		if err != nil {
			return fmt.Errorf("subroutine: return: %w", err)
		}
		rt = t
	}
	n = p.next()
	if n.Type != token.Identifier {
		return fmt.Errorf("subroutine: want identifier got: %s", n)
	}
	lp := p.next()
	if lp.Type != token.LeftParen {
		return fmt.Errorf("subroutine: want ( got: %s", lp)
	}
	ps := Parameters{[]token.Token{}}
	for {
		if p.peek().Type == token.RightParen {
			break
		}
		t, err := p.parseType()
		if err != nil {
			return fmt.Errorf("subroutine: parameters: %w", err)
		}
		ps.Tokens = append(ps.Tokens, t)
		vn := p.next()
		if vn.Type != token.Identifier {
			return fmt.Errorf("subroutine: parameters: want identifier got: %s", vn)
		}
		ps.Tokens = append(ps.Tokens, vn)
		if p.peek().Type == token.Comma {
			ps.Tokens = append(ps.Tokens, p.next())
			continue
		}
		break
	}
	rp := p.next()
	if rp.Type != token.RightParen {
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
	if p.peek().Type == token.RightBrace {
		return nil
	}
	return p.parseSubroutineDecs(subdecs)
}

func (p *Parser) parseSubroutineBody() (SubroutineBody, error) {
	lb := p.next()
	if lb.Type != token.LeftBrace {
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
	if rb.Type != token.RightBrace {
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
	if p.peek().Type != token.Var {
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
	if semi.Type != token.SemiColon {
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
	if tt != token.Let && tt != token.If && tt != token.While && tt != token.Do && tt != token.Return {
		return fmt.Errorf("statements: want statement got: %s", tok)
	}
	var stmt interface{}
	var err error
	switch tok.Type {
	case token.Let:
		stmt, err = p.parseLetStatement()
	case token.If:
		stmt, err = p.parseIfStatement()
	case token.While:
		stmt, err = p.parseWhileStatement()
	case token.Do:
		stmt, err = p.parseDoStatement()
	case token.Return:
		stmt, err = p.parseReturn()
	}
	if err != nil {
		return fmt.Errorf("statements: %w", err)
	}
	stmts.Statements = append(stmts.Statements, stmt)
	if p.peek().Type == token.RightBrace { // finished all statements
		return nil
	}
	return p.parseStatements(stmts)
}

func (p *Parser) parseLetStatement() (LetStatement, error) {
	var ls LetStatement
	ls.Let = p.next()
	if ls.Let.Type != token.Let {
		return LetStatement{}, fmt.Errorf("let statement: want let got: %s", ls.Let)
	}
	ls.Name = Identifier{p.next()}
	if ls.Name.Type != token.Identifier {
		return LetStatement{}, fmt.Errorf("let statement: want identifier got: %s", ls.Name)
	}
	if p.peek().Type == token.LeftBraket {
		lb := p.next()
		ls.LeftB = &lb
		exp, err := p.parseExpression()
		if err != nil {
			return LetStatement{}, fmt.Errorf("let statement: %w", err)
		}
		ls.ArrExpression = &exp
		rb := p.next()
		if rb.Type != token.RightBraket {
			return LetStatement{}, fmt.Errorf("let statement: want ] got: %s", rb)
		}
		ls.RightB = &rb
	}
	ls.Equal = p.next()
	if ls.Equal.Type != token.Equal {
		return LetStatement{}, fmt.Errorf("let statement: want = got: %s", ls.Equal)
	}
	exp, err := p.parseExpression()
	if err != nil {
		return LetStatement{}, fmt.Errorf("let statement: %w", err)
	}
	ls.Expression = exp
	ls.Semi = p.next()
	if ls.Semi.Type != token.SemiColon {
		return LetStatement{}, fmt.Errorf("let statement: want ; got: %s", ls.Semi)
	}
	return ls, nil
}

func (p *Parser) parseIfStatement() (IfStatement, error) {
	var is IfStatement
	is.If = p.next()
	if is.If.Type != token.If {
		return IfStatement{}, fmt.Errorf("if statement: want if got: %s", is.If)
	}
	is.LeftP = p.next()
	if is.LeftP.Type != token.LeftParen {
		return IfStatement{}, fmt.Errorf("if statement: want ( got: %s", is.LeftP)
	}
	exp, err := p.parseExpression()
	if err != nil {
		return IfStatement{}, fmt.Errorf("if statement: %w", err)
	}
	is.Expression = exp
	is.RightP = p.next()
	if is.RightP.Type != token.RightParen {
		return IfStatement{}, fmt.Errorf("if statement: want ) got: %s", is.RightP)
	}
	is.LeftB = p.next()
	if is.LeftB.Type != token.LeftBrace {
		return IfStatement{}, fmt.Errorf("if statement: want { got: %s", is.LeftB)
	}
	statements := Statements{[]interface{}{}}
	if err := p.parseStatements(&statements); err != nil {
		return IfStatement{}, fmt.Errorf("if statement: %w", err)
	}
	is.Statements = statements
	is.RightB = p.next()
	if is.RightB.Type != token.RightBrace {
		return IfStatement{}, fmt.Errorf("if statement: want } got: %s", is.RightB)
	}
	if p.peek().Type == token.Else {
		e := p.next()
		is.Else = &e
		lb := p.next()
		if lb.Type != token.LeftBrace {
			return IfStatement{}, fmt.Errorf("if statement: else: want { got: %s", lb)
		}
		is.ElseLeftB = &lb
		statements := Statements{[]interface{}{}}
		if err := p.parseStatements(&statements); err != nil {
			return IfStatement{}, fmt.Errorf("if statement: else: %w", err)
		}
		is.ElseStatements = &statements
		rb := p.next()
		if rb.Type != token.RightBrace {
			return IfStatement{}, fmt.Errorf("if statement: else: want } got: %s", rb)
		}
		is.ElseRightB = &rb
	}
	return is, nil
}

func (p *Parser) parseWhileStatement() (WhileStatement, error) {
	var ws WhileStatement
	ws.While = p.next()
	if ws.While.Type != token.While {
		return WhileStatement{}, fmt.Errorf("while statement: want while got: %s", ws.While)
	}
	ws.LeftP = p.next()
	if ws.LeftP.Type != token.LeftParen {
		return WhileStatement{}, fmt.Errorf("while statement: want ( got: %s", ws.LeftP)
	}
	exp, err := p.parseExpression()
	if err != nil {
		return WhileStatement{}, fmt.Errorf("while statement: %w", err)
	}
	ws.Expression = exp
	ws.RightP = p.next()
	if ws.RightP.Type != token.RightParen {
		return WhileStatement{}, fmt.Errorf("while statement: want ) got: %s", ws.RightP)
	}
	ws.LeftB = p.next()
	if ws.LeftB.Type != token.LeftBrace {
		return WhileStatement{}, fmt.Errorf("while statement: want { got: %s", ws.LeftB)
	}
	statements := Statements{[]interface{}{}}
	if err := p.parseStatements(&statements); err != nil {
		return WhileStatement{}, fmt.Errorf("while statement: %w", err)
	}
	ws.Statements = statements
	ws.RightB = p.next()
	if ws.RightB.Type != token.RightBrace {
		return WhileStatement{}, fmt.Errorf("if statement: want } got: %s", ws.RightB)
	}
	return ws, nil
}

func (p *Parser) parseDoStatement() (DoStatement, error) {
	var ds DoStatement
	ds.Do = p.next()
	if ds.Do.Type != token.Do {
		return DoStatement{}, fmt.Errorf("do statement: want do got: %s", ds.Do)
	}
	sc, err := p.parseSubroutineCall(p.next())
	if err != nil {
		return DoStatement{}, fmt.Errorf("do statement: %w", err)
	}
	ds.SubroutineCall = sc
	ds.Semi = p.next()
	if ds.Semi.Type != token.SemiColon {
		return DoStatement{}, fmt.Errorf("do statement: want ; got: %s", ds.Semi)
	}
	return ds, nil
}

func (p *Parser) parseSubroutineCall(tok token.Token) (SubroutineCall, error) {
	var sc SubroutineCall
	sc.Name = Identifier{tok}
	if sc.Name.Type != token.Identifier {
		return SubroutineCall{}, fmt.Errorf("subroutine call: want identifier got: %s", sc.Name)
	}
	if p.peek().Type == token.Dot {
		d := p.next()
		sc.Dot = &d
		cn := sc.Name
		sc.Dest = &cn
		sc.Name = Identifier{p.next()}
		if sc.Name.Type != token.Identifier {
			return SubroutineCall{}, fmt.Errorf("subroutine call: want identifier got: %s", sc.Name)
		}

	}
	sc.LeftP = p.next()
	if sc.LeftP.Type != token.LeftParen {
		return SubroutineCall{}, fmt.Errorf("subroutine call: want ( got: %s", sc.LeftP)
	}
	el, err := p.parseExpressionList()
	if err != nil {
		return SubroutineCall{}, fmt.Errorf("subroutine call: %w", err)
	}
	sc.Expressions = &el
	sc.RightP = p.next()
	if sc.RightP.Type != token.RightParen {
		return SubroutineCall{}, fmt.Errorf("subroutine call: want ) got: %s", sc.RightP)
	}
	return sc, nil
}

func (p *Parser) parseExpressionList() (ExpressionList, error) {
	if p.peek().Type == token.RightParen {
		return ExpressionList{}, nil
	}
	el := ExpressionList{Expressions: []interface{}{}}
	for {
		e, err := p.parseExpression()
		if err != nil {
			return ExpressionList{}, fmt.Errorf("expression list: %w", err)
		}
		el.Expressions = append(el.Expressions, e)
		if p.peek().Type == token.Comma {
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
	if rs.Return.Type != token.Return {
		return ReturnStatement{}, fmt.Errorf("return statement: want return got: %w", rs.Return)
	}
	if p.peek().Type != token.SemiColon {
		e, err := p.parseExpression()
		if err != nil {
			return ReturnStatement{}, fmt.Errorf("return statement: %w", err)
		}
		rs.Expression = &e
	}
	rs.Semi = p.next()
	if rs.Semi.Type != token.SemiColon {
		return ReturnStatement{}, fmt.Errorf("return statement: want ; got: %s", rs.Semi)
	}
	return rs, nil
}

var (
	ops = []token.TokenType{
		token.Plus, token.Minus, token.Multiply, token.Divide,
		token.Ampersand, token.Pipe,
		token.GreaterThan, token.LessThan, token.Equal,
	}
	unaryOps         = []token.TokenType{token.Minus, token.Tilde}
	keywordConstants = []token.TokenType{
		token.True, token.False,
		token.This, token.Null,
	}
)

func tokenIn(t token.Token, ts []token.TokenType) bool {
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
	if tok.Type == token.IntegerConstant ||
		tok.Type == token.StringConstant ||
		tokenIn(tok, keywordConstants) {
		p.next() // ignore but make sure we advance
		return Term{Token: &tok}, nil
	}
	if tok.Type == token.LeftParen {
		lp := p.next()
		t := Term{}
		t.ParenLeftP = &lp
		e, err := p.parseExpression()
		if err != nil {
			return Term{}, fmt.Errorf("term: %w", err)
		}
		t.ParenExpression = &e
		rp := p.next()
		if rp.Type != token.RightParen {
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
	if tok.Type == token.Identifier {
		ident := p.next()
		if p.peek().Type == token.LeftBraket {
			lb := p.next()
			e, err := p.parseExpression()
			if err != nil {
				return Term{}, fmt.Errorf("term: array expression: %w", err)
			}
			rb := p.next()
			if rb.Type != token.RightBraket {
				return Term{}, fmt.Errorf("term: array expression want ] got: %s", rb)
			}
			return Term{
				ArrayIdent:  &ident,
				ArrayLeftB:  &lb,
				ArrayExpr:   &e,
				ArrayRightB: &rb,
			}, nil
		}
		if p.peek().Type == token.Dot ||
			p.peek().Type == token.LeftParen {
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
