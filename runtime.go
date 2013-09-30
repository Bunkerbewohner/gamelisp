package main

import "errors"
import "fmt"
import "io/ioutil"
import "strings"
import "os"

var modules map[string]*Module = make(map[string]*Module)
var moduleSearchPaths = []string{"."}

type Context struct {
	symbols map[string]Data
	parent  *Context
}

type Module struct {
	name    string
	source  string
	context *Context
}

// Gets a module by name. Loads the module beforehand if necessary
func GetModule(name string) *Module {
	module, ok := modules[name]
	if !ok {
		module = LoadModule(name)
		modules[name] = module
	}

	return module
}

func FindModuleFile(name string) string {
	for _, path := range moduleSearchPaths {
		filepath := path + "/" + name
		if _, err := os.Stat(filepath); err == nil {
			return filepath
		}
	}

	return ""
}

func LoadModule(name string) *Module {
	path := FindModuleFile(name)
	if path == "" {
		panic(fmt.Sprintf("Module %s could not be found in search path", name))
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to load module %s: %s", name, err.Error()))
	}

	text := "(do " + string(bytes) + ")"
	context := NewContext()
	_, err = EvaluateString(text, context)
	if err == nil {
		module := new(Module)
		module.name = strings.TrimSuffix(strings.ToLower(name), ".gl")
		module.context = context
		module.source = name

		return module
	} else {
		panic(err.Error())
	}
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

	context.symbols["do"] = NativeFunction{_do}
	context.symbols["foreach"] = NativeFunction{_foreach}
	context.symbols["map"] = NativeFunction{_map}
	context.symbols["filter"] = NativeFunction{_filter}
	context.symbols["apply"] = NativeFunction{_apply}

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
