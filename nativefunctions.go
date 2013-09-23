package main

import "reflect"
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

func _type(code List, context *Context) Data {
	code.RequireArity(2)
	return String{reflect.TypeOf(code.Second()).String()}
}

func _def(code List, context *Context) Data {
	code.RequireArity(3)

	// the symbol referring to the defined value
	symbol, ok := code.Second().(Symbol)
	if !ok {
		panic("First argument to def must be a symbol")
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

func _list(code List, context *Context) Data {
	result := code.Filter(func(data Data, i int) bool {
		return i > 0
	})

	result.evaluated = true
	return result
}

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

func _symbol(code List, context *Context) Data {
	code.RequireArity(2)

	str, ok := code.Second().(String)
	if ok {
		return Symbol{str.Value}
	}

	panic("symbol expects string as first argument")
}

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

	return nil
}
