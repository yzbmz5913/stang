package parser

import (
	"fmt"
	"github.com/yzbmz5913/stang/ast"
	"github.com/yzbmz5913/stang/token"
	"strconv"
)

// All token-parsing function must follow a protocol:
// Start with curToken being the type of token associated with
// Return with curToken being the last token that’s part of the expression type
// Never advance the tokens too far.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	il := &ast.IntegerLiteral{Token: p.curToken}
	i, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	il.Value = i
	return il
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	expr := &ast.ArrayLiteral{Token: p.curToken}
	expr.Elements = p.parseExpressionList(token.RBRACKET)
	return expr
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fl := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	fl.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	fl.Body = p.parseBlockStatement()
	return fl
}
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := make([]ast.Expression, 0)
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var idents []*ast.Identifier
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return idents
	}
	p.nextToken()
	idents = append(idents, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		idents = append(idents, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return idents
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	return expr
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	return &ast.PostfixExpression{Token: p.curToken, Left: left, Operator: p.curToken.Literal}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	// curToken is '('
	p.nextToken()
	expr := p.parseExpression(LOWEST) // boost the precedence of the enclosed expression
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return expr
}

func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	expr.Condition = p.parseExpression(LOWEST)
	if !p.curTokenIs(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expr.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		expr.Alternative = p.parseBlockStatement()
	}
	return expr
}

func (p *Parser) parseWhileExpression() ast.Expression {
	expr := &ast.WhileExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	expr.Condition = p.parseExpression(LOWEST)
	if !p.curTokenIs(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expr.Body = p.parseBlockStatement()
	return expr
}

func (p *Parser) parseBreakExpression() ast.Expression {
	return &ast.BreakExpression{Token: p.curToken}
}

func (p *Parser) parseContinueExpression() ast.Expression {
	return &ast.ContinueExpression{Token: p.curToken}
}

func (p *Parser) parseNullExpression() ast.Expression {
	return &ast.NullExpression{Token: p.curToken}
}

func (p *Parser) parseForExpression() ast.Expression {
	curToken := p.curToken

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	var result ast.Expression

	var init ast.Node
	var cond ast.Expression
	var update ast.Expression

	p.nextToken()
	if !p.curTokenIs(token.SEMICOLON) {
		if p.curTokenIs(token.LET) {
			init = p.parseLetStatement()
		} else {
			init = p.parseExpression(LOWEST)
		}
		p.nextToken()
	}

	p.nextToken()
	if !p.curTokenIs(token.SEMICOLON) {
		cond = p.parseExpression(LOWEST)
		p.nextToken()
	}

	p.nextToken()
	if !p.curTokenIs(token.RPAREN) {
		update = p.parseExpression(LOWEST)
		p.nextToken()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	loop := &ast.ForExpression{Token: curToken, Init: init, Condition: cond, Update: update}
	loop.Body = p.parseBlockStatement()
	result = loop

	return result
}

func (p *Parser) parseTypeofExpression() ast.Expression {
	te := &ast.TypeofExpression{Token: p.curToken}
	p.nextToken()
	te.Expr = p.parseExpression(LOWEST)
	return te
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	expr := &ast.CallExpression{Token: p.curToken, Function: left}
	expr.Arguments = p.parseExpressionList(token.RPAREN)
	return expr
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	expr := &ast.IndexExpression{Token: p.curToken, Left: left}
	var index ast.Expression
	if p.peekTokenIs(token.COLON) { // [:end]
		start := &ast.IntegerLiteral{Token: token.NewToken(token.INT, "0"[0]), Value: 0}
		p.nextToken() // :
		index = p.parseSliceExpression(start)
	} else { // [index] or [start:end]
		p.nextToken()
		index = p.parseExpression(LOWEST)
	}
	expr.Index = index
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return expr
}

func (p *Parser) parseAssignExpression(name ast.Expression) ast.Expression {
	e := &ast.AssignExpression{Token: p.curToken}

	if n, ok := name.(*ast.Identifier); ok {
		e.Name = n
	} else if indexExp, ok := name.(*ast.IndexExpression); ok {
		e.Name = indexExp
	} else {
		msg := fmt.Sprintf("expected assign token to be an identifier, got %s instead", name.TokenLiteral())
		p.errors = append(p.errors, msg)
		return e
	}

	p.nextToken()
	e.Value = p.parseExpression(LOWEST)

	return e
}
