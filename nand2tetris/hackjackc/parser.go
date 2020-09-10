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
		Name:   Identifier{i},
		Vars:   varDecs,
		Subs:   subDecs,
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
			p.next() // skip comma
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
	ps := []Parameter{}
	for {
		if p.peek().Type == RightParenToken {
			break
		}
		t, err := p.parseType()
		if err != nil {
			return fmt.Errorf("subroutine: parameters: %w", err)
		}
		vn := p.next()
		if vn.Type != IdentifierToken {
			return fmt.Errorf("subroutine: parameters: want identifier got: %s", vn)
		}
		ps = append(ps, Parameter{Name: Identifier{vn}, Type:t})
		if p.peek().Type == CommaToken {
			p.next()
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
		Params:  ps,
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
	statements := Statements{[]Compiler{}}
	if err := p.parseStatements(&statements); err != nil {
		return SubroutineBody{}, fmt.Errorf("body: %w", err)
	}
	rb := p.next()
	if rb.Type != RightBraceToken {
		return SubroutineBody{}, fmt.Errorf("body: want } got: %s", rb)
	}
	return SubroutineBody{
		VarDec:     varDecs,
		Statements: statements,
	}, nil
}

func (p *Parser) parseVarDeclaration(vardecs *[]VarDeclaration) error {
	if p.peek().Type != VarToken {
		return nil
	}
	p.next() // var
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
		VarType: t,
		Names:   names,
	})
	return p.parseVarDeclaration(vardecs)
}

func (p *Parser) parseStatements(stmts *Statements) error {
	tok := p.peek()
	tt := tok.Type
	if tt != LetToken && tt != IfToken && tt != WhileToken && tt != DoToken && tt != ReturnToken {
		return fmt.Errorf("statements: want statement got: %s", tok)
	}
	var stmt Compiler
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

func (p *Parser) parseLetStatement() (*LetStatement, error) {
	var ls LetStatement
	lsLet := p.next()
	if lsLet.Type != LetToken {
		return nil, fmt.Errorf("let statement: want let got: %s", lsLet)
	}
	ls.Name = Identifier{p.next()}
	if ls.Name.Type != IdentifierToken {
		return nil, fmt.Errorf("let statement: want identifier got: %s", ls.Name)
	}
	if p.peek().Type == LeftBraketToken {
		p.next() // Left braket 
		exp, err := p.parseExpression()
		if err != nil {
			return nil, fmt.Errorf("let statement: %w", err)
		}
		ls.ArrExpression = &exp
		rb := p.next()
		if rb.Type != RightBraketToken {
			return nil, fmt.Errorf("let statement: want ] got: %s", rb)
		}
	}
	eq := p.next()
	if eq.Type != EqualToken {
		return nil, fmt.Errorf("let statement: want = got: %s", eq)
	}
	exp, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("let statement: %w", err)
	}
	ls.Expression = exp
	semi := p.next()
	if semi.Type != SemiColonToken {
		return nil, fmt.Errorf("let statement: want ; got: %s", semi)
	}
	return &ls, nil
}

func (p *Parser) parseIfStatement() (*IfStatement, error) {
	var is IfStatement
	isIf := p.next()
	if isIf.Type != IfToken {
		return nil, fmt.Errorf("if statement: want if got: %s", isIf)
	}
	isLeftP := p.next()
	if isLeftP.Type != LeftParenToken {
		return nil, fmt.Errorf("if statement: want ( got: %s", isLeftP)
	}
	exp, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("if statement: %w", err)
	}
	is.Expression = exp
	isRightP := p.next()
	if isRightP.Type != RightParenToken {
		return nil, fmt.Errorf("if statement: want ) got: %s", isRightP)
	}
	isLeftB := p.next()
	if isLeftB.Type != LeftBraceToken {
		return nil, fmt.Errorf("if statement: want { got: %s", isLeftB)
	}
	statements := Statements{[]Compiler{}}
	if err := p.parseStatements(&statements); err != nil {
		return nil, fmt.Errorf("if statement: %w", err)
	}
	is.Statements = statements
	isRightB := p.next()
	if isRightB.Type != RightBraceToken {
		return nil, fmt.Errorf("if statement: want } got: %s", isRightB)
	}
	if p.peek().Type == ElseToken {
		p.next() // 
		lb := p.next()
		if lb.Type != LeftBraceToken {
			return nil, fmt.Errorf("if statement: else: want { got: %s", lb)
		}
		statements := Statements{[]Compiler{}}
		if err := p.parseStatements(&statements); err != nil {
			return nil, fmt.Errorf("if statement: else: %w", err)
		}
		is.ElseStatements = &statements
		rb := p.next()
		if rb.Type != RightBraceToken {
			return nil, fmt.Errorf("if statement: else: want } got: %s", rb)
		}
	}
	return &is, nil
}

