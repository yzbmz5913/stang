package evaluator

import (
	"context"
	"math"
	"stang/ast"
)

var (
	NULL     = &Null{}
	BREAK    = &Break{}
	CONTINUE = &Continue{}
	TRUE     = &Boolean{Value: true}
	FALSE    = &Boolean{Value: false}
)

func Eval(ctx context.Context, node ast.Node, s *Scope) Object {
	select {
	case <-ctx.Done():
		return newError(TIMEOUT)
	default:
		switch node := node.(type) {
		// statements
		case *ast.Program:
			return evalProgram(ctx, node.Statements, s)
		case *ast.ExpressionStatement:
			return Eval(ctx, node.Expression, s)
		case *ast.BlockStatement:
			return evalBlockStatement(ctx, node.Statements, s)
		case *ast.ReturnStatement:
			return &ReturnValue{Value: Eval(ctx, node.ReturnValue, s)}
		case *ast.LetStatement:
			return evalLetStatement(ctx, node, s)
		case *ast.DeleteStatement:
			return evalDeleteStatement(ctx, node, s)

		// expressions
		case *ast.IntegerLiteral:
			return &Integer{Value: node.Value}
		case *ast.FloatLiteral:
			return &Float{Value: node.Value}
		case *ast.BooleanLiteral:
			return nativeBoolToBooleanObject(node.Value)
		case *ast.StringLiteral:
			return &String{Value: node.Value}
		case *ast.ArrayLiteral:
			return evalArrayLiteral(ctx, node, s)
		case *ast.HashLiteral:
			return evalHashLiteral(ctx, node, s)
		case *ast.NullExpression:
			return NULL
		case *ast.FunctionLiteral:
			return &Function{Parameters: node.Parameters, Body: node.Body, Scope: s}
		case *ast.PrefixExpression:
			return evalPrefixExpression(node.Operator, Eval(ctx, node.Right, s))
		case *ast.InfixExpression:
			return evalInfixExpression(Eval(ctx, node.Left, s), node.Operator, Eval(ctx, node.Right, s))
		case *ast.PostfixExpression:
			return evalPostfixExpression(Eval(ctx, node.Left, s), node.Operator)
		case *ast.IfExpression:
			return evalIfExpression(ctx, node, s)
		case *ast.WhileExpression:
			return evalWhileExpression(ctx, node, s)
		case *ast.BreakExpression:
			return BREAK
		case *ast.ContinueExpression:
			return CONTINUE
		case *ast.ForExpression:
			return evalForExpression(ctx, node, s)
		case *ast.Identifier:
			return evalIdentifier(node, s)
		case *ast.TypeofExpression:
			return evalTypeofExpression(ctx, node, s)
		case *ast.AssignExpression:
			return evalAssignExpression(ctx, node, s)
		case *ast.CallExpression:
			return evalCallExpression(ctx, node, s)
		case *ast.MethodCallExpression:
			return evalMethodCallExpression(ctx, node, s)
		case *ast.IndexExpression:
			return evalIndexExpression(ctx, node, s)
		}
	}

	return nil
}

func evalProgram(ctx context.Context, stmts []ast.Statement, s *Scope) Object {
	var result Object
	for _, stmt := range stmts {
		if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			if exprStmt.Expression == nil {
				continue
			}
		}
		result = Eval(ctx, stmt, s)
		if returnValue, ok := result.(*ReturnValue); ok {
			return returnValue.Value
		}
		if err, ok := result.(*Error); ok {
			return err
		}
	}
	return result
}

func evalBlockStatement(ctx context.Context, stmts []ast.Statement, s *Scope) Object {
	var result Object
	for _, statement := range stmts {
		if exprStmt, ok := statement.(*ast.ExpressionStatement); ok {
			if exprStmt.Expression == nil {
				continue
			}
		}
		result = Eval(ctx, statement, s)
		if result != nil {
			typ := result.Type()
			if typ == ReturnValueObj || typ == ErrorObj {
				return result
			}
			if _, ok := result.(*Break); ok {
				return result
			}
			if _, ok := result.(*Continue); ok {
				return result
			}
		}
	}
	return result
}

