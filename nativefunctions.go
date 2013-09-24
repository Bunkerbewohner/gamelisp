package main

//
// This file contains elemental functions defined natively in Go,
// which all user-defined code and extra libraries are based on
//

import "reflect"
import "strings"
import "fmt"

//
// code List is the expression list handed to the evaluator,
// the first element being the function name, all following
// elements are arguments to the denoted function
//

// returns a function that evaluates all arguments of a expression list,
// i.e. every item but the first, using the given context
func __evalArgs(context *Context) func(data Data, i int) Data {
	return func(data Data, i int) Data {
		if i > 0 {
			result, err := Evaluate(data, context)
			if err != nil {
				panic(err.Error())
			}
			return result
		}

		return data
	}
}

// (type x) - Returns the type of x as a string
func _type(code List, context *Context) Data {
	code.RequireArity(2)
	if typer, ok := code.Second().(DataTyper); ok {
		return typer.GetType()
	}

	name := reflect.TypeOf(code.Second()).String()
	return String{strings.Replace(name, "main.", "", 1)}
}

// (def symbol value) - Defines a new symbol and assigns the value
func _def(code List, context *Context) Data {
	code.RequireArity(3)

	// the symbol referring to the defined value
	symbol, ok := code.Second().(Symbol)
	if !ok {
		panic("First argument to def must be a symbol")
	}

	// check if that symbol is already defined in the current context
	if context.IsDefined(symbol) && code.First().String() != "def!" {
		panic("Symbol is already defined. Use def! if you want to overwrite in case of an existing symbol")
	}

	// get the value that shall be associated to the symbol
	value := code.Third()
	value, err := Evaluate(value, context)

	if err == nil {
		context.symbols[symbol.String()] = value
	} else {
		fmt.Printf(err.Error())
		return nil
	}

	return value
}

// returns a list of items
func _list(code List, context *Context) Data {
	result := code.Filter(func(data Data, i int) bool {
		return i > 0
	})

	result.evaluated = true
	return result
}

// creates a dictionary. expects even number of arguments
// (dict :key1 value1 :key2 value2 ...)
func _dict(code List, context *Context) Data {
	dict := CreateDict()

	if (code.Len()-1)%2 == 1 {
		panic("Dictionary requires an even number of arguments")
	}

	for e := code.Front().Next(); e != nil; e = e.Next() {
		key, _ := e.Value.(Data)
		value, _ := e.Next().Value.(Data)
		e = e.Next()

		dict.entries[key] = value
	}

	return dict
}

// (symbol name) - return a symbol with given name
func _symbol(code List, context *Context) Data {
	code.RequireArity(2)

	str, ok := code.Second().(String)
	if ok {
		return Symbol{str.Value}
	}

	panic("symbol expects string as first argument")
}

// (keyword name) - return a keyword with given name (prepended with a colon if not supplied)
func _keyword(code List, context *Context) Data {
	code.RequireArity(2)

	str, ok := code.Second().(String)
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
func _get(code List, context *Context) Data {
	code.RequireArity(3)

	list, isList := code.Second().(List)
	dict, isDict := code.Second().(Dict)
	if !isList && !isDict {
		panic("get expects first argument to be a dictionary or list")
	}

	if isList {
		index, ok := code.Third().(Int)
		if !ok {
			panic("get expects second argument to be an integer if used on lists")
		}

		return list.Get(int(index.Value))
	}

	if isDict {
		key := code.Third()
		value, ok := dict.entries[key]
		if !ok {
			return Nothing{}
		}
		return value
	}

	return Nothing{}
}

// (put dict key value) - sets the dictionary entry "key" to value
func _put(code List, context *Context) Data {
	code.RequireArity(3)
	dict, isDict := code.Second().(Dict)
	if !isDict {
		list, isList := code.Second().(List)
		if isList {
			index, isInt := code.Get(2).(Int)
			if isInt {
				return list.Set(index.Value, code.Get(3))
			} else {
				panic("Put requires index as second argument")
			}
		} else {
			panic("First argument must be a list or dictionary")
		}
		panic("First argument must be a list or dictionary")
	}

	key := code.Get(2)
	value := code.Get(3)

	dict.entries[key] = value
	return value
}

func _print(code List, context *Context) Data {
	code.RequireArity(2)

	code.Foreach(func(data Data, i int) {
		if i > 0 {
			switch t := data.(type) {
			case String:
				fmt.Println(t.Value)
			default:
				fmt.Println(data.String())
			}
		}
	})

	return nil
}

// (append list xs1 xs2 ...) - appends lists of items to the list and returns the modified list
func _append(code List, context *Context) Data {
	code.RequireArity(3)
	list, isList := code.Get(1).(List)
	if isList {
		code.SliceFrom(2).Foreach(func(data Data, i int) {
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
func _prepend(code List, context *Context) Data {
	code.RequireArity(3)
	list, isList := code.Get(1).(List)
	if isList {
		code.SliceFrom(2).Foreach(func(data Data, i int) {
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
func _slice(code List, context *Context) Data {
	code.RequireArity(3)

	list, isList := code.Second().(List)
	if !isList {
		panic("First argument expected to be list")
	}

	if code.Len() == 4 {
		start, startInt := code.Get(2).(Int)
		end, endInt := code.Get(3).(Int)
		if !startInt || !endInt {
			panic("Indices must be integers!")
		}

		return list.Slice(start.Value, end.Value)
	} else if code.Len() == 3 {
		start, startInt := code.Get(2).(Int)
		if !startInt {
			panic("Index must be integer!")
		}

		return list.Slice(start.Value, list.Len())
	}

	panic("Invalid invocation")
}

func _foreach(code List, context *Context) Data {
	code.RequireArity(3)

	items, gotItems := code.Second().(List)
	fn, gotFn := code.Third().(Caller)

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

func _map(code List, context *Context) Data {
	code.RequireArity(3)

	fn, gotFn := code.Second().(Caller)
	items, gotItems := code.Third().(List)

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
func _filter(code List, context *Context) Data {
	code.RequireArity(3)
	fn, gotFn := code.Second().(Caller)
	items, gotItems := code.Third().(List)

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
func _len(code List, context *Context) Data {
	code.RequireArity(2)

	list, okList := code.Second().(List)
	dict, okDict := code.Second().(Dict)
	str, okStr := code.Second().(String)

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
func _str(code List, context *Context) Data {
	code.RequireArity(2)
	return String{code.Second().String()}
}
