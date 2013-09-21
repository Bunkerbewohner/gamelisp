package main

import "container/list"
import "strconv"
import "fmt"
import "bytes"

type Data interface {
	String() string
}

type List struct {
	*list.List
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
