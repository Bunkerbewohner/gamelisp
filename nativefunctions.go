package main

//
// This file contains elemental functions defined natively in Go,
// which all user-defined args and extra libraries are based on
//

import "fmt"
import "reflect"

//
// args List is the expression list handed to the evaluator,
// the first element being the function name, all following
// elements are arguments to the denoted function
//

// Checks argument types against a variable number of allowed signatures
// Each signature is represented as a list of type names
func ValidateArgs(args List, expected ...[]string) {
	valid := false

	for _, checks := range expected {
		partvalid := true

		// if there are more checked arguments then provided ones, it cannot be a valid call
		if len(checks) != args.Len() {
			continue
		}

		// compare the required types for each argument
		for i, t := range checks {
			argtype := reflect.TypeOf(args.GetElement(i).Value)
			if argtype.Name() != t && t != "Data" {
				partvalid = false
				break
			}
		}

		valid = valid || partvalid
	}

	if !valid {
		msg := "Invalid arguments\nExpected: "
		for i, checks := range expected {
			if i > 0 {
				msg += " or "
			}
			msg += "("
			for j, t := range checks {
				if j > 0 {
					msg += " "
				}
				msg += t
			}
			msg += ")"
		}

		msg += "\nFound: ("
		for e := args.Front(); e != nil; e = e.Next() {
			if e != args.Front() {
				msg += " "
			}
			msg += reflect.TypeOf(e.Value).Name()
		}
		msg += ")"

		panic(msg)
	}
}

// returns a function that evaluates all arguments
func __evalArgs(context *Context) func(data Data, i int) Data {
	return func(data Data, i int) Data {
		result, err := Evaluate(data, context)
		if err != nil {
			panic(err.Error())
		}
		return result
	}
}

// (type x) - Returns the type of x as a string
func _type(args List, context *Context) Data {
	args.RequireArity(1)
	return args.First().GetType()
}

// (def symbol value) - Defines a new symbol and assigns the value
func _def(args List, context *Context) Data {
	ValidateArgs(args, []string{"Symbol", "Data"})

	// the symbol referring to the defined value
	symbol := args.First().(Symbol)

	// get the value that shall be associated to the symbol
	value := args.Second()
	value, err := Evaluate(value, context)

	if err == nil {
		context.Define(symbol, value)
	} else {
		fmt.Printf(err.Error())
		return nil
	}

	return value
}

// (fn [name] args* stmts*)
func _fn(args List, context *Context) Data {
	ValidateArgs(args, []string{"Symbol", "List", "Data"}, []string{"List", "Data"})

	if args.Len() < 3 {
		args.InsertBefore(Symbol{"anonymous"}, args.Front())
	}

	return CreateFunction(args, context)
}

// (defn name args* stmts*)
func _defn(args List, context *Context) Data {
	ValidateArgs(args, []string{"Symbol", "List", "Data"})

	name := args.First().(Symbol)
	fn := CreateFunction(args, context)

	def := context.LookUp(name)
	if def == nil {
		// define new function
		context.Define(name, fn)
	} else {
		// prevent native functions from being overwritten
		if _, ok := def.(NativeFunction); ok {
			panic(fmt.Sprintf("%s (%s) cannot be overwritten", def.String(), def.GetType().String()))
		}

		// just overwrite anything else
		context.Define(name, fn)
	}

	return fn
}

// (fn+= name args* stmts*)
func _extend_function(args List, context *Context) Data {
	ValidateArgs(args, []string{"Symbol", "List", "Data"})
	name := args.First().(Symbol)
	fn := CreateFunction(args, context)
	def := context.LookUp(name)
	if def == nil {
		context.Define(name, fn)
		return Nothing{}
	}
	if basisFn, ok := def.(*Function); ok {
		for _, dispatch := range fn.Dispatchers {
			basisFn.AddDispatch(&dispatch)
		}
	} else {
		panic(fmt.Sprintf("%s is not a function", def.String()))
	}

	return Nothing{}
}