func evalLetStatement(ctx context.Context, node *ast.LetStatement, s *Scope) Object {
	if _, ok := s.GetCurrent(node.Name.String()); ok {
		return newError(REDEFINE, node.Name.String())
	}
	v := Eval(ctx, node.Value, s)
	if v.Type() != ErrorObj {
		return s.Set(node.Name.String(), v)
	}
	return v
}

func evalDeleteStatement(ctx context.Context, node *ast.DeleteStatement, s *Scope) Object {
	value := node.Value
	switch deleted := value.(type) {
	case *ast.Identifier:
		old := Eval(ctx, value, s)
		s.Delete(deleted.Value)
		return old
	case *ast.IndexExpression:
		left := Eval(ctx, deleted.Left, s)
		index := Eval(ctx, deleted.Index, s)
		switch l := left.(type) {
		case *Array:
			i, e := calcIndex(len(l.Elements), index, false)
			if e != nil {
				return e
			}
			old := l.Elements[i]
			l.Elements[i] = NULL
			return old
		case *Hash:
			if hashable, ok := index.(Hashable); ok {
				old, ok := l.Pairs[hashable.HashKey()]
				if !ok {
					return NULL
				}
				delete(l.Pairs, hashable.HashKey())
				return old.Value
			}
		}
		return NULL
	default:
		return newError(NOTLVALUE, value.String())
	}
}

func nativeBoolToBooleanObject(b bool) *Boolean {
	if b {
		return TRUE
	}
	return FALSE
}
func isNumber(obj Object) bool {
	return obj.Type() == IntegerObj || obj.Type() == FloatObj
}

func evalIfExpression(ctx context.Context, node *ast.IfExpression, s *Scope) Object {
	cond := Eval(ctx, node.Condition, s)
	if isTruthy(cond) {
		return Eval(ctx, node.Consequence, s)
	} else if node.Alternative != nil {
		return Eval(ctx, node.Alternative, s)
	}
	return NULL
}

func evalWhileExpression(ctx context.Context, wl *ast.WhileExpression, scope *Scope) Object {
	innerScope := NewScope(scope)

	condition := Eval(ctx, wl.Condition, innerScope)
	if condition.Type() == ErrorObj {
		return condition
	}

	var result Object
	for isTruthy(condition) {
		result = Eval(ctx, wl.Body, innerScope)
		if result.Type() == ErrorObj {
			return result
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			if v.Value != nil {
				return v
			}
			break
		}
		condition = Eval(ctx, wl.Condition, innerScope)
		if condition.Type() == ErrorObj {
			return condition
		}
	}

	if result == nil || result.Type() == BreakObj || result.Type() == ContinueObj {
		return NULL
	}
	return result
}

func evalForExpression(ctx context.Context, node *ast.ForExpression, s *Scope) Object {
	innerScope := NewScope(s)

	if node.Init != nil {
		init := Eval(ctx, node.Init, innerScope)
		if init.Type() == ErrorObj {
			return init
		}
	}

	var condition Object
	if node.Condition != nil {
		condition = Eval(ctx, node.Condition, innerScope)
		if condition.Type() == ErrorObj {
			return condition
		}
	}

	var result Object
	for isTruthy(condition) {
		newSubScope := NewScope(innerScope)
		result = Eval(ctx, node.Body, newSubScope)

		if result != nil && result.Type() == ErrorObj {
			return result
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			newVal := Eval(ctx, node.Update, newSubScope) //Before continue, we need to call 'Update' and 'Cond'
			if newVal.Type() == ErrorObj {
				return newVal
			}

			condition = Eval(ctx, node.Condition, newSubScope)
			if condition.Type() == ErrorObj {
				return condition
			}

			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			if v.Value != nil {
				//return v.Value
				return v
			}
			break
		}

		if node.Update != nil {
			newVal := Eval(ctx, node.Update, newSubScope)
			if newVal.Type() == ErrorObj {
				return newVal
			}
		}

		if node.Condition != nil {
			condition = Eval(ctx, node.Condition, newSubScope)
			if condition.Type() == ErrorObj {
				return condition
			}
		}
	}

	if result == nil || result.Type() == BreakObj || result.Type() == ContinueObj {
		return NULL
	}
	return result
}

