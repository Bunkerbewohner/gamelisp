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
	switch t := code.(type) {
	case List:
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
				}
			}
		}

		return nil, errors.New(fmt.Sprintf("Not a function: %s\n", t.Get(0)))
	case Symbol:
		if value, ok := context.symbols[t.Value]; ok {
			return value, nil
		} else {
			return t, nil
		}

	default:
		return code, nil
	}
}

func CreateMainContext() *Context {
	context := NewContext()
	context.symbols["type"] = NativeFunction{_type}
	context.symbols["def"] = NativeFunction{def}

	return context
}

var MainContext = CreateMainContext()
