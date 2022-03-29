package evaluator

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"stang/ast"
	"strconv"
	"strings"
)

// object.go defines the internal representation of all runtime data

type ObjectType string

const (
	IntegerObj     = "INTEGER"
	FloatObj       = "FLOAT"
	BooleanObj     = "BOOLEAN"
	NullObj        = "NULL"
	ReturnValueObj = "RETURN_VALUE"
	BreakObj       = "BREAK"
	ContinueObj    = "CONTINUE"
	ErrorObj       = "ERROR"
	FunctionObj    = "FUNCTION"
	StringObj      = "STRING"
	BuiltinObj     = "BUILTIN"
	ArrayObj       = "ARRAY"
	HashObj        = "HASH"
)

type Object interface {
	Type() ObjectType
	String(stack int) string
	CallMethod(method string, args ...Object) Object
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType  { return IntegerObj }
func (i *Integer) String(int) string { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}
func (i *Integer) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, i.Type())
}

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FloatObj }
func (f *Float) String(int) string {
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}
func (f *Float) HashKey() HashKey {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strconv.FormatFloat(f.Value, 'f', -1, 64)))
	return HashKey{Type: f.Type(), Value: h.Sum64()}
}
func (f *Float) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, f.Type())
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType  { return BooleanObj }
func (b *Boolean) String(int) string { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}
func (b *Boolean) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, b.Type())
}

type Null struct{}

func (n *Null) Type() ObjectType  { return NullObj }
func (n *Null) String(int) string { return "null" }
func (n *Null) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, n.Type())
}

type Error struct {
	Msg string
}

func (e *Error) Type() ObjectType  { return ErrorObj }
func (e *Error) String(int) string { return "Error: " + e.Msg }
func (e *Error) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, e.Type())
}

type Break struct{}

func (b *Break) Type() ObjectType  { return BreakObj }
func (b *Break) String(int) string { return "break" }
func (b *Break) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, b.Type())
}

type Continue struct{}

func (c *Continue) Type() ObjectType  { return ContinueObj }
func (c *Continue) String(int) string { return "continue" }
func (c *Continue) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, c.Type())
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType        { return ReturnValueObj }
func (rv *ReturnValue) String(stack int) string { return rv.Value.String(stack) }
func (rv *ReturnValue) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, rv.Type())
}

type String struct{ Value string }

func (s *String) String(int) string { return s.Value }
func (s *String) Type() ObjectType  { return StringObj }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}
func (s *String) CallMethod(method string, args ...Object) Object {
	switch method {
	case "toLower":
		if len(args) != 0 {
			return newError(ARGUMENTNUMERROR, "0", len(args))
		}
		lower := strings.ToLower(s.Value)
		return &String{Value: lower}
	case "toUpper":
		if len(args) != 0 {
			return newError(ARGUMENTNUMERROR, "0", len(args))
		}
		upper := strings.ToUpper(s.Value)
		return &String{Value: upper}
	case "split":
		if len(args) != 1 {
			return newError(ARGUMENTNUMERROR, "1", len(args))
		}
		if args[0].Type() != StringObj {
			return newError(ARGUMENTTYPEERROR, StringObj, args[1].Type())
		}
		strs := strings.Split(s.Value, args[0].(*String).Value)
		elements := make([]Object, 0)
		for _, str := range strs {
			elements = append(elements, &String{Value: str})
		}
		return &Array{Elements: elements}
	}
	return newError(NOMETHODERROR, method, s.Type())
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Scope      *Scope
}

func (f *Function) Type() ObjectType { return FunctionObj }
func (f *Function) String(int) string {
	out := bytes.Buffer{}
	params := make([]string, 0)
	for _, param := range f.Parameters {
		params = append(params, param.String())
	}
	out.WriteString("function(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") { ")
	out.WriteString(f.Body.String())
	out.WriteString(" }")
	return out.String()
}
func (f *Function) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, f.Type())
}

type BuiltinFunction func(args ...Object) Object
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType  { return BuiltinObj }
func (b *Builtin) String(int) string { return "[builtin]" }
func (b *Builtin) CallMethod(method string, _ ...Object) Object {
	return newError(NOMETHODERROR, method, b.Type())
}

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType {
	return ArrayObj
}

func (a *Array) String(stack int) string {
	if stack == 10 {
		return "ARRAY..."
	}
	var out bytes.Buffer
	var elements []string
	for _, e := range a.Elements {
		elements = append(elements, e.String(stack+1))
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}
func (a *Array) CallMethod(method string, args ...Object) Object {
	switch method {
	case "push":
		for _, obj := range args {
			a.Elements = append(a.Elements, obj)
		}
		return &Integer{Value: int64(len(a.Elements))}
	case "pop":
		l := len(a.Elements)
		if l == 0 {
			return newErrorf("array is empty")
		}
		ret := a.Elements[l-1]
		a.Elements = a.Elements[:l-1]
		return ret
	}
	return newError(NOMETHODERROR, method, a.Type())
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}
type HashPair struct {
	Key   Object
	Value Object
}
type Hash struct {
	Pairs map[HashKey]HashPair
}
type Hashable interface {
	HashKey() HashKey
}

func (h *Hash) Type() ObjectType { return HashObj }
func (h *Hash) String(stack int) string {
	if stack == 10 {
		return "HASH..."
	}
	var out bytes.Buffer
	var pairs []string
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s:%s", pair.Key.String(stack+1), pair.Value.String(stack+1)))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

func (h *Hash) CallMethod(method string, args ...Object) Object {
	return newError(NOMETHODERROR, method, h.Type())
}