func isTruthy(o Object) bool {
	switch obj := o.(type) {
	case *Boolean:
		return obj.Value
	case *Null:
		return false
	case *Integer:
		if obj.Value == 0 {
			return false
		}
		return true
	case *Float:
		if obj.Value == 0.0 {
			return false
		}
		return true
	default:
		return true
	}
}

func evalIdentifier(node *ast.Identifier, s *Scope) Object {
	key := node.Value
	var v Object
	var ok bool
	if v, ok = s.Get(key); !ok {
		if v, ok = builtins[key]; !ok {
			return newError(UNKNOWNIDENT, key)
		}
	}
	return v
}

func evalTypeofExpression(ctx context.Context, node *ast.TypeofExpression, s *Scope) Object {
	return &String{Value: string(Eval(ctx, node.Expr, s).Type())}
}

func evalAssignExpression(ctx context.Context, node *ast.AssignExpression, s *Scope) Object {
	newValue := Eval(ctx, node.Value, s)
	if newValue.Type() == ErrorObj {
		return newValue
	}
	var name string
	var oldValue Object
	switch nodeType := node.Name.(type) {
	case *ast.Identifier:
		name = nodeType.Value
	case *ast.IndexExpression:
		left := nodeType.Left
		_, ok := left.(*ast.IndexExpression)
		for ok {
			if _, ok = left.(*ast.IndexExpression); ok {
				left = left.(*ast.IndexExpression).Left
			}
		}
		if ident, ok := left.(*ast.Identifier); ok {
			name = ident.Value
		} else {
			oldValue = Eval(ctx, left, s)
		}
	}
	var ok bool
	if oldValue == nil {
		if oldValue, ok = s.Get(name); !ok {
			return newError(UNKNOWNIDENT, name)
		}
	}
	op := node.Token.Literal
	if op == "=" {
		switch nodeType := node.Name.(type) {
		case *ast.Identifier:
			name := nodeType.Value
			v, _ := s.Reset(name, newValue)
			return v
		}
	}
	switch oldValue.Type() {
	case IntegerObj, FloatObj:
		return evalNumberAssignExpression(name, oldValue, newValue, op, s)
	case StringObj:
		if _, ok := node.Name.(*ast.IndexExpression); ok {
			return newErrorf("string is immutable")
		}
		if op == "+=" && newValue.Type() == StringObj {
			v, _ := s.Reset(name, &String{Value: oldValue.(*String).Value + newValue.(*String).Value})
			return v
		}
		return newError(INFIXOP, op, oldValue.Type(), newValue.Type())
	case ArrayObj:
		return evalArrayIndexExpressionFunc(ctx, node.Name.(*ast.IndexExpression), s, op, newValue, func(objects []Object, idx int) Object {
			return updateArray(objects, idx, op, newValue)
		})
	case HashObj:
		return evalHashIndexExpressionFunc(ctx, node.Name.(*ast.IndexExpression), s, op, newValue, func(hash *Hash, key Object) Object {
			return updateHash(hash, key, op, newValue)
		})
	}
	return newError(INFIXOP, op, oldValue.Type(), newValue.Type())
}

func evalNumberAssignExpression(name string, oldValue Object, newValue Object, op string, s *Scope) Object {
	if !isNumber(newValue) {
		return newError(INFIXOP, op, oldValue.Type(), newValue.Type())
	}
	needInt := oldValue.Type() == IntegerObj && newValue.Type() == IntegerObj
	var oldV float64
	if oldValue.Type() == IntegerObj {
		oldV = float64(oldValue.(*Integer).Value)
	} else if oldValue.Type() == FloatObj {
		oldV = oldValue.(*Float).Value
	}
	var newV float64
	if newValue.Type() == IntegerObj {
		newV = float64(newValue.(*Integer).Value)
	} else if newValue.Type() == FloatObj {
		newV = newValue.(*Float).Value
	}
	switch op {
	case "+=":
		if needInt {
			ret, _ := s.Reset(name, &Integer{Value: int64(oldV) + int64(newV)})
			return ret
		}
		ret, _ := s.Reset(name, &Float{Value: oldV + newV})
		return ret
	case "-=":
		if needInt {
			ret, _ := s.Reset(name, &Integer{Value: int64(oldV) - int64(newV)})
			return ret
		}
		ret, _ := s.Reset(name, &Float{Value: oldV - newV})
		return ret
	case "*=":
		if needInt {
			ret, _ := s.Reset(name, &Integer{Value: int64(oldV) * int64(newV)})
			return ret
		}
		ret, _ := s.Reset(name, &Float{Value: oldV * newV})
		return ret
	case "/=":
		if needInt {
			if newV == 0 {
				return newError(DIVIDEBYZERO)
			}
			ret, _ := s.Reset(name, &Integer{Value: int64(oldV) / int64(newV)})
			return ret
		}
		ret, _ := s.Reset(name, &Float{Value: oldV / newV})
		return ret
	default:
		return newError(INFIXOP, op, oldValue.Type(), newValue.Type())
	}
}

