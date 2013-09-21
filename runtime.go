package main

type Context struct {
	symbols map[string]Data
}

func NewContext() *Context {
	return &Context{
		make(map[string]Data),
	}
}

func Evaluate(code Data, context *Context) Data {
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
					return fn.Call(t, context)
				}
			}
		}

		panic("Expression is not a valid function invocation: " + code.String())
	case Symbol:
		if value, ok := context.symbols[t.Value]; ok {
			return value
		} else {
			return t
		}

	default:
		return code
	}
}

func CreateMainContext() *Context {
	context := NewContext()
	context.symbols["type"] = NativeFunction{_type}
	context.symbols["def"] = NativeFunction{def}

	return context
}

var MainContext = CreateMainContext()
