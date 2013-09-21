package main

import "reflect"

func _type(code List, context *Context) Data {
	return String{reflect.TypeOf(code.Second()).String()}
}

func def(code List, context *Context) Data {
	symbol := code.Second()
	value := code.Third()
	value = Evaluate(value, context)

	context.symbols[symbol.String()] = Evaluate(value, context)

	return value
}