func evalPrefixExpression(op string, right Object) Object {
	if right.Type() == ErrorObj {
		return right
	}
	switch op {
	case "!":
		return evalBangExpression(right)
	case "-":
		return evalMinusPrefixExpression(right)
	case "++":
		return evalIncrPrefixExpression(right)
	case "--":
		return evalDecrPrefixExpression(right)
	default:
		return newError(PREFIXOP, op, right.Type())
	}
}

func evalInfixExpression(left Object, op string, right Object) Object {
	if left.Type() == ErrorObj {
		return left
	}
	if right.Type() == ErrorObj {
		return right
	}
	switch {
	case op == "&&":
		return nativeBoolToBooleanObject(isTruthy(left) && isTruthy(right))
	case op == "||":
		return nativeBoolToBooleanObject(isTruthy(left) || isTruthy(right))
	case isNumber(left) && isNumber(right):
		return evalNumberInfixExpression(left, op, right)
	case op == "==":
		return nativeBoolToBooleanObject(evalEquality(left, right))
	case op == "!=":
		return nativeBoolToBooleanObject(!evalEquality(left, right))
	case left.Type() == StringObj || right.Type() == StringObj:
		return evalStringInfixExpression(left, op, right)
	default:
		return newError(INFIXOP, op, left.Type(), right.Type())
	}
}
func evalPostfixExpression(left Object, op string) Object {
	if left.Type() == ErrorObj {
		return left
	}
	switch op {
	case "++":
		return evalIncrPostfixExpression(left)
	case "--":
		return evalDecrPostfixExpression(left)
	default:
		return newError(POSTFIXOP, op, left.Type())
	}
}

func evalEquality(left Object, right Object) bool {
	if left.Type() != right.Type() {
		return false
	}
	switch left.(type) {
	case *Boolean:
		return left.(*Boolean).Value == right.(*Boolean).Value
	case *Null:
		return true
	case *Integer:
		return left.(*Integer).Value == right.(*Integer).Value
	case *Float:
		return left.(*Float).Value == right.(*Float).Value
	case *String:
		return left.(*String).Value == right.(*String).Value
	}
	return false
}

func evalNumberInfixExpression(left Object, op string, right Object) Object {
	needInt := left.Type() == IntegerObj && right.Type() == IntegerObj
	var lv, rv float64
	if i1, ok := left.(*Integer); ok {
		lv = float64(i1.Value)
	} else {
		lv = left.(*Float).Value
	}
	if i2, ok := right.(*Integer); ok {
		rv = float64(i2.Value)
	} else {
		rv = right.(*Float).Value
	}
	switch op {
	case "+":
		if needInt {
			return &Integer{Value: int64(lv + rv)}
		}
		return &Float{Value: lv + rv}
	case "-":
		if needInt {
			return &Integer{Value: int64(lv - rv)}
		}
		return &Float{Value: lv - rv}
	case "*":
		if needInt {
			return &Integer{Value: int64(lv * rv)}
		}
		return &Float{Value: lv * rv}
	case "/":
		if rv == 0 {
			return newError(DIVIDEBYZERO)
		}
		if needInt {
			return &Integer{Value: int64(lv / rv)}
		}
		return &Float{Value: lv / rv}
	case "%":
		mod := math.Mod(lv, rv)
		if needInt {
			return &Integer{Value: int64(mod)}
		}
		return &Float{Value: mod}
	case ">":
		return nativeBoolToBooleanObject(lv > rv)
	case ">=":
		return nativeBoolToBooleanObject(lv >= rv)
	case "<":
		return nativeBoolToBooleanObject(lv < rv)
	case "<=":
		return nativeBoolToBooleanObject(lv <= rv)
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError(INFIXOP, op, left.Type(), right.Type())
	}
}

