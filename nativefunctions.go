package main

import "reflect"

func _type(code List, context *Context) Data {
	return String{reflect.TypeOf(code.Second()).String()}
}
