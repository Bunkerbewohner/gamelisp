package main

import "container/list"
import "fmt"
import "bytes"

type Data interface {
	String() string
}

type DataTyper interface {
	GetType() DataType
}

type DataType struct {
	TypeName string
}

type List struct {
	*list.List
	evaluated bool
}

type Dict struct {
	entries map[Data]Data
}

type String struct {
	Value string
}

type Symbol struct {
	Value string
}

type Keyword struct {
	Value string
}

type Bool struct {
	Value bool
}

type Nothing struct {
}

func (t DataType) String() string {
	return t.TypeName
}

func (s String) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

func (s Symbol) String() string {
	return s.Value
}

func (k Keyword) String() string {
	return k.Value
}

func (n Nothing) String() string {
	return "Nothing"
}

func (l List) String() string {
	var buffer bytes.Buffer
	if l.evaluated {
		buffer.WriteString("[")
	} else {
		buffer.WriteString("(")
	}

	for e := l.Front(); e != nil; e = e.Next() {
		if buffer.Len() > 1 {
			buffer.WriteString(" ")
		}

		if data, ok := e.Value.(Data); ok {
			buffer.WriteString(data.String())
		}
	}

	if l.evaluated {
		buffer.WriteString("]")
	} else {
		buffer.WriteString(")")
	}
	return buffer.String()
}

func (d Dict) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	i := 0

	for key, value := range d.entries {
		if i > 0 {
			buffer.WriteString(" ")
		}

		buffer.WriteString(key.String())
		buffer.WriteString(" ")
		buffer.WriteString(value.String())

		i++
	}

	buffer.WriteString("}")
	return buffer.String()
}

func (b Bool) String() string {
	if b.Value {
		return "true"
	}

	return "false"
}

func CreateList() List {
	return List{
		List: list.New(),
	}
}

func MakeList(args ...Data) List {
	list := CreateList()
	for arg := range args {
		list.PushBack(arg)
	}
	return list
}

func CreateDict() Dict {
	return Dict{
		entries: make(map[Data]Data),
	}
}

func (ls List) Plus(a Data) Data {
	switch t := a.(type) {
	case List:
		copy := CreateList()
		copy.PushBackList(ls.List)
		copy.PushBackList(t.List)
		return copy
	default:
		copy := CreateList()
		copy.PushBackList(ls.List)
		copy.PushBack(t)
		return copy
	}

	return Nothing{}
}

func (ls List) Set(n int, value Data) Data {
	// positive indices = offset from front
	if n >= 0 {
		i := 0
		for e := ls.Front(); e != nil; e = e.Next() {
			if i == n {
				e.Value = value
				return value
			}
			i++
		}
	}

	// negative indices = offset from back
	i := -1
	for e := ls.Back(); e != nil; e = e.Prev() {
		if i == n {
			e.Value = value
			return value
		}
		i--
	}

	return Nothing{}
}

func (ls List) Get(n int) Data {
	// positive indices = offset from front
	if n >= 0 {
		i := 0
		for e := ls.Front(); e != nil; e = e.Next() {
			if i == n {
				if data, ok := e.Value.(Data); ok {
					return data
				}
			}
			i++
		}
	}

	// negative indices = offset from back
	i := -1
	for e := ls.Back(); e != nil; e = e.Prev() {
		if i == n {
			if data, ok := e.Value.(Data); ok {
				return data
			}
		}
		i--
	}

	return Nothing{}
}

func (ls List) Last() Data {
	last := ls.Back()
	if last == nil {
		return Nothing{}
	} else {
		if data, ok := last.Value.(Data); ok {
			return data
		} else {
			panic("List contains invalid data type")
		}
	}
}

func (ls List) First() Data {
	return ls.Get(0)
}

func (ls List) Second() Data {
	return ls.Get(1)
}

func (ls List) Third() Data {
	return ls.Get(2)
}

func (ls List) RequireArity(n int) {
	if ls.Len() < n {
		panic(fmt.Sprintf("%d elements expected, only %d provided", n, ls.Len()))
	}
}

func (ls List) Foreach(f func(a Data, i int)) {
	i := 0
	for e := ls.Front(); e != nil; e = e.Next() {
		switch t := e.Value.(type) {
		case Data:
			f(t, i)
		}
		i++
	}
}

func (ls List) Slice(startIncl int, endExcl int) List {
	if startIncl < 0 {
		startIncl += ls.Len()
	}
	if endExcl < 0 {
		endExcl += ls.Len()
	}

	return ls.Filter(func(a Data, i int) bool {
		return i >= startIncl && i < endExcl
	})
}

func (ls List) SliceFrom(startIncl int) List {
	if startIncl < 0 {
		startIncl += ls.Len()
	}

	return ls.Filter(func(a Data, i int) bool {
		return i >= startIncl
	})
}

func (ls List) Filter(f func(a Data, i int) bool) List {
	list := CreateList()
	i := 0

	for e := ls.Front(); e != nil; e = e.Next() {
		switch t := e.Value.(type) {
		case Data:
			if f(t, i) {
				list.PushBack(t)
			}
		}
		i++
	}

	return list
}

func (ls List) Map(f func(a Data, i int) Data) List {
	list := CreateList()
	i := 0

	for e := ls.Front(); e != nil; e = e.Next() {
		switch t := e.Value.(type) {
		case Data:
			list.PushBack(f(t, i))
		}
		i++
	}

	return list
}

func (x String) GetType() DataType {
	return StringType
}

func (x Int) GetType() DataType {
	return IntType
}

func (x Float) GetType() DataType {
	return FloatType
}

func (x Bool) GetType() DataType {
	return BoolType
}

func (x Symbol) GetType() DataType {
	return SymbolType
}

func (x Keyword) GetType() DataType {
	return KeywordType
}

func (x List) GetType() DataType {
	return ListType
}

func (x Dict) GetType() DataType {
	return DictType
}

func (x Nothing) GetType() DataType {
	return NothingType
}

func (x NativeFunction) GetType() DataType {
	return NativeFunctionType
}

func (x NativeFunctionB) GetType() DataType {
	return NativeFunctionBType
}

//=============================================================================
// Global Variables
//=============================================================================

// built-in intrinsic types
var BoolType = DataType{"Bool"}
var DictType = DataType{"Dict"}
var FloatType = DataType{"Float"}
var IntType = DataType{"Int"}
var KeywordType = DataType{"Keyword"}
var ListType = DataType{"List"}
var NativeFunctionBType = DataType{"NativeFunctionB"}
var NativeFunctionType = DataType{"NativeFunction"}
var NothingType = DataType{"Nothing"}
var StringType = DataType{"String"}
var SymbolType = DataType{"Symbol"}