func evalStringInfixExpression(left Object, op string, right Object) Object {
	switch op {
	case "+":
		return &String{Value: left.String(0) + right.String(0)}
	default:
		return newError(INFIXOP, op, left.Type(), right.Type())
	}
}

func evalIncrPrefixExpression(right Object) Object {
	switch r := right.(type) {
	case *Integer:
		r.Value++
		return &Integer{Value: r.Value}
	case *Float:
		r.Value++
		return &Float{Value: r.Value}
	default:
		return NULL
	}
}

func evalIncrPostfixExpression(left Object) Object {
	switch r := left.(type) {
	case *Integer:
		v := r.Value
		r.Value++
		return &Integer{Value: v}
	case *Float:
		v := r.Value
		r.Value++
		return &Float{Value: v}
	default:
		return NULL
	}
}
func evalDecrPrefixExpression(right Object) Object {
	switch r := right.(type) {
	case *Integer:
		r.Value--
		return &Integer{Value: r.Value}
	case *Float:
		r.Value--
		return &Float{Value: r.Value}
	default:
		return NULL
	}
}
func evalDecrPostfixExpression(left Object) Object {
	switch r := left.(type) {
	case *Integer:
		v := r.Value
		r.Value--
		return &Integer{Value: v}
	case *Float:
		v := r.Value
		r.Value--
		return &Float{Value: v}
	default:
		return NULL
	}
}

func evalMinusPrefixExpression(right Object) Object {
	switch r := right.(type) {
	case *Integer:
		return &Integer{Value: -r.Value}
	case *Float:
		return &Float{Value: -r.Value}
	default:
		return NULL
	}
}

func evalBangExpression(right Object) Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		switch r := right.(type) {
		case *Integer:
			if r.Value == 0 {
				return TRUE
			}
			return FALSE
		case *Float:
			if r.Value == 0 {
				return TRUE
			}
			return FALSE
		default:
			return FALSE
		}
	}
}

func evalCallExpression(ctx context.Context, node *ast.CallExpression, s *Scope) Object {
	function := Eval(ctx, node.Function, s)
	if function.Type() == ErrorObj {
		return function
	}

	args := evalExpressions(ctx, node.Arguments, s)
	if len(args) == 1 && args[0].Type() == ErrorObj {
		return args[0]
	}
	return applyFunction(ctx, function, args)
}

func evalExpressions(ctx context.Context, expressions []ast.Expression, s *Scope) []Object {
	results := make([]Object, 0)
	for _, argument := range expressions {
		result := Eval(ctx, argument, s)
		if result.Type() == ErrorObj {
			return []Object{result}
		}
		results = append(results, result)
	}
	return results
}

func applyFunction(ctx context.Context, funcObj Object, args []Object) Object {
	switch function := funcObj.(type) {
	case *Function:
		sub := NewScope(function.Scope)
		for i, param := range function.Parameters {
			sub.Set(param.Value, args[i])
		}
		result := Eval(ctx, function.Body, sub)
		if rv, ok := result.(*ReturnValue); ok {
			return rv.Value
		}
		return result
	case *Builtin:
		return function.Fn(args...)
	}
	return newError(NOTFUNC, funcObj.String(0))
}

func evalArrayLiteral(ctx context.Context, node *ast.ArrayLiteral, s *Scope) *Array {
	arr := &Array{}
	elements := make([]Object, 0)
	for _, expr := range node.Elements {
		elements = append(elements, Eval(ctx, expr, s))
	}
	arr.Elements = elements
	return arr
}

func evalIndexExpression(ctx context.Context, node *ast.IndexExpression, s *Scope) Object {
	e := Eval(ctx, node.Left, s)
	if sliceExpr, ok := node.Index.(*ast.SliceExpression); ok {
		return evalSliceExpression(ctx, e, sliceExpr, s)
	} else {
		switch e.(type) {
		case *Array:
			return evalArrayIndexExpressionFunc(ctx, node, s, "", nil, func(arr []Object, idx int) Object { return arr[idx] })
		case *String:
			return evalStringIndexExpression(ctx, node, s)
		case *Hash:
			return evalHashIndexExpressionFunc(ctx, node, s, "", nil, func(hash *Hash, key Object) Object {
				if hashable, ok := key.(Hashable); ok {
					if kv, ok := hash.Pairs[hashable.HashKey()]; ok {
						return kv.Value
					} else {
						return NULL
					}
				} else {
					return newError(NOTHASHABLE, key.Type())
				}
			})
		default:
			return newError(NOINDEXERROR, e.Type())
		}
	}
}

