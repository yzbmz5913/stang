package ast

import (
	"bytes"
	"github.com/yzbmz5913/stang/token"
	"strings"
)

// Node has two types: statement and expression
type Node interface {
	TokenLiteral() string // help debug
	String() string
}

type Statement interface {
	statementNode() //dummy
	Node
}
type Expression interface {
	expressionNode() //dummy
	Node
}

// Program is the root of AST
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}
func (p *Program) String() string {
	out := bytes.Buffer{}
	for _, stmt := range p.Statements {
		out.WriteString(stmt.String())
	}
	return out.String()
}

type NullExpression struct {
	Token token.Token // the NULL token
}

func (n *NullExpression) expressionNode()      {}
func (n *NullExpression) TokenLiteral() string { return "null" }
func (n *NullExpression) String() string       { return "null" }

type LetStatement struct {
	Token token.Token // the LET token
	Name  *Identifier // identifier name
	Value Expression  // RHS value
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	out := bytes.Buffer{}
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	return out.String()
}

type DeleteStatement struct {
	Token token.Token // the DELETE token
	Value Expression  // RHS expr(identifier or indexExpression)
}

func (d *DeleteStatement) statementNode()       {}
func (d *DeleteStatement) TokenLiteral() string { return d.Token.Literal }
func (d *DeleteStatement) String() string {
	out := bytes.Buffer{}
	out.WriteString("delete ")
	out.WriteString(d.Value.String())
	return out.String()
}