func (p *Parser) parseWhileStatement() (*WhileStatement, error) {
	var ws WhileStatement
	wsWhile := p.next()
	if wsWhile.Type != WhileToken {
		return nil, fmt.Errorf("while statement: want while got: %s", wsWhile)
	}
	wsLeftP := p.next()
	if wsLeftP.Type != LeftParenToken {
		return nil, fmt.Errorf("while statement: want ( got: %s", wsLeftP)
	}
	exp, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("while statement: %w", err)
	}
	ws.Expression = exp
	wsRightP := p.next()
	if wsRightP.Type != RightParenToken {
		return nil, fmt.Errorf("while statement: want ) got: %s", wsRightP)
	}
	wsLeftB := p.next()
	if wsLeftB.Type != LeftBraceToken {
		return nil, fmt.Errorf("while statement: want { got: %s", wsLeftB)
	}
	statements := Statements{[]Compiler{}}
	if err := p.parseStatements(&statements); err != nil {
		return nil, fmt.Errorf("while statement: %w", err)
	}
	ws.Statements = statements
	wsRightB := p.next()
	if wsRightB.Type != RightBraceToken {
		return nil, fmt.Errorf("if statement: want } got: %s", wsRightB)
	}
	return &ws, nil
}

func (p *Parser) parseDoStatement() (*DoStatement, error) {
	var ds DoStatement
	dsDo := p.next()
	if dsDo.Type != DoToken {
		return nil, fmt.Errorf("do statement: want do got: %s", dsDo)
	}
	sc, err := p.parseSubroutineCall(p.next())
	if err != nil {
		return nil, fmt.Errorf("do statement: %w", err)
	}
	ds.SubCall = sc
	dsSemi := p.next()
	if dsSemi.Type != SemiColonToken {
		return nil, fmt.Errorf("do statement: want ; got: %s", dsSemi)
	}
	return &ds, nil
}

func (p *Parser) parseSubroutineCall(tok Token) (SubroutineCall, error) {
	var sc SubroutineCall
	sc.Name = Identifier{tok}
	if sc.Name.Type != IdentifierToken {
		return SubroutineCall{}, fmt.Errorf("subroutine call: want identifier got: %s", sc.Name)
	}
	if p.peek().Type == DotToken {
		p.next() // Dot.
		cn := sc.Name
		sc.Dest = &cn
		sc.Name = Identifier{p.next()}
		if sc.Name.Type != IdentifierToken {
			return SubroutineCall{}, fmt.Errorf("subroutine call: want identifier got: %s", sc.Name)
		}

	}
	scLeftP := p.next()
	if scLeftP.Type != LeftParenToken {
		return SubroutineCall{}, fmt.Errorf("subroutine call: want ( got: %s", scLeftP)
	}
	el, err := p.parseExpressionList()
	if err != nil {
		return SubroutineCall{}, fmt.Errorf("subroutine call: %w", err)
	}
	sc.Expressions = el
	scRightP := p.next()
	if scRightP.Type != RightParenToken {
		return SubroutineCall{}, fmt.Errorf("subroutine call: want ) got: %s", scRightP)
	}
	return sc, nil
}

func (p *Parser) parseExpressionList() ([]Expression, error) {
	if p.peek().Type == RightParenToken {
		return nil, nil
	}
	el := []Expression{}
	for {
		e, err := p.parseExpression()
		if err != nil {
			return nil, fmt.Errorf("expression list: %w", err)
		}
		el = append(el, e)
		if p.peek().Type == CommaToken {
			p.next() // skip comma
			continue
		}
		break
	}
	return el, nil
}

func (p *Parser) parseReturn() (*ReturnStatement, error) {
	var rs ReturnStatement
	rsReturn := p.next()
	if rsReturn.Type != ReturnToken {
		return nil, fmt.Errorf("return statement: want return got: %s", rsReturn)
	}
	if p.peek().Type != SemiColonToken {
		e, err := p.parseExpression()
		if err != nil {
			return nil, fmt.Errorf("return statement: %w", err)
		}
		rs.Expression = &e
	}
	rsSemi := p.next()
	if rsSemi.Type != SemiColonToken {
		return nil, fmt.Errorf("return statement: want ; got: %s", rsSemi)
	}
	return &rs, nil
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
		p.next() // skip parenthesis
		e, err := p.parseExpression()
		if err != nil {
			return Term{}, fmt.Errorf("term: %w", err)
		}
		rp := p.next()
		if rp.Type != RightParenToken {
			return Term{}, fmt.Errorf("term: want ) got: %s", rp)
		}
		return Term{ParenExpression: &e}, nil
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
			p.next() // left braket
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
				ArrayExpr: &e,
			}, nil
		}
		if p.peek().Type == DotToken ||
			p.peek().Type == LeftParenToken {
			sc, err := p.parseSubroutineCall(ident)
			if err != nil {
				return Term{}, fmt.Errorf("term: %s", err)
			}
			return Term{SubCall: &sc}, nil
		}
		return Term{Token: &ident}, nil
	}
	return Term{}, fmt.Errorf("term: unknown term starting with token: %s", tok)
}
