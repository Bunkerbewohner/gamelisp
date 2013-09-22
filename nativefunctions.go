package main

import "reflect"
import "fmt"

func _type(code List, context *Context) Data {
	return String{reflect.TypeOf(code.Second()).String()}
}

func def(code List, context *Context) Data {
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