type Identifier struct {
	Token token.Token // the IDENT token
	Value string      // the name of the identifier, for convenience
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string {
	return i.Value
}

type ReturnStatement struct {
	Token       token.Token // the RETURN token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	out := bytes.Buffer{}
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	return out.String()
}

// ExpressionStatement is a wrapper for expression to statement
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (i *IntegerLiteral) expressionNode()      {}
func (i *IntegerLiteral) TokenLiteral() string { return i.Token.Literal }
func (i *IntegerLiteral) String() string       { return i.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

type PrefixExpression struct {
	Token    token.Token // the prefix token, e.g. ! - ++
	Operator string
	Right    Expression
}

func (p *PrefixExpression) expressionNode()      {}
func (p *PrefixExpression) TokenLiteral() string { return p.Token.Literal }
func (p *PrefixExpression) String() string {
	out := bytes.Buffer{}
	out.WriteString("(")
	out.WriteString(p.Operator)
	out.WriteString(p.Right.String())
	out.WriteString(")")
	return out.String()
}

type InfixExpression struct {
	Token    token.Token // the operator e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (i *InfixExpression) expressionNode()      {}
func (i *InfixExpression) TokenLiteral() string { return i.Token.Literal }
func (i *InfixExpression) String() string {
	out := bytes.Buffer{}
	out.WriteString("(")
	out.WriteString(i.Left.String())
	out.WriteString(" " + i.Operator + " ")
	out.WriteString(i.Right.String())
	out.WriteString(")")
	return out.String()
}

type PostfixExpression struct {
	Token    token.Token // the prefix token, e.g. ! - ++
	Operator string
	Left     Expression
}

func (p *PostfixExpression) expressionNode()      {}
func (p *PostfixExpression) TokenLiteral() string { return p.Token.Literal }
func (p *PostfixExpression) String() string {
	out := bytes.Buffer{}
	out.WriteString("(")
	out.WriteString(p.Left.String())
	out.WriteString(p.Operator)
	out.WriteString(")")
	return out.String()
}

type IfExpression struct {
	Token       token.Token // the IF token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (i *IfExpression) expressionNode() {}

func (i *IfExpression) TokenLiteral() string { return i.Token.Literal }

func (i *IfExpression) String() string {
	out := bytes.Buffer{}
	out.WriteString("if")
	out.WriteString(i.Condition.String())
	out.WriteString(" ")
	out.WriteString(i.Consequence.String())
	if i.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(i.Alternative.String())
	}
	return out.String()
}

type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (b *BlockStatement) statementNode()       {}
func (b *BlockStatement) TokenLiteral() string { return b.Token.Literal }
func (b *BlockStatement) String() string {
	out := bytes.Buffer{}
	for _, stmt := range b.Statements {
		if exprStmt, ok := stmt.(*ExpressionStatement); ok {
			if exprStmt.Expression == nil {
				continue
			}
		}
		out.WriteString(stmt.String())
		out.WriteString("; ")
	}
	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token // the FUNCTION token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (f *FunctionLiteral) expressionNode()      {}
func (f *FunctionLiteral) TokenLiteral() string { return f.Token.Literal }
func (f *FunctionLiteral) String() string {
	out := bytes.Buffer{}
	var params []string
	for _, param := range f.Parameters {
		params = append(params, param.String())
	}
	out.WriteString(f.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ","))
	out.WriteString(")")
	out.WriteString(f.Body.String())

	return out.String()
}

type CallExpression struct {
	Token     token.Token // the ( token
	Function  Expression  // function identifier expression
	Arguments []Expression
}

func (c *CallExpression) expressionNode()      {}
func (c *CallExpression) TokenLiteral() string { return c.Token.Literal }
func (c *CallExpression) String() string {
	out := bytes.Buffer{}
	var args []string
	for _, arg := range c.Arguments {
		args = append(args, arg.String())
	}
	out.WriteString(c.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type WhileExpression struct {
	Token     token.Token // the WHILE token
	Condition Expression
	Body      *BlockStatement
}

func (dl *WhileExpression) expressionNode()      {}
func (dl *WhileExpression) TokenLiteral() string { return dl.Token.Literal }
func (dl *WhileExpression) String() string {
	var out bytes.Buffer

	out.WriteString(dl.Token.Literal)
	out.WriteString(" ( ")
	if dl.Condition != nil {
		out.WriteString(dl.Condition.String())
	}
	out.WriteString(" )")
	out.WriteString(" { ")
	if dl.Body != nil {
		out.WriteString(dl.Body.String())
	}
	out.WriteString(" }")
	return out.String()
}

type BreakExpression struct {
	Token token.Token
}

func (be *BreakExpression) expressionNode()      {}
func (be *BreakExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BreakExpression) String() string       { return be.Token.Literal }

type ContinueExpression struct {
	Token token.Token
}

func (ce *ContinueExpression) expressionNode()      {}
func (ce *ContinueExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *ContinueExpression) String() string       { return ce.Token.Literal }

type TypeofExpression struct {
	Token token.Token // the TYPEOF token
	Expr  Expression
}

func (t *TypeofExpression) expressionNode()      {}
func (t *TypeofExpression) TokenLiteral() string { return t.Token.Literal }
func (t *TypeofExpression) String() string       { return "typeof" + t.Expr.String() }

type AssignExpression struct {
	Token token.Token // the assign token e.g. = -= +=
	Name  Expression
	Value Expression
}

func (a *AssignExpression) expressionNode()      {}
func (a *AssignExpression) TokenLiteral() string { return a.Token.Literal }
func (a *AssignExpression) String() string {
	var out bytes.Buffer
	out.WriteString(a.Name.String())
	out.WriteString(a.Token.Literal)
	out.WriteString(a.Value.String())

	return out.String()
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("]")
	out.WriteString(")")
	return out.String()
}

type ForExpression struct {
	Token     token.Token
	Init      Node
	Condition Expression
	Update    Expression
	Body      *BlockStatement
}

func (f *ForExpression) expressionNode()      {}
func (f *ForExpression) TokenLiteral() string { return f.Token.Literal }
func (f *ForExpression) String() string {
	var out bytes.Buffer
	out.WriteString("for")
	out.WriteString(" ( ")
	if f.Init != nil {
		out.WriteString(f.Init.String())
	}
	out.WriteString(" ; ")
	if f.Condition != nil {
		out.WriteString(f.Condition.String())
	}
	out.WriteString(" ; ")
	if f.Update != nil {
		out.WriteString(f.Update.String())
	}
	out.WriteString(" ) ")
	out.WriteString(" { ")
	if f.Body != nil {
		out.WriteString(f.Body.String())
	}
	out.WriteString(" }")

	return out.String()
}

type StringLiteral struct {
	Token token.Token // the STRING token
	Value string
}

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }
func (s *StringLiteral) String() string       { return s.Token.Literal }

type ArrayLiteral struct {
	Token    token.Token // the [ token
	Elements []Expression
}

func (a *ArrayLiteral) expressionNode()      {}
func (a *ArrayLiteral) TokenLiteral() string { return a.Token.Literal }
func (a *ArrayLiteral) String() string {
	out := bytes.Buffer{}
	out.WriteString("[")
	strs := make([]string, 0)
	for _, ele := range a.Elements {
		strs = append(strs, ele.String())
	}
	out.WriteString(strings.Join(strs, ", "))
	out.WriteString("]")
	return out.String()
}

type MethodCallExpression struct {
	Token  token.Token
	Object Expression
	Call   Expression
}

func (mc *MethodCallExpression) expressionNode()      {}
func (mc *MethodCallExpression) TokenLiteral() string { return mc.Token.Literal }
func (mc *MethodCallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(mc.Object.String())
	out.WriteString(".")
	out.WriteString(mc.Call.String())

	return out.String()
}

type SliceExpression struct {
	Token token.Token // the : token
	Start Expression
	End   Expression
}

func (s *SliceExpression) expressionNode()      {}
func (s *SliceExpression) TokenLiteral() string { return s.Token.Literal }
func (s *SliceExpression) String() string {
	out := bytes.Buffer{}
	out.WriteString("(")
	out.WriteString(s.Start.String())
	out.WriteString(":")
	out.WriteString(s.End.String())
	out.WriteString(")")
	return out.String()
}

type HashLiteral struct {
	Token token.Token // the { token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	var pairs []string
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
