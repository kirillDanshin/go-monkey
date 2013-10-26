package monkey

/*
#include "monkey.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// JavaScript Context
type Context struct {
	rt            *Runtime
	jscx          *C.JSContext
	jsglobal      *C.JSObject
	callbacks     map[string]JsFunc
	errorReporter ErrorReporter
}

func (r *Runtime) NewContext() *Context {
	c := new(Context)

	c.rt = r
	c.callbacks = make(map[string]JsFunc)

	c.jscx = C.JS_NewContext(r.jsrt, 8192)
	if c.jscx == nil {
		return nil
	}

	C.JS_SetOptions(c.jscx, C.JSOPTION_VAROBJFIX|C.JSOPTION_JIT|C.JSOPTION_METHODJIT)
	C.JS_SetVersion(c.jscx, C.JSVERSION_LATEST)
	C.JS_SetErrorReporter(c.jscx, C.the_error_callback)

	c.jsglobal = C.JS_NewCompartmentAndGlobalObject(c.jscx, &C.global_class, nil)

	if C.JS_InitStandardClasses(c.jscx, c.jsglobal) != C.JS_TRUE {
		return nil
	}

	// User defined function use this to find callback.
	C.JS_SetContextPrivate(c.jscx, unsafe.Pointer(c))

	runtime.SetFinalizer(c, func(c *Context) {
		C.JS_DestroyContext(c.jscx)
	})

	return c
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
	c.errorReporter = reporter
}

// Evaluate JavaScript
// When you need high efficiency or run same script many times, please look at Compile() method.
func (c *Context) Eval(script string) *Value {
	c.rt.lock()
	defer c.rt.unlock()

	cscript := C.CString(script)
	defer C.free(unsafe.Pointer(cscript))

	var rval C.jsval
	if C.JS_EvaluateScript(
		c.jscx, c.jsglobal,
		cscript, C.uintN(len(script)),
		C.eval_filename, 0,
		&rval,
	) == C.JS_TRUE {
		return newValue(c, rval)
	}

	return nil
}

// Compiled Script
type Script struct {
	cx  *Context
	obj *C.JSObject
}

func (s *Script) Context() *Context {
	return s.cx
}

func (s *Script) Runtime() *Runtime {
	return s.cx.rt
}

// Execute the script
func (s *Script) Execute() *Value {
	s.cx.rt.lock()
	defer s.cx.rt.unlock()

	var rval C.jsval
	if C.JS_ExecuteScript(s.cx.jscx, s.cx.jsglobal, s.obj, &rval) == C.JS_TRUE {
		return newValue(s.cx, rval)
	}

	return nil
}

// Compile JavaScript
// When you need run a script many times, you can use this to avoid dynamic compile.
func (c *Context) Compile(code, filename string, lineno int) *Script {
	c.rt.lock()
	defer c.rt.unlock()

	ccode := C.CString(code)
	defer C.free(unsafe.Pointer(ccode))

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	var obj = C.JS_CompileScript(c.jscx, c.jsglobal, ccode, C.size_t(len(code)), cfilename, C.uintN(lineno))

	if obj != nil {
		script := &Script{c, obj}

		C.JS_AddObjectRoot(c.jscx, &script.obj)

		runtime.SetFinalizer(script, func(s *Script) {
			C.JS_RemoveObjectRoot(s.cx.jscx, &s.obj)
		})

		return script
	}

	return nil
}

type JsFunc func(context *Context, argv []*Value) *Value

//export call_go_func
func call_go_func(c unsafe.Pointer, name *C.char, argc C.uintN, vp *C.jsval) C.JSBool {
	var context = (*Context)(c)

	var argv = make([]*Value, int(argc))

	for i := 0; i < len(argv); i++ {
		argv[i] = newValue(context, C.GET_ARGV(context.jscx, vp, C.int(i)))
	}

	var result = context.callbacks[C.GoString(name)](context, argv)

	if result != nil {
		C.SET_RVAL(context.jscx, vp, result.val)
		return C.JS_TRUE
	}

	return C.JS_FALSE
}

// Define a function into runtime
// @name     The function name
// @callback The function implement
func (c *Context) DefineFunction(name string, callback JsFunc) bool {
	c.rt.lock()
	defer c.rt.unlock()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if C.JS_DefineFunction(c.jscx, c.jsglobal, cname, C.the_go_func_callback, 0, 0) == nil {
		return false
	}

	c.callbacks[name] = callback

	return true
}

func (c *Context) Runtime() *Runtime {
	return c.rt
}

// Warp null
func (c *Context) Null() *Value {
	c.rt.lock()
	defer c.rt.unlock()
	return newValue(c, C.GET_NULL())
}

// Warp void
func (c *Context) Void() *Value {
	c.rt.lock()
	defer c.rt.unlock()
	return newValue(c, C.GET_VOID())
}

// Warp integer
func (c *Context) Int(v int32) *Value {
	c.rt.lock()
	defer c.rt.unlock()
	return newValue(c, C.INT_TO_JSVAL(C.int32(v)))
}

// Warp float
func (c *Context) Number(v float64) *Value {
	c.rt.lock()
	defer c.rt.unlock()
	return newValue(c, C.DOUBLE_TO_JSVAL(C.jsdouble(v)))
}

// Warp string
func (c *Context) String(v string) *Value {
	c.rt.lock()
	defer c.rt.unlock()

	cv := C.CString(v)
	defer C.free(unsafe.Pointer(cv))

	return newValue(c, C.STRING_TO_JSVAL(C.JS_NewStringCopyN(c.jscx, cv, C.size_t(len(v)))))
}

// Warp boolean
func (c *Context) Boolean(v bool) *Value {
	c.rt.lock()
	defer c.rt.unlock()
	if v {
		return newValue(c, C.JS_TRUE)
	}
	return newValue(c, C.JS_FALSE)
}

// Create an empty array, like: []
func (c *Context) NewArray() *Array {
	c.rt.lock()
	defer c.rt.unlock()
	return newArray(c, C.JS_NewArrayObject(c.jscx, 0, nil))
}

// Create an empty object, like: {}
func (c *Context) NewObject() *Object {
	c.rt.lock()
	defer c.rt.unlock()
	return newObject(c, C.JS_NewObject(c.jscx, nil, nil, nil))
}
