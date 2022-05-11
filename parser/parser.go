package parser

import (
	"github.com/yzbmz5913/stang/ast"
	"github.com/yzbmz5913/stang/lexer"
	"github.com/yzbmz5913/stang/token"
)

const (
	_ int = iota
	LOWEST
	ASSIGN
	OR
	AND
	EQUALS
	LESSGREATER
	SLICE
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
	INCRDECR
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:     ASSIGN,
	token.AND:        AND,
	token.OR:         OR,
	token.EQ:         EQUALS,
	token.NEQ:        EQUALS,
	token.LT:         LESSGREATER,
	token.LTE:        LESSGREATER,
	token.GT:         LESSGREATER,
	token.GTE:        LESSGREATER,
	token.COLON:      SLICE,
	token.PLUS:       SUM,
	token.MINUS:      SUM,
	token.PLUS_A:     SUM,
	token.MINUS_A:    SUM,
	token.MOD:        PRODUCT,
	token.ASTERISK:   PRODUCT,
	token.ASTERISK_A: PRODUCT,
	token.SLASH:      PRODUCT,
	token.SLASH_A:    PRODUCT,
	token.LPAREN:     CALL,
	token.DOT:        CALL,
	token.LBRACKET:   INDEX,
	token.INCR:       INCRDECR,
	token.DECR:       INCRDECR,
}

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token
	errors    []string

	// parsing functions for each  type
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []string{},
		prefixParseFns: map[token.TokenType]prefixParseFn{},
		infixParseFns:  map[token.TokenType]infixParseFn{},
	}

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.INCR, p.parsePrefixExpression)
	p.registerPrefix(token.DECR, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.WHILE, p.parseWhileExpression)
	p.registerPrefix(token.BREAK, p.parseBreakExpression)
	p.registerPrefix(token.CONTINUE, p.parseContinueExpression)
	p.registerPrefix(token.FOR, p.parseForExpression)
	p.registerPrefix(token.TYPEOF, p.parseTypeofExpression)
	p.registerPrefix(token.NULL, p.parseNullExpression)
	p.registerPrefix(token.LBRACE, p.parseHashExpression)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMethodCallExpression)
	p.registerInfix(token.COLON, p.parseSliceExpression)

	p.registerInfix(token.ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.PLUS_A, p.parseAssignExpression)
	p.registerInfix(token.MINUS_A, p.parseAssignExpression)
	p.registerInfix(token.ASTERISK_A, p.parseAssignExpression)
	p.registerInfix(token.SLASH_A, p.parseAssignExpression)

	p.registerInfix(token.INCR, p.parsePostfixExpression)
	p.registerInfix(token.DECR, p.parsePostfixExpression)
	// read first two token so curToken and peekToken is set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Statements: make([]ast.Statement, 0)}
	if p.curTokenIs(token.SEMICOLON) && p.peekTokenIs(token.EOF) {
		return program
	}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.DELETE:
		return p.parseDeleteStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if p.expectPeek(token.IDENT) {
		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	if p.expectPeek(token.ASSIGN) {
		p.nextToken()
		stmt.Value = p.parseExpressionStatement().Expression
	}

	return stmt
}

func (p *Parser) parseDeleteStatement() *ast.DeleteStatement {
	stmt := &ast.DeleteStatement{Token: p.curToken}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
		return stmt
	}
	p.nextToken()
	stmt.ReturnValue = p.parseExpressionStatement().Expression
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	//if p.peekTokenIs(token.SEMICOLON) {
	//	p.nextToken()
	//}
	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

// everything except LET, RETURN is an expression.
// core of the Pratt parsing algo
func (p *Parser) parseExpression(precedence int) ast.Expression {
	if p.curTokenIs(token.SEMICOLON) {
		return nil
	}
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExpr := prefix()

	// call infixParseFns again and again until encounters a token with higher precedence
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExpr
		}
		p.nextToken()
		leftExpr = infix(leftExpr)
	}
	return leftExpr
}

func (p *Parser) parseMethodCallExpression(left ast.Expression) ast.Expression {
	methodCall := &ast.MethodCallExpression{Token: p.curToken, Object: left}
	p.nextToken()
	name := p.parseIdentifier()
	if !p.peekTokenIs(token.LPAREN) {
		methodCall.Call = p.parseExpression(LOWEST)
	} else {
		p.nextToken()
		methodCall.Call = p.parseCallExpression(name)
	}
	return methodCall
}

func (p *Parser) parseSliceExpression(left ast.Expression) ast.Expression {
	expr := &ast.SliceExpression{Token: p.curToken}
	expr.Start = left
	if p.peekTokenIs(token.RBRACKET) { // [:end]
		expr.End = nil
	} else {
		p.nextToken()
		expr.End = p.parseExpression(LOWEST)
	}
	return expr
}

func (p *Parser) parseHashExpression() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return hash
	}
	for !p.curTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(SLICE)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		hash.Pairs[key] = p.parseExpression(LOWEST)
		p.nextToken()
	}
	return hash
}
