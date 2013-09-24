package main

import "strconv"
import "strings"

type Int struct {
	Value int
}

type Float struct {
	Value float64
}

type Adder interface {
	Plus(a Data) Data
}

type Subtracter interface {
	Minus(a Data) Data
}

type Multiplyer interface {
	Multiply(a Data) Data
}

type Divider interface {
	Divide(a Data) Data
}

func (i Int) Plus(a Data) Data {
	switch num := a.(type) {
	case Int:
		return Int{i.Value + num.Value}
	case Float:
		return Float{float64(i.Value) + num.Value}
	}

	panic("Addition only works with Ints and Floats")
}

func (i Int) Minus(a Data) Data {
	switch num := a.(type) {
	case Int:
		return Int{i.Value - num.Value}
	case Float:
		return Float{float64(i.Value) - num.Value}
	}

	panic("Subtraction only works with Ints and Floats")
}

func (i Int) Multiply(a Data) Data {
	switch num := a.(type) {
	case Int:
		return Int{i.Value * num.Value}
	case Float:
		return Float{float64(i.Value) * num.Value}
	case String:
		return String{strings.Repeat(num.Value, i.Value)}
	}

	panic("Multiplication only works with Ints and Floats, or Strings")
}

func (s String) Multiply(a Data) Data {
	switch t := a.(type) {
	case Int:
		return String{strings.Repeat(s.Value, t.Value)}
	}

	panic("Strings can only be multiplied with Ints")
}

func (s String) Plus(a Data) Data {
	switch t := a.(type) {
	case String:
		return String{s.Value + t.Value}
	default:
		return String{s.Value + t.String()}
	}
}

func (i Int) Divide(a Data) Data {
	switch num := a.(type) {
	case Int:
		return Int{i.Value / num.Value}
	case Float:
		return Float{float64(i.Value) / num.Value}
	}

	panic("Division only works with Ints and Floats")
}

func (f Float) Plus(a Data) Data {
	switch num := a.(type) {
	case Int:
		return Float{f.Value + float64(num.Value)}
	case Float:
		return Float{f.Value + num.Value}
	}

	panic("Addition only works with Ints and Floats")
}

func (f Float) Minus(a Data) Data {
	switch num := a.(type) {
	case Int:
		return Float{f.Value - float64(num.Value)}
	case Float:
		return Float{f.Value - num.Value}
	}

	panic("Subtraction only works with Ints and Floats")
}

func (f Float) Multiply(a Data) Data {
	switch num := a.(type) {
	case Int:
		return Float{f.Value * float64(num.Value)}
	case Float:
		return Float{f.Value * num.Value}
	}

	panic("Multiplication only works with Ints and Floats")
}

func (f Float) Divide(a Data) Data {
	switch num := a.(type) {
	case Int:
		return Float{f.Value / float64(num.Value)}
	case Float:
		return Float{f.Value / num.Value}
	}

	panic("Division only works with Ints and Floats")
}

//=============================================================================
// Native functions
//=============================================================================

func _plus(code List, context *Context) Data {
	code.RequireArity(3)

	if adder, ok := code.Second().(Adder); ok {
		return adder.Plus(code.Third())
	}

	panic("Operand not supported")
}

func _minus(code List, context *Context) Data {
	code.RequireArity(3)

	if subber, ok := code.Second().(Subtracter); ok {
		return subber.Minus(code.Third())
	}

	panic("Operand not supported")
}

func _multiply(code List, context *Context) Data {
	code.RequireArity(3)

	if m, ok := code.Second().(Multiplyer); ok {
		return m.Multiply(code.Third())
	}

	panic("Operand not supported")
}

func _divide(code List, context *Context) Data {
	code.RequireArity(3)

	if divider, ok := code.Second().(Divider); ok {
		return divider.Divide(code.Third())
	}

	panic("Operand not supported")
}

//=============================================================================
// String conversion
//=============================================================================

func (i Int) String() string {
	return strconv.FormatInt(int64(i.Value), 10)
}

func (f Float) String() string {
	return strconv.FormatFloat(f.Value, 'g', -1, 64)
}
