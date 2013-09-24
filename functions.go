package main

type Caller interface {
	// Call this function in it's written form, i.e. with a list of expressions, where the first one is the function name
	Call(code List, context *Context) Data
}

type NativeFunction struct {
	// Native functions receive a list of arguments, the first being the name under
	// which the function itself was called
	Function func(List, *Context) Data
}

type NativeFunctionB struct {
	// Same as NativeFunction, except it doesnt expect evaluated arguments
	Function func(List, *Context) Data
}

type Function struct {
}

func (fn NativeFunction) Call(code List, context *Context) Data {
	args := code.Map(__evalArgs(context))
	return fn.Function(args, context)
}

func (fn NativeFunction) String() string {
	return "native function"
}

func (fn NativeFunctionB) Call(code List, context *Context) Data {
	return fn.Function(code, context)
}

func (fn NativeFunctionB) String() string {
	return "native function"
}