func evalStringIndexExpression(ctx context.Context, node *ast.IndexExpression, s *Scope) Object {
	left := Eval(ctx, node.Left, s)
	if left.Type() == ErrorObj {
		return left
	}
	index := Eval(ctx, node.Index, s)
	if index.Type() == ErrorObj {
		return index
	}
	str := left.(*String)
	length := len(str.Value)
	if i, e := calcIndex(length, index, false); e != nil {
		return e
	} else {
		return &String{Value: string(str.Value[i])}
	}
}

func evalArrayIndexExpressionFunc(ctx context.Context, node *ast.IndexExpression, s *Scope, op string, newValue Object, f func([]Object, int) Object) Object {
	left := Eval(ctx, node.Left, s)
	if left.Type() == ErrorObj {
		return left
	}
	index := Eval(ctx, node.Index, s)
	if index.Type() == ErrorObj {
		return index
	}
	switch l := left.(type) {
	case *Array:
		length := len(l.Elements)
		if i, e := calcIndex(length, index, false); e != nil {
			return e
		} else {
			return f(l.Elements, i)
		}
	case *String:
		return newErrorf("string is immutable")
	case *Hash:
		return updateHash(l, index, op, newValue)
	default:
		return newErrorf("%s is not a hash", left.String(0))
	}
}
func updateArray(objects []Object, idx int, op string, newValue Object) Object {
	if op == "=" {
		objects[idx] = newValue
		return newValue
	}
	switch old := objects[idx].(type) {
	case *Integer:
		switch op {
		case "+=":
			if newValue.Type() == IntegerObj {
				old.Value += newValue.(*Integer).Value
				return old
			} else if newValue.Type() == FloatObj {
				objects[idx] = &Float{Value: float64(old.Value) + newValue.(*Float).Value}
				return objects[idx]
			}
			return newError(INFIXOP, op, old.Type(), newValue.Type())
		case "-=":
			if newValue.Type() == IntegerObj {
				old.Value -= newValue.(*Integer).Value
				return old
			} else if newValue.Type() == FloatObj {
				objects[idx] = &Float{Value: float64(old.Value) - newValue.(*Float).Value}
				return objects[idx]
			}
			return newError(INFIXOP, op, old.Type(), newValue.Type())
		case "*=":
			if newValue.Type() == IntegerObj {
				old.Value *= newValue.(*Integer).Value
				return old
			} else if newValue.Type() == FloatObj {
				objects[idx] = &Float{Value: float64(old.Value) * newValue.(*Float).Value}
				return objects[idx]
			}
			return newError(INFIXOP, op, old.Type(), newValue.Type())
		case "/=":
			if newValue.Type() == IntegerObj {
				if newValue.(*Integer).Value == 0 {
					return newError(DIVIDEBYZERO)
				}
				old.Value /= newValue.(*Integer).Value
				return old
			} else if newValue.Type() == FloatObj {
				if newValue.(*Float).Value == 0 {
					return newError(DIVIDEBYZERO)
				}
				objects[idx] = &Float{Value: float64(old.Value) / newValue.(*Float).Value}
				return objects[idx]
			}
			return newError(INFIXOP, op, old.Type(), newValue.Type())
		}
	case *Float:
		switch op {
		case "+=":
			if newValue.Type() == IntegerObj {
				old.Value += float64(newValue.(*Integer).Value)
			} else if newValue.Type() == FloatObj {
				old.Value += newValue.(*Float).Value
			}
			return old
		case "-=":
			if newValue.Type() == IntegerObj {
				old.Value -= float64(newValue.(*Integer).Value)
			} else if newValue.Type() == FloatObj {
				old.Value -= newValue.(*Float).Value
			}
			return old
		case "*=":
			if newValue.Type() == IntegerObj {
				old.Value *= float64(newValue.(*Integer).Value)
			} else if newValue.Type() == FloatObj {
				old.Value *= newValue.(*Float).Value
			}
			return old
		case "/=":
			if newValue.Type() == IntegerObj {
				if newValue.(*Integer).Value == 0 {
					return newError(DIVIDEBYZERO)
				}
				old.Value /= float64(newValue.(*Integer).Value)
			} else if newValue.Type() == FloatObj {
				if newValue.(*Float).Value == 0 {
					return newError(DIVIDEBYZERO)
				}
				old.Value /= newValue.(*Float).Value
			}
			return old
		}
	case *String:
		switch op {
		case "+=":
			old.Value += newValue.String(0)
			return old
		}
	}
	return newError(INFIXOP, op, objects[idx].Type(), newValue.Type())
}

