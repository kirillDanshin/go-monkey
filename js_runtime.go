package monkey

/*
#cgo linux  LDFLAGS: -lmozjs185
#cgo darwin LDFLAGS: -lmozjs185

#include "monkey.h"
*/
import "C"
import (
	"errors"
	"github.com/realint/monkey/goid"
	"runtime"
	"sync"
	"unsafe"
)

// JavaScript Runtime
type Runtime struct {
	rt            *C.JSRuntime
	cx            *C.JSContext
	global        *C.JSObject
	callbacks     map[string]JsFunc
	errorReporter ErrorReporter
	lockBy        int32
	lockLevel     int
	mutex         sync.Mutex
}

// Initializes the JavaScript runtime.
// @maxbytes Maximum number of allocated bytes after which garbage collection is run.
func NewRuntime(maxbytes uint32) (*Runtime, error) {
	r := new(Runtime)
	r.callbacks = make(map[string]JsFunc)

	r.rt = C.JS_NewRuntime(C.uint32(maxbytes))
	if r.rt == nil {
		return nil, errors.New("Could't create JSRuntime")
	}

	r.cx = C.JS_NewContext(r.rt, 8192)
	if r.cx == nil {
		return nil, errors.New("Could't create JSContext")
	}

	C.JS_SetOptions(r.cx, C.JSOPTION_VAROBJFIX|C.JSOPTION_JIT|C.JSOPTION_METHODJIT)
	C.JS_SetVersion(r.cx, C.JSVERSION_LATEST)
	C.JS_SetErrorReporter(r.cx, C.the_error_callback)

	r.global = C.JS_NewCompartmentAndGlobalObject(r.cx, &C.global_class, nil)

	if C.JS_InitStandardClasses(r.cx, r.global) != C.JS_TRUE {
		return nil, errors.New("Could't init global class")
	}

	// User defined function use this to find callback.
	C.JS_SetRuntimePrivate(r.rt, unsafe.Pointer(r))

	runtime.SetFinalizer(r, func(r *Runtime) {
		C.JS_DestroyContext(r.cx)
		C.JS_DestroyRuntime(r.rt)
	})

	return r, nil
}

// Because we can't prevent Go to execute a JavaScript that maybe will execute another JavaScript by invoke Go function.
// Like this: runtime.Eval("eval('1 + 1')")
// So I designed this lock mechanism to let runtime can lock by same goroutine many times.
func (r *Runtime) lock() {
	id := goid.Get()
	if r.lockBy != id {
		r.mutex.Lock()
		r.lockBy = id
	} else {
		r.lockLevel += 1
	}
}

