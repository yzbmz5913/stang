package lexer

import (
	"errors"
	"github.com/yzbmz5913/stang/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to ch)
	readPosition int  // next position after ch, sometimes we need to check forward input
	ch           byte // current character
	line         int
	col          int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
		if l.ch == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
	}
	l.position = l.readPosition
	l.readPosition++
}

var tokenMap = map[byte]token.TokenType{
	'=': token.ASSIGN,
	'.': token.DOT,
	';': token.SEMICOLON,
	'(': token.LPAREN,
	')': token.RPAREN,
	'{': token.LBRACE,
	'}': token.RBRACE,
	'[': token.LBRACKET,
	']': token.RBRACKET,
	'+': token.PLUS,
	',': token.COMMA,
	'-': token.MINUS,
	'!': token.BANG,
	'*': token.ASTERISK,
	'/': token.SLASH,
	'<': token.LT,
	'>': token.GT,
	':': token.COLON,
	'%': token.MOD,
	'&': token.AND,
	'|': token.OR,
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()
	var tok token.Token
	pos := token.Position{
		Offset: l.position,
		Line:   l.line,
		Col:    l.col - 1,
	}
	// starts with a symbol
	if t, ok := tokenMap[l.ch]; ok {
		switch t {
		// multiple-characters operators: <= >= == != += ++ -= -- *= /=
		case token.LT:
			if l.peekChar() == '=' {
				tok = token.Token{Type: token.LTE, Literal: "<="}
				l.readChar()
			}
		case token.GT:
			if l.peekChar() == '=' {
				tok = token.Token{Type: token.GTE, Literal: ">="}
				l.readChar()
			}
		case token.ASSIGN:
			if l.peekChar() == '=' {
				tok = token.Token{Type: token.EQ, Literal: "=="}
				l.readChar()
			}
		case token.BANG:
			if l.peekChar() == '=' {
				tok = token.Token{Type: token.NEQ, Literal: "!="}
				l.readChar()
			}
		case token.PLUS:
			if l.peekChar() == '=' {
				tok = token.Token{Type: token.PLUS_A, Literal: "+="}
				l.readChar()
			} else if l.peekChar() == '+' {
				tok = token.Token{Type: token.INCR, Literal: "++"}
				l.readChar()
			}
		case token.MINUS:
			if l.peekChar() == '=' {
				tok = token.Token{Type: token.MINUS_A, Literal: "-="}
				l.readChar()
			} else if l.peekChar() == '-' {
				tok = token.Token{Type: token.DECR, Literal: "--"}
				l.readChar()
			}
		case token.ASTERISK:
			if l.peekChar() == '=' {
				tok = token.Token{Type: token.ASTERISK_A, Literal: "*="}
				l.readChar()
			}
		case token.SLASH:
			if l.peekChar() == '=' {
				tok = token.Token{Type: token.SLASH_A, Literal: "/="}
				l.readChar()
			}
		case token.AND:
			if l.peekChar() == '&' {
				tok = token.Token{Type: token.AND, Literal: "&&"}
				l.readChar()
			}
		case token.OR:
			if l.peekChar() == '|' {
				tok = token.Token{Type: token.OR, Literal: "||"}
				l.readChar()
			}
		}
		if tok.Type == "" {
			tok = token.NewToken(t, l.ch)
		}
		l.readChar()
		tok.Pos = pos
		return tok
	}
	// not starts with a symbol
	tok = l.readMultiCharToken()
	tok.Pos = pos
	return tok
}

func (l *Lexer) readMultiCharToken() token.Token {
	var tok token.Token
	switch {
	case l.ch == 0:
		tok.Type = token.EOF
		tok.Literal = ""
		return tok
	case isLetter(l.ch): // keyword or identifier
		start := l.position
		for isLetter(l.ch) || isDigit(l.ch) {
			l.readChar()
		}
		tok.Literal = l.input[start:l.position]
		tok.Type = token.IdentOrKeyword(tok.Literal)
		return tok
	case isDigit(l.ch): // number
		tok = l.readNumber()
		return tok
	case l.ch == '\'':
		if str, err := l.readString(1); err == nil {
			tok.Type = token.STRING
			tok.Literal = str
			l.readChar()
			return tok
		}
	case l.ch == '"':
		if str, err := l.readString(2); err == nil {
			tok.Type = token.STRING
			tok.Literal = str
			l.readChar()
			return tok
		}
	}
	l.readChar()
	return token.NewToken(token.ILLEGAL, l.ch)
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readNumber() token.Token {
	var tok token.Token
	buf := make([]byte, 0)
	hasDot := false
	for isDigit(l.ch) || l.ch == '.' {
		if l.ch == '.' {
			hasDot = true
		}
		buf = append(buf, l.ch)
		l.readChar()
	}
	if hasDot {
		tok.Type = token.FLOAT
	} else {
		tok.Type = token.INT
	}
	tok.Literal = string(buf)
	return tok
}

func (l *Lexer) readString(quote uint8) (string, error) {
	start := l.readPosition
	if quote == 1 {
		for {
			l.readChar()
			if l.ch == '\'' {
				break
			} else if l.ch == 0 {
				return "", errors.New("")
			}
		}
		return l.input[start:l.position], nil
	} else if quote == 2 {
		for {
			l.readChar()
			if l.ch == '"' {
				break
			} else if l.ch == 0 {
				return "", errors.New("")
			}
		}
		return l.input[start:l.position], nil
	}
	return "", errors.New("")
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