func updateHash(h *Hash, k Object, op string, newValue Object) Object {
	key, hash := k, h
	var hashkey HashKey
	if k, ok := key.(Hashable); !ok {
		return newError(NOTHASHABLE, key.Type())
	} else {
		hashkey = k.HashKey()
	}
	if op == "=" {
		hash.Pairs[hashkey] = HashPair{Key: key, Value: newValue}
		return newValue
	}
	switch old := hash.Pairs[hashkey].Value.(type) {
	case *Integer:
		switch op {
		case "+=":
			if newValue.Type() == IntegerObj {
				old.Value += newValue.(*Integer).Value
				return old
			} else if newValue.Type() == FloatObj {
				hash.Pairs[hashkey] = HashPair{Key: key, Value: &Float{Value: float64(old.Value) + newValue.(*Float).Value}}
				return hash.Pairs[hashkey].Value
			}
			return newError(INFIXOP, op, old.Type(), newValue.Type())
		case "-=":
			if newValue.Type() == IntegerObj {
				old.Value -= newValue.(*Integer).Value
				return old
			} else if newValue.Type() == FloatObj {
				hash.Pairs[hashkey] = HashPair{Key: key, Value: &Float{Value: float64(old.Value) - newValue.(*Float).Value}}
				return hash.Pairs[hashkey].Value
			}
			return newError(INFIXOP, op, old.Type(), newValue.Type())
		case "*=":
			if newValue.Type() == IntegerObj {
				old.Value *= newValue.(*Integer).Value
				return old
			} else if newValue.Type() == FloatObj {
				hash.Pairs[hashkey] = HashPair{Key: key, Value: &Float{Value: float64(old.Value) * newValue.(*Float).Value}}
				return hash.Pairs[hashkey].Value
			}
			return newError(INFIXOP, op, old.Type(), newValue.Type())
		case "/=":
			if newValue.Type() == IntegerObj {
				if newValue.(*Integer).Value == 0 {
					return newError(DIVIDEBYZERO)
				}
				old.Value /= newValue.(*Integer).Value
				return old
			} else if newValue.Type() == FloatObj {
				if newValue.(*Float).Value == 0 {
					return newError(DIVIDEBYZERO)
				}
				hash.Pairs[hashkey] = HashPair{Key: key, Value: &Float{Value: float64(old.Value) / newValue.(*Float).Value}}
				return hash.Pairs[hashkey].Value
			}
			return newError(INFIXOP, op, old.Type(), newValue.Type())
		}
	case *Float:
		switch op {
		case "+=":
			if newValue.Type() == IntegerObj {
				old.Value += float64(newValue.(*Integer).Value)
			} else if newValue.Type() == FloatObj {
				old.Value += newValue.(*Float).Value
			}
			return old
		case "-=":
			if newValue.Type() == IntegerObj {
				old.Value -= float64(newValue.(*Integer).Value)
			} else if newValue.Type() == FloatObj {
				old.Value -= newValue.(*Float).Value
			}
			return old
		case "*=":
			if newValue.Type() == IntegerObj {
				old.Value *= float64(newValue.(*Integer).Value)
			} else if newValue.Type() == FloatObj {
				old.Value *= newValue.(*Float).Value
			}
			return old
		case "/=":
			if newValue.Type() == IntegerObj {
				if newValue.(*Integer).Value == 0 {
					return newError(DIVIDEBYZERO)
				}
				old.Value /= float64(newValue.(*Integer).Value)
			} else if newValue.Type() == FloatObj {
				if newValue.(*Float).Value == 0 {
					return newError(DIVIDEBYZERO)
				}
				old.Value /= newValue.(*Float).Value
			}
			return old
		}
	case *String:
		switch op {
		case "+=":
			old.Value += newValue.String(0)
			return old
		}
	}
	return newError(INFIXOP, op, hash.Pairs[hashkey].Value.Type(), newValue.Type())
}

