package main

import "errors"
import "fmt"

type Context struct {
	symbols map[string]Data
}

func NewContext() *Context {
	return &Context{
		make(map[string]Data),
	}
}

func Evaluate(code Data, context *Context) (Data, error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%v in %v", e, code)
		}
	}()

	switch t := code.(type) {
	case List:
		if t.evaluated {
			// if the list was already evaluated just return its contents as is
			return code, nil
		} else if t.Len() == 0 {
			return nil, errors.New("invalid function invocation")
		}

		// first expression must be a symbol
		symbol, ok := t.Front().Value.(Symbol)
		if ok {
			// look up the value for that symbol
			fn, ok := context.symbols[symbol.Value]
			if ok {
				// check if we can call it as a function
				fn, ok := fn.(Caller)
				if ok {
					return fn.Call(t, context), nil
				} else {
					return nil, errors.New(fmt.Sprintf("%s is not a function", t.Get(0)))
				}
			} else {
				return nil, errors.New(fmt.Sprintf("%s is not defined", t.Get(0)))
			}
		} else {
			return nil, errors.New(fmt.Sprintf("%s is not a symbol", t.Get(0)))
		}
	case Keyword:
		return t, nil
	case Symbol:
		// look up the symbol and returns its value
		if value, ok := context.symbols[t.Value]; ok {
			return value, nil
		} else {
			return nil, errors.New(fmt.Sprintf("%s is not defined", t))
		}
	}

	return code, nil
}

func CreateMainContext() *Context {
	context := NewContext()
	context.symbols["def"] = NativeFunctionB{_def}
	context.symbols["type"] = NativeFunction{_type}

	context.symbols["symbol"] = NativeFunction{_symbol}
	context.symbols["keyword"] = NativeFunction{_keyword}
	context.symbols["list"] = NativeFunction{_list}
	context.symbols["dict"] = NativeFunction{_dict}

	context.symbols["print"] = NativeFunction{_print}

	context.symbols["foreach"] = NativeFunction{_foreach}
	context.symbols["map"] = NativeFunction{_map}

	context.symbols["get"] = NativeFunction{_get}

	return context
}

var MainContext = CreateMainContext()
