package main

import "fmt"

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

// data type for user-defined gamelisp functions
type Function struct {
	Name        string
	Dispatchers []DispatchPattern
	Code        List
}

func (f Function) String() string {
	return "NativeFunction<" + f.Name + ">"
}

type ParameterDeclaration interface {
	ParameterName() string
	Match(arg Data) bool
	Bind(args List, index int, context *Context)
}

type DispatchPattern struct {
	Parameters []ParameterDeclaration
}

// Attempts to match a list of arguments to this dispatch pattern
// ignores the first list item, assuming it's the function name
func (dp DispatchPattern) Match(args List) bool {
	if len(dp.Parameters) > args.Len() {
		panic(fmt.Sprintf("Expected at least %n parameters, found only %n", len(dp.Parameters), args.Len()))
	}

	i := 1

	for e := args.Front().Next(); e != nil && i < len(dp.Parameters); e = e.Next() {
		arg := e.Value.(Data)
		if !dp.Parameters[i].Match(arg) {
			return false
		}
		i++
	}

	// Are there still arguments left that were not matched?
	if i < args.Len() {
		// Check if the last parameter is a sink
		if _, isSink := dp.Parameters[i-1].(ArgumentSink); isSink {
			return true
		} else {
			return false
		}
	}

	return true
}

// pattern for a regular function argument
type ArgumentPattern struct {
	// Name of the parameter which will be created as a symbol in the function
	Name string

	// The expected type or nil if any type is accepted
	ExpectedType *DataType

	// The expected value or nil if any value is accepted
	ExpectedValue *Data
}

func (ap ArgumentPattern) Bind(args List, index int, context *Context) {
	context.Define(Symbol{ap.Name}, args.Get(index))
}

func (ap ArgumentPattern) Match(param Data) bool {
	if ap.ExpectedValue != nil {
		return param.Equals(*ap.ExpectedValue)
	}

	if ap.ExpectedType != nil {
		return param.GetType().Equals(ap.ExpectedType)
	}

	return false
}

func (ap ArgumentPattern) ParameterName() string {
	return ap.Name
}

// placeholder for a argument sink that consumes all following arguments passed to the function
type ArgumentSink struct {
	Name string
}

func (as ArgumentSink) Bind(args List, index int, context *Context) {
	// binds arguments i and all following as a list
	context.Define(Symbol{as.Name}, args.SliceFrom(index))
}

func (as ArgumentSink) Match(param Data) bool {
	return true
}

func (as ArgumentSink) ParameterName() string {
	return as.Name
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

// Expects code for a function definition, e.g.
// (defn myfunction [p1 p2 ...] (stmts*))
func CreateFunction(code List, context *Context) *Function {
	// code.Get(0) == "defn"

	fn := new(Function)

	// function name
	if name, ok := code.Get(1).(Symbol); ok {
		fn.Name = name.Value
	} else {
		panic("First argument must be a symbol")
	}

	// create dispatch pattern
	if list, ok := code.Get(2).(List); ok {
		count := list.Len()
		dispatcher := new(DispatchPattern)
		dispatcher.Parameters = make([]ParameterDeclaration, count)
		for i := 0; i < count; i++ {
			dispatcher.Parameters[i] = CreateParameter(list.Get(i), context)
		}

		fn.Dispatchers = make([]DispatchPattern, 1)
		fn.Dispatchers[0] = *dispatcher
	} else {
		panic("Second argument must be a list")
	}

	// Save the code for later execution
	if list, ok := code.Get(3).(List); ok {
		fn.Code = list
	} else {
		panic("Third argument must be a list")
	}

	return fn
}

func CreateParameter(code Data, context *Context) ParameterDeclaration {
	switch t := code.(type) {
	case Symbol:
		return ArgumentPattern{
			Name: t.Value,
		}
	}

	return nil
}

func (fn Function) selectDispatch(code List) *DispatchPattern {
	for _, dispatch := range fn.Dispatchers {
		if dispatch.Match(code) {
			return &dispatch
		}
	}

	return nil
}

func (dp DispatchPattern) bindParameters(code List, context *Context) {
	for i, decl := range dp.Parameters {
		decl.Bind(code, i, context)
	}
}

// ($name args...)
func (fn Function) Call(code List, env *Context) Data {
	// create the function context for this call
	context := NewContext()
	context.parent = env

	// choose dispatcher
	dispatch := fn.selectDispatch(code)
	dispatch.bindParameters(code, context)

	// execute the code in the temporary context
	result, err := Evaluate(fn.Code, context)
	if err != nil {
		panic(fmt.Sprintf("Failed to call %s: %s", fn.Name, err))
	}

	return result
}
