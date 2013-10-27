package main

import "errors"
import "fmt"
import "io/ioutil"
import "strings"
import "os"
import "path/filepath"
import "github.com/howeyc/fsnotify"
import "mk/Apollo/events"

var MainContext *Context

// modules by name
var modules map[string]*Module = make(map[string]*Module)

// modules by path
var modulesByPath map[string]*Module = make(map[string]*Module)
var moduleSearchPaths = []string{"modules"}
var watcher *fsnotify.Watcher
var watcherDone chan bool

type Context struct {
	symbols map[string]Data
	parent  *Context
	usages  []Usage
}

type Usage struct {
	context *Context
	prefix  string
}

type Module struct {
	name    string
	source  string
	context *Context
}

func (module *Module) Refresh() {
	module.Reload()

	// reimport this module into all usage contexts
	for _, usage := range module.context.usages {
		usage.context.Reimport(module.context, usage.prefix)
	}
}

func (c *Context) String() string {
	return "Context"
}

func (c *Context) Equals(other Data) bool {
	return false
}

func (c *Context) GetType() DataType {
	return ContextType
}

// Gets a module by name. Loads the module beforehand if necessary
func GetModule(name string, env *Context) *Module {
	module, ok := modules[name]
	if !ok {
		module = LoadModule(name, env)
		modules[name] = module
	}

	return module
}

func FindModuleFile(name string) string {
	for _, modulePath := range moduleSearchPaths {
		path := modulePath + "/" + strings.Replace(name, ".", "/", -1) + ".glisp"
		absPath, err := filepath.Abs(path)
		if err != nil {
			panic(err.Error())
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	return ""
}

func (module *Module) Reload() {
	bytes, err := ioutil.ReadFile(module.source)
	if err != nil {
		panic(fmt.Sprintf("Failed to reload module %s: %s", module.name, err.Error()))
	}

	text := "(do " + string(bytes) + ""
	_, err = EvaluateString(text, module.context)
	if err != nil {
		panic(err.Error())
	}
}

func LoadModule(name string, env *Context) *Module {
	path := FindModuleFile(name)
	if path == "" {
		panic(fmt.Sprintf("Module %s could not be found in search path", name))
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to load module %s: %s", name, err.Error()))
	}

	text := "(do " + string(bytes) + ""
	context := NewContext()
	context.parent = env

	_, err = EvaluateString(text, context)
	if err == nil {
		module := new(Module)
		module.name = name
		module.context = context
		module.source = path

		err := watcher.Watch(filepath.Dir(path))
		if err != nil {
			fmt.Print(err.Error())
		}

		modulesByPath[path] = module

		return module
	} else {
		panic(err.Error())
	}
}

func NewContext() *Context {
	return &Context{
		make(map[string]Data),
		nil,
		make([]Usage, 0),
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

func (c *Context) Reimport(other *Context, prefix string) {
	for key, value := range other.symbols {
		c.symbols[prefix+key] = value
	}
}

func (c *Context) Import(other *Context, prefix string) {
	for key, value := range other.symbols {
		c.symbols[prefix+key] = value
	}

	other.usages = append(other.usages, Usage{c, prefix})
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
		// copy the list because we're going to mutate it
		t = t.SliceFrom(0)

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

func initWatchdog() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err.Error())
	}

	watcher = w
	//watcherDone = make(chan bool)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev == nil {
					return
				}

				if ev.IsModify() && strings.HasSuffix(ev.Name, ".glisp") {
					if module, ok := modulesByPath[ev.Name]; ok {
						module.Refresh()
					}
				}
			case err := <-watcher.Error:
				if err == nil {
					return
				}
			}
		}
	}()
}

func InitRuntime() {
	initWatchdog()
	MainContext = CreateMainContext()
	ECS_init()
}

func shutdownWatchdog() {
	//<-watcherDone
	watcher.Close()
}

func ShutdownRuntime() {
	eventBus := MainContext.symbols["$events"].(NativeObject).Value.(*events.EventBus)
	eventBus.Shutdown()

	ECS_shutdown()
	shutdownWatchdog()
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

	context.symbols["do"] = NativeFunction{_do}
	context.symbols["def"] = NativeFunctionB{_def}
	context.symbols["type"] = NativeFunction{_type}
	context.symbols["str"] = NativeFunction{_str}
	context.symbols["fn"] = NativeFunctionB{_fn}
	context.symbols["defn"] = NativeFunctionB{_defn}
	context.symbols["defn|"] = NativeFunctionB{_extend_function}
	context.symbols["lambda"] = NativeFunctionB{_lambda}

	context.symbols["symbol"] = NativeFunction{_symbol}
	context.symbols["keyword"] = NativeFunction{_keyword}
	context.symbols["list"] = NativeFunction{_list}
	context.symbols["dict"] = NativeFunction{_dict}

	context.symbols["print"] = NativeFunction{_print}

	context.symbols["do"] = NativeFunction{_do}
	context.symbols["let"] = NativeFunctionB{_let}
	context.symbols["foreach"] = NativeFunction{_foreach}
	context.symbols["map"] = NativeFunction{_map}
	context.symbols["filter"] = NativeFunction{_filter}
	context.symbols["apply"] = NativeFunction{_apply}

	// control flow
	context.symbols["if"] = NativeFunctionB{_if}
	context.symbols["="] = NativeFunction{_equals}

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

	context.symbols["compare"] = NativeFunction{_compare}
	context.symbols["<"] = NativeFunction{_lesser_than}
	context.symbols[">"] = NativeFunction{_greater_than}
	context.symbols["<="] = NativeFunction{_lesser_than_or_equal}
	context.symbols[">="] = NativeFunction{_greater_than_or_equal}
	context.symbols["=="] = NativeFunction{_equals}

	context.symbols["range"] = NativeFunction{_range}

	context.symbols["import"] = NativeFunctionB{_import}
	context.symbols["$core"] = context
	context.symbols["code"] = NativeFunction{_code}

	context.symbols["entity"] = NativeFunction{_entity}

	// event system
	eventBus := new(events.EventBus)
	eventBus.Init()
	context.symbols["$events"] = NativeObject{eventBus}

	// import aux. functions defined in gamelisp itself
	coreModule := GetModule("$core", context)
	context.Import(coreModule.context, "")

	return context
}