// (list x1 x2 ...)
func _list(args List, context *Context) Data {
	args.evaluated = true
	return args
}

// creates a dictionary. expects even number of arguments
// (dict :key1 value1 :key2 value2 ...)
func _dict(args List, context *Context) Data {
	dict := CreateDict()

	if args.Len()%2 == 1 {
		panic("Dictionary requires an even number of arguments")
	}

	for e := args.Front(); e != nil; e = e.Next() {
		key, _ := e.Value.(Data)
		value, _ := e.Next().Value.(Data)
		e = e.Next()

		dict.entries[key] = value
	}

	return dict
}

// (symbol name) - return a symbol with given name
func _symbol(args List, context *Context) Data {
	ValidateArgs(args, []string{"Symbol"})
	return args.First().(String)
}

// (keyword name) - return a keyword with given name (prepended with a colon if not supplied)
func _keyword(args List, context *Context) Data {
	ValidateArgs(args, []string{"String"})

	str := args.Second().(String)
	if str.Value[0] != ':' {
		str.Value = ":" + str.Value
	}

	return Keyword{str.Value}
}

// (get dict key)	- gets an entry from a dictionary
// (get list index)	- gets an entry from a list
// returns Nothing if entry doesn't exist
func _get(args List, context *Context) Data {
	ValidateArgs(args, []string{"Dict", "Data"}, []string{"List", "Int"})

	if list, ok := args.First().(List); ok {
		index := args.Second().(Int)
		return list.Get(int(index.Value))
	}

	if dict, ok := args.First().(Dict); ok {
		if value, defined := dict.entries[args.Second()]; defined {
			return value
		}
	}

	return Nothing{}
}

// (put dict key value) - sets the dictionary entry "key" to value
func _put(args List, context *Context) Data {
	ValidateArgs(args, []string{"Dict", "Data", "Data"}, []string{"List", "Int", "Data"})

	if dict, ok := args.First().(Dict); ok {
		key := args.Second()
		value := args.Third()
		dict.entries[key] = value
		return value
	}

	list := args.First().(List)
	index := args.Second().(Int)
	value := args.Third()
	list.Set(index.Value, value)
	return value
}

func _print(args List, context *Context) Data {
	args.Foreach(func(data Data, i int) {
		switch t := data.(type) {
		case String:
			fmt.Println(t.Value)
		default:
			fmt.Println(data.String())
		}
	})

	return Nothing{}
}

// (append list xs1 xs2 ...) - appends lists of items to the list and returns the modified list
func _append(args List, context *Context) Data {
	args.RequireArity(2)
	if list, ok := args.First().(List); ok {
		args.Foreach(func(data Data, i int) {
			if i > 0 {
				if datas, ok := data.(List); ok {
					list.PushBackList(datas.List)
				} else {
					list.PushBack(data)
				}
			}
		})
		return list
	}

	panic("First argument must be a list!")
}

// (prepend list xs1 xs2 ...) - prepends lists of items to the list and returns the modified list
func _prepend(args List, context *Context) Data {
	args.RequireArity(2)
	if list, ok := args.Get(1).(List); ok {
		args.Foreach(func(data Data, i int) {
			if i > 0 {
				if datas, ok := data.(List); ok {
					list.PushFrontList(datas.List)
				} else {
					list.PushFront(data)
				}
			}
		})
		return list
	}

	panic("First argument must be a list!")
}

// (slice list startInl endExcl) - get all items between startIncl and endExcl
// (slice list 0 endExcl) - get all items till end
// (slice list startIncl) - get all items from startIncl till end
func _slice(args List, context *Context) Data {
	ValidateArgs(args, []string{"List", "Int", "Int"}, []string{"List", "Int"})

	list := args.First().(List)

	if args.Len() == 3 {
		start := args.Second().(Int)
		end := args.Third().(Int)
		return list.Slice(start.Value, end.Value)
	}

	start := args.Second().(Int)
	return list.Slice(start.Value, list.Len())
}

