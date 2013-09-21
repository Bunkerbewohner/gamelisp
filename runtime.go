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
	default:
		return code
	}
}

func CreateMainContext() *Context {
	context := NewContext()
	context.symbols["type"] = NativeFunction{_type}

	return context
}

var MainContext = CreateMainContext()
