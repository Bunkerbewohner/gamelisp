package main

import "container/list"
import "fmt"
import "bytes"

type Data interface {
	String() string
	Equals(other Data) bool
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
	for _, arg := range args {
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

func (ls List) GetElement(n int) *list.Element {
	// positive indices = offset from front
	if n >= 0 {
		i := 0
		for e := ls.Front(); e != nil; e = e.Next() {
			if i == n {
				return e
			}
			i++
		}
	}

	// negative indices = offset from back
	i := -1
	for e := ls.Back(); e != nil; e = e.Prev() {
		if i == n {
			return e
		}
		i--
	}

	return nil
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

func (x DataType) GetType() DataType {
	return DataTypeType
}

func (x Function) GetType() DataType {
	return FunctionType
}

//=============================================================================
// Equality
//=============================================================================

func (x Int) Equals(other Data) bool {
	switch t := other.(type) {
	case Int:
		return t.Value == x.Value
	case Float:
		return t.Value == float64(x.Value)
	}

	return false
}

func (x Float) Equals(other Data) bool {
	switch t := other.(type) {
	case Float:
		return t.Value == x.Value
	case Int:
		return float64(t.Value) == x.Value
	}

	return false
}

func (x Bool) Equals(other Data) bool {
	switch t := other.(type) {
	case Bool:
		return t.Value == x.Value
	}

	return false
}

func (x String) Equals(other Data) bool {
	switch t := other.(type) {
	case String:
		return t.Value == x.Value
	}

	return false
}

// two lists are defined as equal when their contents are equal
func (x List) Equals(other Data) bool {
	switch t := other.(type) {
	case List:
		if t.Len() != x.Len() {
			return false
		}

		a := x.Front()
		b := t.Front()

		for a != nil && b != nil {
			dataA := a.Value.(Data)
			dataB := b.Value.(Data)
			if !dataA.Equals(dataB) {
				return false
			}
			a = a.Next()
			b = b.Next()
		}
	}

	return false
}

func (x Nothing) Equals(other Data) bool {
	switch other.(type) {
	case Nothing:
		return true
	}
	return false
}

func (x Symbol) Equals(other Data) bool {
	switch t := other.(type) {
	case Symbol:
		return t.Value == x.Value
	}
	return false
}

func (x Keyword) Equals(other Data) bool {
	switch t := other.(type) {
	case Keyword:
		return t.Value == x.Value
	}
	return false
}

func (x DataType) Equals(other Data) bool {
	switch t := other.(type) {
	case DataType:
		return t.TypeName == x.TypeName
	}
	return false
}

func (x Dict) Equals(other Data) bool {
	otherDict, ok := other.(Dict)
	if !ok {
		return false
	}

	for key, value := range x.entries {
		otherValue, ok := otherDict.entries[key]
		if !ok {
			return false
		}
		if !otherValue.Equals(value) {
			return false
		}
	}

	return true
}

func (x NativeFunction) Equals(other Data) bool {
	panic("native functions cannot be compared")
}

func (x NativeFunctionB) Equals(other Data) bool {
	panic("native functions cannot be compared")
}

func (x Function) Equals(other Data) bool {
	switch t := other.(type) {
	case Function:
		return t.String() == x.String()
	}
	return false
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
var FunctionType = DataType{"Function"}
var NativeFunctionBType = DataType{"NativeFunctionB"}
var NativeFunctionType = DataType{"NativeFunction"}
var NothingType = DataType{"Nothing"}
var StringType = DataType{"String"}
var SymbolType = DataType{"Symbol"}
var DataTypeType = DataType{"DataType"}
var ContextType = DataType{"Context"}