// (apply f collection) -
func _apply(args List, context *Context) Data {
	ValidateArgs(args, []string{"Function", "List"}, []string{"NativeFunction", "List"})
	// TODO: Implement apply
	return Nothing{}
}

// (foreach collection f)
func _foreach(args List, context *Context) Data {
	ValidateArgs(args, []string{"List", "Function"}, []string{"List", "NativeFunction"},
		[]string{"Dict", "Function"}, []string{"Dict", "NativeFunction"})

	// List
	if list, ok := args.First().(List); ok {
		f := args.Second().(Caller)
		list.Foreach(func(data Data, i int) {
			fArgs := MakeList(data)
			f.Call(fArgs, context)
		})
	} else {
		// Dictionary
		dict := args.First().(Dict)
		for key, value := range dict.entries {
			f := args.Second().(Caller)
			fArgs := MakeList(key, value)
			f.Call(fArgs, context)
		}
	}

	return Nothing{}
}

// (map f collection)
func _map(args List, context *Context) Data {
	ValidateArgs(args, []string{"Function", "List"}, []string{"NativeFunction", "List"},
		[]string{"Function", "Dict"}, []string{"NativeFunction", "Dict"})

	fn := args.First().(Caller)
	if list, ok := args.Second().(List); ok {
		return list.Map(func(data Data, i int) Data {
			fnArgs := MakeList(data)
			return fn.Call(fnArgs, context)
		})
	} else {
		dict := args.Second().(Dict)
		results := CreateList()

		for key, value := range dict.entries {
			fnArgs := MakeList(key, value)
			results.PushBack(fn.Call(fnArgs, context))
		}

		return results
	}

	return Nothing{}
}

// (filter f list) - Returns a list of items for which f returns true
func _filter(args List, context *Context) Data {
	ValidateArgs(args, []string{"Function", "List"}, []string{"NativeFunction", "List"},
		[]string{"Function", "Dict"}, []string{"NativeFunction", "Dict"})

	fn := args.First().(Caller)

	if list, ok := args.Second().(List); ok {
		return list.Filter(func(data Data, i int) bool {
			fnArgs := MakeList(data)
			return fn.Call(fnArgs, context).(Bool).Value
		})
	} else {
		dict := args.Second().(Dict)
		results := CreateList()
		for key, value := range dict.entries {
			fnArgs := MakeList(key, value)
			if fn.Call(fnArgs, context).(Bool).Value {
				results.PushBack(value)
			}
		}
		return results
	}

	return Nothing{}
}

// (len list) - Returns number of items in list
// (len dict) - Returns number of key-value pairs in dictionary
func _len(args List, context *Context) Data {
	ValidateArgs(args, []string{"Dict"}, []string{"List"}, []string{"String"})

	switch t := args.First().(type) {
	case List:
		return Int{t.Len()}
	case Dict:
		return Int{len(t.entries)}
	case String:
		return Int{len(t.Value)}
	}

	return Nothing{}
}

// (str x) - returns the string representation of x
func _str(args List, context *Context) Data {
	args.RequireArity(1)
	if args.Len() == 1 {
		return String{args.First().String()}
	}

	str := ""
	for e := args.Front(); e != nil; e = e.Next() {
		switch t := e.Value.(type) {
		case String:
			str += t.Value
		case Data:
			str += t.String()
		}
	}

	return String{str}
}

func _first(args List, context *Context) Data {
	ValidateArgs(args, []string{"List"})

	switch t := args.First().(type) {
	case List:
		return t.Front().Value.(Data)
	}

	return Nothing{}
}

func _last(args List, context *Context) Data {
	ValidateArgs(args, []string{"List"})

	switch t := args.First().(type) {
	case List:
		return t.Back().Value.(Data)
	}

	return Nothing{}
}
