package evaluator

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var builtins = map[string]*Builtin{
	"len": {func(args ...Object) Object {
		if len(args) != 1 {
			return newError(ARGUMENTNUMERROR, "1", len(args))
		}
		switch iterable := args[0].(type) {
		case *String:
			return &Integer{Value: int64(len(iterable.Value))}
		case *Array:
			return &Integer{Value: int64(len(iterable.Elements))}
		case *Hash:
			return &Integer{Value: int64(len(iterable.Pairs))}
		default:
			return newError(ARGUMENTTYPEERROR, "STRING", args[0].Type())
		}
	}},
	"number": {
		func(args ...Object) Object {
			if len(args) != 1 {
				return newError(ARGUMENTNUMERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *Integer, *Float:
				return input
			case *String:
				n, err := strconv.Atoi(input.Value)
				if err != nil {
					float, err := strconv.ParseFloat(input.Value, 64)
					if err != nil {
						return newErrorf("%s is not a number", input.Value)
					}
					return &Float{Value: float}
				}
				return &Integer{Value: int64(n)}
			}
			return newError(ARGUMENTTYPEERROR, "STRING", args[0].Type())
		},
	},
	"string": {
		func(args ...Object) Object {
			if len(args) != 1 {
				return newError(ARGUMENTNUMERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *String:
				return input
			default:
				return &String{Value: input.String(0)}
			}
		},
	},
	"int": {
		func(args ...Object) Object {
			if len(args) != 1 {
				return newError(ARGUMENTNUMERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *Integer:
				return input
			case *Float:
				return &Integer{Value: int64(input.Value)}
			case *Boolean:
				if input.Value {
					return &Integer{Value: 1}
				}
				return &Integer{Value: 0}
			case *String:
				i, err := strconv.Atoi(input.Value)
				if err != nil {
					return newErrorf("%s is not an integer", input.Value)
				}
				return &Integer{Value: int64(i)}
			default:
				return &String{Value: input.String(0)}
			}
		},
	},
	"now": {
		func(args ...Object) Object {
			if len(args) != 0 {
				return newError(ARGUMENTNUMERROR, "0", len(args))
			}
			return &String{Value: time.Now().Format("2006-01-02 15:04:05")}
		},
	},
	"print": {
		func(args ...Object) Object {
			strs := make([]string, 0)
			for _, arg := range args {
				strs = append(strs, arg.String(0))
			}
			fmt.Println(strings.Join(strs, ", "))
			return NULL
		},
	},
}
