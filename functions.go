package main

import "fmt"

type Caller interface {
	// Call this function in it's written form, i.e. with a list of expressions, where the first one is the function name
	Call(args List, context *Context) Data
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
}

func (f Function) String() string {
	return fmt.Sprintf("Function<%s>", f.Name)
}

func (f *Function) AddDispatch(dp *DispatchPattern) {
	index := -1

	// check if the pattern already exists
	for i, dispatcher := range f.Dispatchers {
		if dispatcher.Equals(*dp) {
			index = i
			break
		}
	}

	if index >= 0 {
		f.Dispatchers[index] = *dp
	} else {
		f.Dispatchers = append(f.Dispatchers, *dp)
	}
}

type ParameterDeclaration interface {
	ParameterName() string
	Match(arg Data) bool
	Bind(args List, index int, context *Context)
	Equals(other ParameterDeclaration) bool
}

type DispatchPattern struct {
	Parameters []ParameterDeclaration
	Code       Data
}

// Two dispatch patterns are equal if their patterns are equivalent (Code is not considered)
func (dp DispatchPattern) Equals(other DispatchPattern) bool {
	if len(dp.Parameters) != len(other.Parameters) {
		return false
	}

	for i, param := range dp.Parameters {
		if !param.Equals(other.Parameters[i]) {
			return false
		}
	}

	return true
}

// Attempts to match a list of arguments to this dispatch pattern
// ignores the first list item, assuming it's the function name
func (dp DispatchPattern) Match(args List) bool {
	if len(dp.Parameters) > args.Len() {
		return false
	}

	i := 0
	for e := args.Front(); e != nil && i < len(dp.Parameters); e = e.Next() {
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
	ExpectedValue Data
}

func (ap ArgumentPattern) Equals(other ParameterDeclaration) bool {
	if otherAp, ok := other.(ArgumentPattern); ok {
		types := (ap.ExpectedType == nil && otherAp.ExpectedType == nil) || ap.ExpectedType.Equals(otherAp.ExpectedType)
		values := (ap.ExpectedValue == nil && otherAp.ExpectedValue == nil) || ap.ExpectedValue.Equals(otherAp.ExpectedValue)
		return types && values
	}

	return false
}

func (ap ArgumentPattern) Bind(args List, index int, context *Context) {
	if ap.Name != "" {
		context.Define(Symbol{ap.Name}, args.Get(index))
	} else {
		// nothing to bind since the expected value is known to the user
	}
}

func (ap ArgumentPattern) Match(param Data) bool {
	// check for a required value
	if ap.ExpectedValue != nil {
		return param.Equals(ap.ExpectedValue)
	}

	// check for a required type
	if ap.ExpectedType != nil {
		paramType := param.GetType()
		equal := paramType.Equals(*ap.ExpectedType)
		return equal
	}

	// if no type or value has been specified anything is accepted
	return true
}

func (ap ArgumentPattern) ParameterName() string {
	return ap.Name
}

// placeholder for a argument sink that consumes all following arguments passed to the function
type ArgumentSink struct {
	Name string
}

func (as ArgumentSink) Equals(pd ParameterDeclaration) bool {
	if _, ok := pd.(ArgumentSink); ok {
		return true
	}

	return false
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

func (fn NativeFunction) Call(args List, context *Context) Data {
	args = args.Map(__evalArgs(context))
	return fn.Function(args, context)
}

func (fn NativeFunction) String() string {
	return "native function"
}

func (fn NativeFunctionB) Call(args List, context *Context) Data {
	return fn.Function(args, context)
}

func (fn NativeFunctionB) String() string {
	return "native function"
}

// Expects args for a function definition, e.g.
// (defn myfunction [p1 p2 ...] (stmts*))
func CreateFunction(args List, context *Context) *Function {
	ValidateArgs(args, []string{"Symbol", "List", "Data"})
	fn := new(Function)

	// function name
	fn.Name = args.First().(Symbol).Value

	// create dispatch pattern
	fnArgs := args.Second().(List)
	if first, isSymbol := fnArgs.Get(0).(Symbol); isSymbol {
		if first.Value == "list" {
			fnArgs = fnArgs.SliceFrom(1)
		}
	}

	count := fnArgs.Len()
	dispatcher := new(DispatchPattern)
	dispatcher.Parameters = make([]ParameterDeclaration, count)
	for i := 0; i < count; i++ {
		dispatcher.Parameters[i] = CreateParameter(fnArgs.Get(i), context)
	}
	dispatcher.Code = args.Third()

	fn.Dispatchers = make([]DispatchPattern, 1)
	fn.Dispatchers[0] = *dispatcher

	return fn
}

func CreateParameter(args Data, context *Context) ParameterDeclaration {
	switch t := args.(type) {
	case Symbol:
		// check if the symbol refers to a datatype
		def := context.LookUp(t)
		if def != nil {
			if datatype, ok := def.(DataType); ok {
				return ArgumentPattern{ExpectedType: &datatype}
			}
		}

		return ArgumentPattern{Name: t.Value}
	case Int, Float, Bool, String, Keyword, Nothing:
		return ArgumentPattern{ExpectedValue: t}
	}

	panic(fmt.Sprintf("Couldn't create parameter from %s", args.String()))
}

func (fn Function) selectDispatch(args List) *DispatchPattern {
	for _, dispatch := range fn.Dispatchers {
		if dispatch.Match(args) {
			return &dispatch
		}
	}

	return nil
}

func (dp DispatchPattern) bindParameters(args List, context *Context) {
	for i, decl := range dp.Parameters {
		decl.Bind(args, i, context)
	}
}

// ($name args...)
func (fn Function) Call(args List, env *Context) Data {
	// evaluate the arguments
	args = args.Map(__evalArgs(env))

	// create the function context for this call
	context := NewContext()
	context.parent = env

	// choose dispatcher
	dispatch := fn.selectDispatch(args)
	if dispatch == nil {
		panic("No dispatch pattern matches the given function arguments")
	}
	dispatch.bindParameters(args, context)

	// execute the args in the temporary context
	result, err := Evaluate(dispatch.Code, context)
	if err != nil {
		panic(fmt.Sprintf("Failed to call %s: %s", fn.Name, err))
	}

	return result
}
