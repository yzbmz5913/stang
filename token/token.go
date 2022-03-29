package token

import "fmt"

type Token struct {
	Type    TokenType
	Literal string
	Pos     Position // the position of the token in source code
}

type Position struct {
	Offset int //offset relative to entire file
	Line   int
	Col    int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT"
	INT   = "INT"
	FLOAT = "FLOAT"

	EQ         = "=="
	NEQ        = "!="
	ASSIGN     = "="
	PLUS       = "+"
	PLUS_A     = "+="
	INCR       = "++"
	MINUS      = "-"
	MINUS_A    = "-="
	DECR       = "--"
	BANG       = "!"
	ASTERISK   = "*"
	ASTERISK_A = "*="
	SLASH      = "/"
	SLASH_A    = "/="
	MOD        = "%"
	AND        = "&&"
	OR         = "||"

	LT        = "<"
	GT        = ">"
	LTE       = "<="
	GTE       = ">="
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"
	COLON    = ":"
	DOT      = "."

	FUNCTION = "FUNCTION"
	LET      = "LET"
	DELETE   = "DELETE"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	STRING   = "STRING"
	TYPEOF   = "TYPEOF"
	WHILE    = "WHILE"
	FOR      = "FOR"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	NULL     = "NULL"
)

var keywords = map[string]TokenType{
	"function": FUNCTION,
	"let":      LET,
	"delete":   DELETE,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"typeof":   TYPEOF,
	"while":    WHILE,
	"break":    BREAK,
	"for":      FOR,
	"continue": CONTINUE,
	"null":     NULL,
}

func NewToken(typ TokenType, ch byte) Token {
	return Token{Type: typ, Literal: string(ch)}
}

func IdentOrKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
