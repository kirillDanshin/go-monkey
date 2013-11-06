package monkey

/*
#include "monkey.h"
*/
import "C"
import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// JavaScript Context
type Context struct {
	rt            *Runtime
	jscx          *C.JSContext
	jsglobal      *C.JSObject
	funcs         map[string]JsFunc
	errorReporter ErrorReporter
	disposed      int64
}

func (r *Runtime) NewContext() *Context {
	var result *Context

	r.Use(func() {
		c := new(Context)
		c.rt = r

		c.jscx = C.JS_NewContext(r.jsrt, 8192)
		if c.jscx == nil {
			return
		}

		C.JS_SetOptions(c.jscx, C.JSOPTION_VAROBJFIX|C.JSOPTION_JIT|C.JSOPTION_METHODJIT)
		C.JS_SetVersion(c.jscx, C.JSVERSION_LATEST)

		c.jsglobal = C.JS_NewCompartmentAndGlobalObject(c.jscx, &C.global_class, nil)

		if C.JS_InitStandardClasses(c.jscx, c.jsglobal) != C.JS_TRUE {
			return
		}

		// User defined function use this to find callback.
		C.JS_SetContextPrivate(c.jscx, unsafe.Pointer(c))

		runtime.SetFinalizer(c, func(c *Context) {
			c.Dispose()
		})

		result = c
	})

	return result
}

// Free by manual
func (c *Context) Dispose() {
	if atomic.CompareAndSwapInt64(&c.disposed, 0, 1) {
		c.rt.ctxDisposeChan <- c
	}
}

type ErrorReporter func(report *ErrorReport)

type ErrorReportFlags uint

const (
	JSREPORT_WARNING   = ErrorReportFlags(C.JSREPORT_WARNING)
	JSREPORT_EXCEPTION = ErrorReportFlags(C.JSREPORT_EXCEPTION)
	JSREPORT_STRICT    = ErrorReportFlags(C.JSREPORT_STRICT)
)

type ErrorReport struct {
	Context    *Context
	Message    string
	FileName   string
	LineBuf    string
	LineNum    int
	ErrorNum   int
	TokenIndex int
	Flags      ErrorReportFlags
}

//export call_error_func
func call_error_func(c unsafe.Pointer, message *C.char, report *C.JSErrorReport) {
	cx := (*Context)(c)
	if cx.errorReporter != nil {
		cx.errorReporter(&ErrorReport{
			Context:    cx,
			Message:    C.GoString(message),
			FileName:   C.GoString(report.filename),
			LineNum:    int(report.lineno),
			ErrorNum:   int(report.errorNumber),
			LineBuf:    C.GoString(report.linebuf),
			TokenIndex: int(uintptr(unsafe.Pointer(report.tokenptr)) - uintptr(unsafe.Pointer(report.linebuf))),
		})
	}
}

// Set a error reporter
func (c *Context) SetErrorReporter(reporter ErrorReporter) {
	if c.errorReporter == nil {
		C.JS_SetErrorReporter(c.jscx, C.the_error_callback)
	}
	c.errorReporter = reporter
}

// Evaluate JavaScript
// When you need high efficiency or run same script many times, please look at Compile() method.
func (c *Context) Eval(script string) *Value {
	var result *Value

	c.rt.Use(func() {
		cscript := C.CString(script)
		defer C.free(unsafe.Pointer(cscript))

		var rval C.jsval
		if C.JS_EvaluateScript(c.jscx, c.jsglobal, cscript, C.uintN(len(script)), C.eval_filename, 0, &rval) == C.JS_TRUE {
			result = newValue(c, rval)
		}
	})

	return result
}

// Compiled Script
type Script struct {
	cx       *Context
	obj      *C.JSObject
	disposed int64
}

// Free by manual
func (s *Script) Dispose() {
	if atomic.CompareAndSwapInt64(&s.disposed, 0, 1) {
		s.cx.rt.sptDisposeChan <- s
	}
}

func (s *Script) Context() *Context {
	return s.cx
}

func (s *Script) Runtime() *Runtime {
	return s.cx.rt
}

// Execute the script
func (s *Script) Execute() *Value {
	var result *Value

	s.cx.rt.Use(func() {
		var rval C.jsval
		if C.JS_ExecuteScript(s.cx.jscx, s.cx.jsglobal, s.obj, &rval) == C.JS_TRUE {
			result = newValue(s.cx, rval)
		}
	})

	return result
}

// Execute the script
func (s *Script) ExecuteIn(cx *Context) *Value {
	var result *Value

	cx.rt.Use(func() {
		var rval C.jsval
		if C.JS_ExecuteScript(cx.jscx, cx.jsglobal, s.obj, &rval) == C.JS_TRUE {
			result = newValue(cx, rval)
		}
	})

	return result
}

