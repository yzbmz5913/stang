package parser

import (
	"fmt"
	"github.com/yzbmz5913/stang/ast"
	"github.com/yzbmz5913/stang/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression // param: LHS
)

func (p *Parser) registerPrefix(typ token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[typ] = fn
}
func (p *Parser) registerInfix(typ token.TokenType, fn infixParseFn) {
	p.infixParseFns[typ] = fn
}

func (p *Parser) curTokenIs(typ token.TokenType) bool {
	return p.curToken.Type == typ
}

func (p *Parser) peekTokenIs(typ token.TokenType) bool {
	return p.peekToken.Type == typ
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) expectPeek(typ token.TokenType) bool {
	if p.peekTokenIs(typ) {
		p.nextToken()
		return true
	}
	p.peekError(typ)
	return false
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// parser error handlers
func (p *Parser) peekError(typ token.TokenType) {
	msg := fmt.Sprintf("[%s]expected token to be %s, got %s instead", p.peekToken.Pos, typ, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("[%s]no prefix parse function for %s found", p.curToken.Pos, t)
	p.errors = append(p.errors, msg)
}
