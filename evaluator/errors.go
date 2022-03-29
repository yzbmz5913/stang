package evaluator

import "fmt"

const (
	_ int = iota
	PREFIXOP
	INFIXOP
	POSTFIXOP
	UNKNOWNIDENT
	NOMETHODERROR
	NOINDEXERROR
	NOTHASHABLE
	INDEXERROR
	SLICEERROR
	ARGUMENTNUMERROR
	ARGUMENTTYPEERROR
	RTERROR
	CONSTRUCTERR
	INLENERR
	DIVIDEBYZERO
	NOTFUNC
	REDEFINE
	TIMEOUT
	NOTLVALUE
	INDEXINT
)

var errorType = map[int]string{
	PREFIXOP:          "unsupported prefix operator '%s' for type: %s",
	INFIXOP:           "unsupported infix operator '%s' for type %s and %s",
	POSTFIXOP:         "unsupported postfix operator '%s' for type: %s",
	UNKNOWNIDENT:      "unknown identifier: '%s' is not defined",
	NOMETHODERROR:     "undefined method '%s' for object %s",
	NOINDEXERROR:      "type %s does not support index operator",
	NOTHASHABLE:       "type %s is not hashable",
	INDEXERROR:        "index '%d' is out of range, valid range is [%d, %d]",
	SLICEERROR:        "slicing start index %d must not be greater than end index %d",
	ARGUMENTNUMERROR:  "wrong number of arguments. expected: %s, got: %d",
	ARGUMENTTYPEERROR: "wrong type of arguments. expected: %s, got: %s",
	RTERROR:           "return type should be %s",
	CONSTRUCTERR:      "%s argument for addm should be type %s. got: %s",
	INLENERR:          "function %s takes input with max length %s. got: %s",
	DIVIDEBYZERO:      "cannot divide by zero",
	NOTFUNC:           "%s is not a function",
	REDEFINE:          "variable %s has been defined",
	TIMEOUT:           "evaluation timeout",
	NOTLVALUE:         "the expression %s is not an lvalue",
	INDEXINT:          "index must be integer",
}

func newError(t int, args ...interface{}) Object {
	return &Error{Msg: fmt.Sprintf(errorType[t], args...)}
}
func newErrorf(format string, args ...interface{}) Object {
	return &Error{Msg: fmt.Sprintf(format, args...)}
}