// Compile JavaScript
// When you need run a script many times, you can use this to avoid dynamic compile.
func (c *Context) Compile(code, filename string, lineno int) *Script {
	var result *Script

	c.rt.Use(func() {
		ccode := C.CString(code)
		defer C.free(unsafe.Pointer(ccode))

		cfilename := C.CString(filename)
		defer C.free(unsafe.Pointer(cfilename))

		var obj = C.JS_CompileScript(c.jscx, c.jsglobal, ccode, C.size_t(len(code)), cfilename, C.uintN(lineno))

		if obj != nil {
			script := &Script{c, obj, 0}

			C.JS_AddObjectRoot(c.jscx, &script.obj)

			runtime.SetFinalizer(script, func(s *Script) {
				s.Dispose()
			})

			result = script
		}
	})

	return result
}

// Go defined JS function callback info
type Func struct {
	context *Context
	name    string
	args    []*Value
	result  *Value
}

func (f *Func) Context() *Context {
	return f.context
}

func (f *Func) Name() string {
	return f.name
}

func (f *Func) Argc() int {
	return len(f.args)
}

func (f *Func) Argv(n int) *Value {
	return f.args[n]
}

func (f *Func) Return(v *Value) {
	f.result = v
}

// Go defined JS function callback
type JsFunc func(f *Func)

//export call_go_func
func call_go_func(c unsafe.Pointer, name *C.char, argc C.uintN, vp *C.jsval) C.JSBool {
	var context = (*Context)(c)

	var args = make([]*Value, int(argc))

	for i := 0; i < len(args); i++ {
		args[i] = newValue(context, C.GET_ARGV(context.jscx, vp, C.int(i)))
	}

	var gname = C.GoString(name)
	var f = Func{
		context: context,
		name:    gname,
		args:    args,
	}

	context.funcs[gname](&f)

	if f.result != nil {
		C.SET_RVAL(context.jscx, vp, f.result.val)
		return C.JS_TRUE
	}

	return C.JS_FALSE
}

// Define a function into runtime
// @name     The function name
// @callback The function implement
func (c *Context) DefineFunction(name string, callback JsFunc) bool {
	var result bool

	c.rt.Use(func() {
		cname := C.CString(name)
		defer C.free(unsafe.Pointer(cname))

		if C.JS_DefineFunction(c.jscx, c.jsglobal, cname, C.the_go_func_callback, 0, 0) == nil {
			return
		}

		if c.funcs == nil {
			c.funcs = make(map[string]JsFunc)
		}

		c.funcs[name] = callback

		result = true
	})

	return result
}

func (c *Context) Runtime() *Runtime {
	return c.rt
}

// Warp null
func (c *Context) Null() *Value {
	var result *Value
	c.rt.Use(func() {
		result = newValue(c, C.GET_NULL())
	})
	return result
}

// Warp void
func (c *Context) Void() *Value {
	var result *Value
	c.rt.Use(func() {
		result = newValue(c, C.GET_VOID())
	})
	return result
}

// Warp integer
func (c *Context) Int(v int32) *Value {
	var result *Value
	c.rt.Use(func() {
		result = newValue(c, C.INT_TO_JSVAL(C.int32(v)))
	})
	return result
}

// Warp float
func (c *Context) Number(v float64) *Value {
	var result *Value
	c.rt.Use(func() {
		result = newValue(c, C.DOUBLE_TO_JSVAL(C.jsdouble(v)))
	})
	return result
}

// Warp string
func (c *Context) String(v string) *Value {
	var result *Value
	c.rt.Use(func() {
		cv := C.CString(v)
		defer C.free(unsafe.Pointer(cv))

		result = newValue(c, C.STRING_TO_JSVAL(C.JS_NewStringCopyN(c.jscx, cv, C.size_t(len(v)))))
	})
	return result
}

// Warp boolean
func (c *Context) Boolean(v bool) *Value {
	var result *Value
	c.rt.Use(func() {
		if v {
			result = newValue(c, C.JS_TRUE)
		} else {
			result = newValue(c, C.JS_FALSE)
		}
	})
	return result
}

// Create an empty array, like: []
func (c *Context) NewArray() *Array {
	var result *Array
	c.rt.Use(func() {
		result = newArray(c, C.JS_NewArrayObject(c.jscx, 0, nil))
	})
	return result
}

// Create an empty object, like: {}
func (c *Context) NewObject(gval interface{}) *Object {
	var result *Object
	c.rt.Use(func() {
		result = newObject(c, C.JS_NewObject(c.jscx, nil, nil, nil), gval)
	})
	return result
}
