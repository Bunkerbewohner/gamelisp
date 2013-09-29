package main

import "errors"
import "fmt"

type Context struct {
	symbols map[string]Data
	parent  *Context
}

func NewContext() *Context {
	return &Context{
		make(map[string]Data),
		nil,
	}
}

func (c *Context) Define(symbol Symbol, value Data) {
	c.symbols[symbol.Value] = value
}

func (c *Context) IsDefined(symbol Symbol) bool {
	_, defined := c.symbols[symbol.Value]
	if !defined && c.parent != nil {
		return c.parent.IsDefined(symbol)
	}
	return defined
}

func (c *Context) LookUp(symbol Symbol) Data {
	val, defined := c.symbols[symbol.Value]
	if defined {
		return val
	} else if c.parent != nil {
		return c.parent.LookUp(symbol)
	} else {
		return nil
	}
}

func EvaluateString(code string, context *Context) (Data, error) {
	ast, err := Parse(code)
	if err != nil {
		return nil, err
	}

	result, err := Evaluate(ast, context)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func Evaluate(code Data, context *Context) (Data, error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%v in %v\n", e, code)
		}
	}()

	switch t := code.(type) {
	case List:
		if t.evaluated {
			// if the list was already evaluated just return its contents as is
			return code, nil
		} else if t.Len() == 0 {
			return nil, errors.New("invalid function invocation")
		}

		// first expression must be a symbol
		symbol, ok := t.Front().Value.(Symbol)
		if ok {
			// look up the value for that symbol
			fn := context.LookUp(symbol)
			if fn != nil {
				// check if we can call it as a function
				fn, ok := fn.(Caller)
				if ok {
					t.Remove(t.Front())
					return fn.Call(t, context), nil
				} else {
					return nil, errors.New(fmt.Sprintf("%s is not a function", t.Get(0)))
				}
			} else {
				return nil, errors.New(fmt.Sprintf("%s is not defined", t.Get(0)))
			}
		}

		function, ok := t.Front().Value.(Function)
		if ok {
			t.Remove(t.Front()) // remove function name from list to get only arguments
			return function.Call(t, context), nil
		}

		return nil, errors.New(fmt.Sprintf("%s is neither a symbol nor a function and cannot be called as such", t.Get(0)))
	case Keyword:
		return t, nil
	case Symbol:
		// look up the symbol and returns its value
		result := context.LookUp(t)
		if result != nil {
			return context.LookUp(t), nil
		} else {
			return nil, errors.New(fmt.Sprintf("%s is not defined", t.Value))
		}
	}

	return code, nil
}

func CreateMainContext() *Context {
	context := NewContext()
	context.symbols["Int"] = IntType
	context.symbols["Float"] = FloatType
	context.symbols["Bool"] = BoolType
	context.symbols["String"] = StringType
	context.symbols["Symbol"] = SymbolType
	context.symbols["Keyword"] = KeywordType
	context.symbols["List"] = ListType
	context.symbols["Dict"] = DictType
	context.symbols["NativeFunction"] = NativeFunctionType
	context.symbols["NativeFunctionB"] = NativeFunctionBType

	context.symbols["Nothing"] = Nothing{}
	context.symbols["true"] = Bool{true}
	context.symbols["false"] = Bool{false}

	context.symbols["def"] = NativeFunctionB{_def}
	context.symbols["type"] = NativeFunction{_type}
	context.symbols["str"] = NativeFunction{_str}
	context.symbols["fn"] = NativeFunctionB{_fn}
	context.symbols["defn"] = NativeFunctionB{_defn}
	context.symbols["defn.."] = NativeFunctionB{_extend_function}

	context.symbols["symbol"] = NativeFunction{_symbol}
	context.symbols["keyword"] = NativeFunction{_keyword}
	context.symbols["list"] = NativeFunction{_list}
	context.symbols["dict"] = NativeFunction{_dict}

	context.symbols["print"] = NativeFunction{_print}

	context.symbols["foreach"] = NativeFunction{_foreach}
	context.symbols["map"] = NativeFunction{_map}
	context.symbols["filter"] = NativeFunction{_filter}

	context.symbols["get"] = NativeFunction{_get}
	context.symbols["put"] = NativeFunction{_put}
	context.symbols["slice"] = NativeFunction{_slice}
	context.symbols["len"] = NativeFunction{_len}
	context.symbols["append"] = NativeFunction{_append}
	context.symbols["prepend"] = NativeFunction{_prepend}
	context.symbols["first"] = NativeFunction{_first}
	context.symbols["last"] = NativeFunction{_last}

	context.symbols["+"] = NativeFunction{_plus}
	context.symbols["-"] = NativeFunction{_minus}
	context.symbols["*"] = NativeFunction{_multiply}
	context.symbols["/"] = NativeFunction{_divide}

	return context
}

var MainContext = CreateMainContext()
