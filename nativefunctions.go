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
func CheckSignature(args List, expected ...[]string) {
	valid := false

	for _, checks := range expected {
		partvalid := true
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
			for _, t := range checks {
				msg += t + " "
			}
		}

		msg += "\nProvided instead: "
		for e := args.Front(); e != nil; e = e.Next() {
			msg += reflect.TypeOf(e.Value).Name() + " "
		}

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
	CheckSignature(args, []string{"Symbol", "Data"})

	// the symbol referring to the defined value
	symbol, ok := args.First().(Symbol)
	if !ok {
		panic("First argument to def must be a symbol")
	}

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
	if args.Len() == 3 {
		return CreateFunction(args, context)
	} else if args.Len() == 2 {
		// generate a random function name
		name := "anonymous-function"

		// insert name as first argument
		args.InsertBefore(Symbol{name}, args.Front())

		return CreateFunction(args, context)
	}

	panic("fn expects either 2 or 3 arguments")
}

// (defn name args* stmts*)
func _defn(args List, context *Context) Data {
	name, ok := args.First().(Symbol)
	if !ok {
		panic("First argument must be a symbol")
	}

	fn := CreateFunction(args, context)

	// Save the function into the context
	context.Define(name, fn)

	return fn
}

// returns a list of items
func _list(args List, context *Context) Data {
	result := args.Filter(func(data Data, i int) bool {
		return i > 0
	})

	result.evaluated = true
	return result
}

// creates a dictionary. expects even number of arguments
// (dict :key1 value1 :key2 value2 ...)
func _dict(args List, context *Context) Data {
	dict := CreateDict()

	if (args.Len()-1)%2 == 1 {
		panic("Dictionary requires an even number of arguments")
	}

	for e := args.Front().Next(); e != nil; e = e.Next() {
		key, _ := e.Value.(Data)
		value, _ := e.Next().Value.(Data)
		e = e.Next()

		dict.entries[key] = value
	}

	return dict
}

// (symbol name) - return a symbol with given name
func _symbol(args List, context *Context) Data {
	args.RequireArity(2)

	str, ok := args.Second().(String)
	if ok {
		return Symbol{str.Value}
	}

	panic("symbol expects string as first argument")
}

// (keyword name) - return a keyword with given name (prepended with a colon if not supplied)
func _keyword(args List, context *Context) Data {
	args.RequireArity(2)

	str, ok := args.Second().(String)
	if ok {
		if str.Value[0] != ':' {
			str.Value = ":" + str.Value
		}

		return Keyword{str.Value}
	}

	panic("keyword expects string as first argument")
}

// (get dict key)	- gets an entry from a dictionary
// (get list index)	- gets an entry from a list
// returns Nothing if entry doesn't exist
func _get(args List, context *Context) Data {
	args.RequireArity(3)

	list, isList := args.Second().(List)
	dict, isDict := args.Second().(Dict)
	if !isList && !isDict {
		panic("get expects first argument to be a dictionary or list")
	}

	if isList {
		index, ok := args.Third().(Int)
		if !ok {
			panic("get expects second argument to be an integer if used on lists")
		}

		return list.Get(int(index.Value))
	}

	if isDict {
		key := args.Third()
		value, ok := dict.entries[key]
		if !ok {
			return Nothing{}
		}
		return value
	}

	return Nothing{}
}

// (put dict key value) - sets the dictionary entry "key" to value
func _put(args List, context *Context) Data {
	args.RequireArity(3)
	dict, isDict := args.Second().(Dict)
	if !isDict {
		list, isList := args.Second().(List)
		if isList {
			index, isInt := args.Get(2).(Int)
			if isInt {
				return list.Set(index.Value, args.Get(3))
			} else {
				panic("Put requires index as second argument")
			}
		} else {
			panic("First argument must be a list or dictionary")
		}
		panic("First argument must be a list or dictionary")
	}

	key := args.Get(2)
	value := args.Get(3)

	dict.entries[key] = value
	return value
}

