package main

import "strconv"
import "strings"
import "math"
import "fmt"

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

type Comparer interface {
	// comparing two items with one another. Returns the result
	// of the comparison and a bool indicating whether the items
	// are comparable at all.
	//  * !comparible => comparison = 0
	//  * this < a => comparison < 0
	//	* this > a => comparison > 0
	//  * this == a => comparison = 0
	Compare(a Data) (comparison int, comparible bool)
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

func _plus(args List, context *Context) Data {
	args.RequireArity(2)

	sum, ok := args.First().(Adder)
	if !ok {
		panic("Operand not supported")
	}

	for e := args.Front().Next(); e != nil; e = e.Next() {
		sum = sum.Plus(e.Value.(Data)).(Adder)
	}

	return sum.(Data)
}

func _minus(args List, context *Context) Data {
	args.RequireArity(2)

	sum, ok := args.First().(Subtracter)
	if !ok {
		panic("Operand not supported")
	}

	for e := args.Front().Next(); e != nil; e = e.Next() {
		sum = sum.Minus(e.Value.(Data)).(Subtracter)
	}

	return sum.(Data)
}

func _multiply(args List, context *Context) Data {
	args.RequireArity(2)

	sum, ok := args.First().(Multiplyer)
	if !ok {
		panic("Operand not supported")
	}

	for e := args.Front().Next(); e != nil; e = e.Next() {
		sum = sum.Multiply(e.Value.(Data)).(Multiplyer)
	}

	return sum.(Data)
}

func _divide(args List, context *Context) Data {
	args.RequireArity(2)

	sum, ok := args.First().(Divider)
	if !ok {
		panic("Operand not supported")
	}

	for e := args.Front().Next(); e != nil; e = e.Next() {
		sum = sum.Divide(e.Value.(Data)).(Divider)
	}

	return sum.(Data)
}

func _compare(args List, context *Context) Data {
	args.RequireArity(2)

	a, ok := args.First().(Comparer)
	if !ok {
		panic("Left operand is not comparable")
	}

	comp, ok := a.Compare(args.Second())
	if !ok {
		panic(fmt.Sprintf("Right operand cannot be compared to %s", args.First().GetType().String()))
	}

	return Int{comp}
}

func _lesser_than(args List, context *Context) Data {
	args.RequireArity(2)

	operand, ok := args.First().(Comparer)
	if !ok {
		panic("Left operand is not comparable")
	}

	for e := args.Front().Next(); e != nil; e = e.Next() {
		comparison, comparable := operand.Compare(e.Value.(Data))
		if !comparable || comparison >= 0 {
			return Bool{false}
		}
		operand = e.Value.(Comparer)
	}

	return Bool{true}
}

func _greater_than(args List, context *Context) Data {
	args.RequireArity(2)

	operand, ok := args.First().(Comparer)
	if !ok {
		panic("Left operand is not comparable")
	}

	for e := args.Front().Next(); e != nil; e = e.Next() {
		comparison, comparable := operand.Compare(e.Value.(Data))
		if !comparable || comparison <= 0 {
			return Bool{false}
		}
		operand = e.Value.(Comparer)
	}

	return Bool{true}
}

func _lesser_than_or_equal(args List, context *Context) Data {
	args.RequireArity(2)

	operand, ok := args.First().(Comparer)
	if !ok {
		panic("Left operand is not comparable")
	}

	for e := args.Front().Next(); e != nil; e = e.Next() {
		comparison, comparable := operand.Compare(e.Value.(Data))
		if !comparable || comparison > 0 {
			return Bool{false}
		}
		operand = e.Value.(Comparer)
	}

	return Bool{true}
}

func _greater_than_or_equal(args List, context *Context) Data {
	args.RequireArity(2)

	operand, ok := args.First().(Comparer)
	if !ok {
		panic("Left operand is not comparable")
	}

	for e := args.Front().Next(); e != nil; e = e.Next() {
		comparison, comparable := operand.Compare(e.Value.(Data))
		if !comparable || comparison < 0 {
			return Bool{false}
		}
		operand = e.Value.(Comparer)
	}

	return Bool{true}
}

//=============================================================================
// Comparison
//=============================================================================

func (i Int) Compare(a Data) (int, bool) {
	switch t := a.(type) {
	case Int:
		return i.Value - t.Value, true
	case Float:
		diff := float64(i.Value) - t.Value
		if diff < 0 {
			return int(math.Min(diff, -1)), true
		} else if diff > 0 {
			return int(math.Max(diff, 1)), true
		} else {
			return 0, true
		}
	}

	return 0, false
}

func (f Float) Compare(a Data) (int, bool) {
	switch t := a.(type) {
	case Int:
		diff := f.Value - float64(t.Value)
		if diff < 0 {
			return int(math.Min(diff, -1)), true
		} else if diff > 0 {
			return int(math.Max(diff, 1)), true
		} else {
			return 0, true
		}
	case Float:
		diff := f.Value - t.Value
		if diff < 0 {
			return int(math.Min(diff, -1)), true
		} else if diff > 0 {
			return int(math.Max(diff, 1)), true
		} else {
			return 0, true
		}
	}
	return 0, false
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
