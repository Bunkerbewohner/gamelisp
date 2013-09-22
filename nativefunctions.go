package main

import "reflect"
import "fmt"

//
// code List is the expression list handed to the evaluator,
// the first element being the function name, all following
// elements are arguments to the denoted function
//

func _type(code List, context *Context) Data {
	code.RequireArity(2)
	return String{reflect.TypeOf(code.Second()).String()}
}

func def(code List, context *Context) Data {
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