func evalHashIndexExpressionFunc(ctx context.Context, node *ast.IndexExpression, s *Scope, op string, newValue Object, f func(hash *Hash, key Object) Object) Object {
	left := Eval(ctx, node.Left, s)
	if left.Type() == ErrorObj {
		return left
	}
	index := Eval(ctx, node.Index, s)
	if index.Type() == ErrorObj {
		return index
	}
	switch l := left.(type) {
	case *String:
		return newErrorf("string is immutable")
	case *Hash:
		return f(l, index)
	case *Array:
		objects := l.Elements
		if idx, ok := index.(*Integer); ok {
			return updateArray(objects, int(idx.Value), op, newValue)
		} else {
			return newError(INDEXINT)
		}

	default:
		return newErrorf("%s is not a hash", left.String(0))
	}
}

func evalSliceExpression(ctx context.Context, obj Object, sliceExpr *ast.SliceExpression, s *Scope) Object {
	var l int
	switch obj.(type) {
	case *Array:
		l = len(obj.(*Array).Elements)
	case *String:
		l = len(obj.(*String).Value)
	default:
		return newError(NOINDEXERROR, obj.Type())
	}
	start := Eval(ctx, sliceExpr.Start, s)
	if start.Type() == ErrorObj {
		return start
	}
	startIdx, e := calcIndex(l, start, false) //start: 0~len-1
	if e != nil {
		return e
	}

	var endIdx int
	if sliceExpr.End != nil {
		end := Eval(ctx, sliceExpr.End, s)
		if end.Type() == ErrorObj {
			return end
		}
		endIdx, e = calcIndex(l, end, true) //end: 0~len
		if e != nil {
			return e
		}
	} else {
		endIdx = l
	}

	if startIdx > endIdx {
		return newError(SLICEERROR, startIdx, endIdx)
	}
	switch obj.(type) {
	case *Array:
		return &Array{Elements: obj.(*Array).Elements[startIdx:endIdx]}
	case *String:
		return &String{Value: obj.(*String).Value[startIdx:endIdx]}
	default:
		return newError(NOINDEXERROR, obj.Type())
	}
}

func calcIndex(max int, idxExpr Object, end bool) (int, *Error) {
	if _, ok := idxExpr.(*Integer); !ok {
		return 0, newError(INDEXINT).(*Error)
	}
	idx := int(idxExpr.(*Integer).Value)
	if end {
		max++
	}
	originIdx := idx
	if idx < 0 {
		idx += max
	}
	if idx >= max {
		return idx, newError(INDEXERROR, originIdx, 0, max-1).(*Error)
	}
	if idx < 0 {
		return idx, newError(INDEXERROR, originIdx, -1, -max).(*Error)
	}
	return idx, nil
}

func evalMethodCallExpression(ctx context.Context, node *ast.MethodCallExpression, s *Scope) Object {
	obj := Eval(ctx, node.Object, s)
	if method, ok := node.Call.(*ast.CallExpression); ok {
		args := evalExpressions(ctx, method.Arguments, s)
		return obj.CallMethod(method.Function.String(), args...)
	}
	return newError(NOMETHODERROR, node.String(), obj.Type())
}

func evalHashLiteral(ctx context.Context, node *ast.HashLiteral, s *Scope) Object {
	hashMap := make(map[HashKey]HashPair)
	for key, value := range node.Pairs {
		var k Object
		if ident, ok := key.(*ast.Identifier); ok {
			k = &String{Value: ident.Value}
		} else {
			k = Eval(ctx, key, s)
		}
		if hashable, ok := k.(Hashable); ok {
			hashMap[hashable.HashKey()] = HashPair{Key: k, Value: Eval(ctx, value, s)}
		} else {
			return newError(NOTHASHABLE, k.Type())
		}
	}
	return &Hash{Pairs: hashMap}
}
