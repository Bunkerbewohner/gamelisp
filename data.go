package main

import "container/list"
import "strconv"
import "fmt"
import "bytes"

type Data interface {
	String() string
}

type Caller interface {
	Call(code List, context *Context) Data
}

type NativeFunction struct {
	// Native functions receive a list of arguments, the first being the name under
	// which the function itself was called
	Function func(List, *Context) Data
}

type List struct {
	*list.List
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

type Bool struct {
	Value bool
}

type Int struct {
	Value int64
}

type Float struct {
	Value float64
}

func (s String) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

func (s Symbol) String() string {
	return s.Value
}

func (l List) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("(")

	for e := l.Front(); e != nil; e = e.Next() {
		if buffer.Len() > 1 {
			buffer.WriteString(" ")
		}

		if data, ok := e.Value.(Data); ok {
			buffer.WriteString(data.String())
		}
	}

	buffer.WriteString(")")
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
	} else {
		return "false"
	}
}

func (i Int) String() string {
	return strconv.FormatInt(i.Value, 10)
}

func (f Float) String() string {
	return strconv.FormatFloat(f.Value, 'g', -1, 64)
}

func CreateList() List {
	return List{
		List: list.New(),
	}
}

func CreateDict() Dict {
	return Dict{
		entries: make(map[Data]Data),
	}
}

func (fn NativeFunction) Call(code List, context *Context) Data {
	return fn.Function(code, context)
}

func (fn NativeFunction) String() string {
	return "native function"
}

func (ls List) Get(n int) Data {
	i := 0
	for e := ls.Front(); e != nil; e = e.Next() {
		if i == n {
			if data, ok := e.Value.(Data); ok {
				return data
			}
		}
		i++
	}

	return nil
}

func (ls List) Second() Data {
	return ls.Get(1)
}

func (ls List) Third() Data {
	return ls.Get(2)
}