func (r *Runtime) unlock() {
	r.lockLevel -= 1
	if r.lockLevel < 0 {
		r.lockLevel = 0
		r.lockBy = -1
		r.mutex.Unlock()
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
	Message    string
	FileName   string
	LineBuf    string
	LineNum    int
	ErrorNum   int
	TokenIndex int
	Flags      ErrorReportFlags
}

//export call_error_func
func call_error_func(r unsafe.Pointer, message *C.char, report *C.JSErrorReport) {
	if (*Runtime)(r).errorReporter != nil {
		(*Runtime)(r).errorReporter(&ErrorReport{
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
func (r *Runtime) SetErrorReporter(reporter ErrorReporter) {
	r.errorReporter = reporter
}

// Evaluate JavaScript
// When you need high efficiency or run same script many times, please look at Compile() method.
func (r *Runtime) Eval(script string) *Value {
	r.lock()
	defer r.unlock()

	cscript := C.CString(script)
	defer C.free(unsafe.Pointer(cscript))

	var rval C.jsval
	if C.JS_EvaluateScript(r.cx, r.global, cscript, C.uintN(len(script)), C.eval_filename, 0, &rval) == C.JS_TRUE {
		return newValue(r, rval)
	}

	return nil
}

// Compiled Script
type Script struct {
	rt  *Runtime
	obj *C.JSObject
}

func (s *Script) Runtime() *Runtime {
	return s.rt
}

// Execute the script
func (s *Script) Execute() *Value {
	s.rt.lock()
	defer s.rt.unlock()

	var rval C.jsval
	if C.JS_ExecuteScript(s.rt.cx, s.rt.global, s.obj, &rval) == C.JS_TRUE {
		return newValue(s.rt, rval)
	}

	return nil
}

// Compile JavaScript
// When you need run a script many times, you can use this to avoid dynamic compile.
func (r *Runtime) Compile(code, filename string, lineno int) *Script {
	r.lock()
	defer r.unlock()

	ccode := C.CString(code)
	defer C.free(unsafe.Pointer(ccode))

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	var obj = C.JS_CompileScript(r.cx, r.global, ccode, C.size_t(len(code)), cfilename, C.uintN(lineno))

	if obj != nil {
		script := &Script{r, obj}

		C.JS_AddObjectRoot(r.cx, &script.obj)

		runtime.SetFinalizer(script, func(s *Script) {
			C.JS_RemoveObjectRoot(s.rt.cx, &s.obj)
		})

		return script
	}

	return nil
}

type JsFunc func(runtime *Runtime, argv []*Value) *Value

//export call_go_func
func call_go_func(r unsafe.Pointer, name *C.char, argc C.uintN, vp *C.jsval) C.JSBool {
	var runtime = (*Runtime)(r)

	var argv = make([]*Value, int(argc))

	for i := 0; i < len(argv); i++ {
		argv[i] = newValue(runtime, C.GET_ARGV(runtime.cx, vp, C.int(i)))
	}

	var result = runtime.callbacks[C.GoString(name)](runtime, argv)

	if result != nil {
		C.SET_RVAL(runtime.cx, vp, result.val)
		return C.JS_TRUE
	}

	return C.JS_FALSE
}

// Define a function into runtime
// @name     The function name
// @callback The function implement
func (r *Runtime) DefineFunction(name string, callback JsFunc) bool {
	r.lock()
	defer r.unlock()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if C.JS_DefineFunction(r.cx, r.global, cname, C.the_go_func_callback, 0, 0) == nil {
		return false
	}

	r.callbacks[name] = callback

	return true
}

// Warp null
func (r *Runtime) Null() *Value {
	r.lock()
	defer r.unlock()
	return newValue(r, C.GET_NULL())
}

// Warp void
func (r *Runtime) Void() *Value {
	r.lock()
	defer r.unlock()
	return newValue(r, C.GET_VOID())
}

// Warp integer
func (r *Runtime) Int(v int32) *Value {
	r.lock()
	defer r.unlock()
	return newValue(r, C.INT_TO_JSVAL(C.int32(v)))
}

// Warp float
func (r *Runtime) Number(v float64) *Value {
	r.lock()
	defer r.unlock()
	return newValue(r, C.DOUBLE_TO_JSVAL(C.jsdouble(v)))
}

// Warp string
func (r *Runtime) String(v string) *Value {
	r.lock()
	defer r.unlock()

	cv := C.CString(v)
	defer C.free(unsafe.Pointer(cv))

	return newValue(r, C.STRING_TO_JSVAL(C.JS_NewStringCopyN(r.cx, cv, C.size_t(len(v)))))
}

// Warp boolean
func (r *Runtime) Boolean(v bool) *Value {
	r.lock()
	defer r.unlock()
	if v {
		return newValue(r, C.JS_TRUE)
	}
	return newValue(r, C.JS_FALSE)
}

// Create an empty array, like: []
func (r *Runtime) NewArray() *Array {
	r.lock()
	defer r.unlock()
	return newArray(r, C.JS_NewArrayObject(r.cx, 0, nil))
}

// Create an empty object, like: {}
func (r *Runtime) NewObject() *Object {
	r.lock()
	defer r.unlock()
	return newObject(r, C.JS_NewObject(r.cx, nil, nil, nil))
}
