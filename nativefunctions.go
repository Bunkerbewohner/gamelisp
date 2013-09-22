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

	symbol := code.Second()
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
	return code.Filter(func(data Data, i int) bool {
		return i > 0
	})
}

func _print(code List, context *Context) Data {
	code.RequireArity(2)

	code.Foreach(func(data Data, i int) {
		if i > 0 {
			switch t := data.(type) {
			case String:
				fmt.Print(t.Value)
			default:
				fmt.Print(data.String())
			}
		}
	})

	return nil
}