func _print(args List, context *Context) Data {
	args.RequireArity(2)

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
	args.RequireArity(3)
	list, isList := args.Get(1).(List)
	if isList {
		args.SliceFrom(2).Foreach(func(data Data, i int) {
			if datas, ok := data.(List); ok {
				list.PushBackList(datas.List)
			} else {
				list.PushBack(data)
			}
		})
		return list
	}

	panic("First argument must be a list!")
}

// (prepend list xs1 xs2 ...) - prepends lists of items to the list and returns the modified list
func _prepend(args List, context *Context) Data {
	args.RequireArity(3)
	list, isList := args.Get(1).(List)
	if isList {
		args.SliceFrom(2).Foreach(func(data Data, i int) {
			if datas, ok := data.(List); ok {
				list.PushFrontList(datas.List)
			} else {
				list.PushFront(data)
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
	args.RequireArity(3)

	list, isList := args.Second().(List)
	if !isList {
		panic("First argument expected to be list")
	}

	if args.Len() == 4 {
		start, startInt := args.Get(2).(Int)
		end, endInt := args.Get(3).(Int)
		if !startInt || !endInt {
			panic("Indices must be integers!")
		}

		return list.Slice(start.Value, end.Value)
	} else if args.Len() == 3 {
		start, startInt := args.Get(2).(Int)
		if !startInt {
			panic("Index must be integer!")
		}

		return list.Slice(start.Value, list.Len())
	}

	panic("Invalid invocation")
}

func _foreach(args List, context *Context) Data {
	args.RequireArity(3)

	items, gotItems := args.Second().(List)
	fn, gotFn := args.Third().(Caller)

	if gotItems && gotFn {
		items.Foreach(func(data Data, i int) {
			args := CreateList()
			args.PushBack(fn)
			args.PushBack(data)

			fn.Call(args, context)
		})
	}

	return nil
}

func _map(args List, context *Context) Data {
	args.RequireArity(3)

	fn, gotFn := args.Second().(Caller)
	items, gotItems := args.Third().(List)

	if gotFn && gotItems {
		return items.Map(func(data Data, i int) Data {
			args := CreateList()
			args.PushBack(fn)
			args.PushBack(data)
			return fn.Call(args, context)
		})
	}

	if !gotFn {
		panic("First argument expected to be a function")
	}

	if !gotItems {
		panic("Second argument expected to be a list")
	}

	return Nothing{}
}

// (filter f list) - Returns a list of items for which f returns true
func _filter(args List, context *Context) Data {
	args.RequireArity(3)
	fn, gotFn := args.Second().(Caller)
	items, gotItems := args.Third().(List)

	if gotFn && gotItems {
		return items.Filter(func(data Data, i int) bool {
			args := CreateList()
			args.PushBack(fn)
			args.PushBack(data)
			result := fn.Call(args, context)
			value, isBool := result.(Bool)
			if isBool {
				return value.Value
			}

			panic("Filter function does not return Bool")
		})
	}

	if !gotFn {
		panic("First arguments expected to be a function")
	}

	if !gotItems {
		panic("Second argument expected to be a list")
	}

	return Nothing{}
}

// (len list) - Returns number of items in list
// (len dict) - Returns number of key-value pairs in dictionary
func _len(args List, context *Context) Data {
	args.RequireArity(2)

	list, okList := args.Second().(List)
	dict, okDict := args.Second().(Dict)
	str, okStr := args.Second().(String)

	if okList {
		return Int{list.Len()}
	} else if okDict {
		return Int{len(dict.entries)}
	} else if okStr {
		return Int{len(str.Value)}
	}

	panic("First arguments must be a list, dictionary or string")
}

// (str x) - returns the string representation of x
func _str(args List, context *Context) Data {
	args.RequireArity(2)
	return String{args.Second().String()}
}

func _first(args List, context *Context) Data {
	args.RequireArity(2)
	list, ok := args.Get(1).(List)
	if list.Len() == 0 {
		panic("Empty list has no first element")
	}
	if ok {
		return list.Front().Value.(Data)
	} else {
		panic("argument must be a list")
	}
}

func _last(args List, context *Context) Data {
	args.RequireArity(2)
	list, ok := args.Get(1).(List)
	if list.Len() == 0 {
		panic("Empty list has no last element")
	}
	if ok {
		return list.Back().Value.(Data)
	} else {
		panic("argument must be a list")
	}
}
