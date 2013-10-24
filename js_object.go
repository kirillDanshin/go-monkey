package monkey

/*
#cgo linux  LDFLAGS: -lmozjs185
#cgo darwin LDFLAGS: -lmozjs185

#include "js/jsapi.h"

extern JSPropertyOp       the_go_getter_callback;
extern JSStrictPropertyOp the_go_setter_callback;
extern JSNative           the_go_obj_func_callback;

extern void _JS_SET_RVAL(JSContext *cx, jsval* vp, jsval v);
extern jsval JS_GET_ARGV(JSContext *cx, jsval* vp, int n);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// JavaScript Object
type Object struct {
	rt      *Runtime
	obj     *C.JSObject
	funcs   map[string]JsFunc
	getters map[string]JsPropertyGetter
	setters map[string]JsPropertySetter
}

// Add the JSObject to the garbage collector's root set.
// See: https://developer.mozilla.org/en-US/docs/Mozilla/Projects/SpiderMonkey/JSAPI_reference/JS_AddRoot
func newObject(rt *Runtime, obj *C.JSObject) *Object {
	result := &Object{rt, obj, nil, nil, nil}

	C.JS_AddObjectRoot(rt.cx, &result.obj)

	runtime.SetFinalizer(result, func(o *Object) {
		C.JS_RemoveObjectRoot(o.rt.cx, &o.obj)
	})

	// User defined property and function object use this to find callback.
	C.JS_SetPrivate(rt.cx, result.obj, unsafe.Pointer(result))

	return result
}

func (o *Object) ToValue() *Value {
	return newValue(o.rt, C.OBJECT_TO_JSVAL(o.obj))
}

func (o *Object) GetProperty(name string) (*Value, bool) {
	o.rt.lock()
	defer o.rt.unlock()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	r := o.rt.Void()
	if C.JS_GetProperty(o.rt.cx, o.obj, cname, &r.val) == C.JS_TRUE {
		return r, true
	}
	return r, false
}

func (o *Object) SetProperty(name string, v *Value) bool {
	o.rt.lock()
	defer o.rt.unlock()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.JS_SetProperty(o.rt.cx, o.obj, cname, &v.val) == C.JS_TRUE
}

type JsPropertyAttrs uint

// Property Attributes
const (
	JSPROP_ENUMERATE = C.JSPROP_ENUMERATE // The property is visible to JavaScript for…in and for each…in loops.
	JSPROP_READONLY  = C.JSPROP_READONLY  // The property's value cannot be set.
	JSPROP_PERMANENT = C.JSPROP_PERMANENT // The property cannot be deleted.
)

type JsPropertyGetter func(o *Object) (*Value, bool)
type JsPropertySetter func(o *Object, v *Value) bool

//export call_go_getter
func call_go_getter(obj unsafe.Pointer, name *C.char, val *C.jsval) C.JSBool {
	o := (*Object)(obj)
	if o.getters != nil {
		if v, ok := o.getters[C.GoString(name)](o); ok {
			*val = v.val
			return C.JS_TRUE
		}
	}
	return C.JS_FALSE
}

//export call_go_setter
func call_go_setter(obj unsafe.Pointer, name *C.char, val *C.jsval) C.JSBool {
	o := (*Object)(obj)
	if o.setters != nil {
		if o.setters[C.GoString(name)](o, newValue(o.rt, *val)) {
			return C.JS_TRUE
		}
	}
	return C.JS_FALSE
}

func (o *Object) DefineProperty(name string, value *Value, getter JsPropertyGetter, setter JsPropertySetter, attrs JsPropertyAttrs) bool {
	o.rt.lock()
	defer o.rt.unlock()

	if C.JS_IsArrayObject(o.rt.cx, o.obj) == C.JS_TRUE {
		panic("Could't define property on array.")
	}

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var r C.JSBool

	if getter != nil && setter != nil {
		r = C.JS_DefineProperty(o.rt.cx, o.obj, cname, value.val, C.the_go_getter_callback, C.the_go_setter_callback, C.uintN(uint(attrs))|C.JSPROP_SHARED)
	} else if getter != nil && setter == nil {
		r = C.JS_DefineProperty(o.rt.cx, o.obj, cname, value.val, C.the_go_getter_callback, nil, C.uintN(uint(attrs)))
	} else if getter == nil && setter != nil {
		r = C.JS_DefineProperty(o.rt.cx, o.obj, cname, value.val, nil, C.the_go_setter_callback, C.uintN(uint(attrs)))
	} else {
		panic("The getter and setter both nil")
	}

	if r == C.JS_TRUE {
		if getter != nil {
			if o.getters == nil {
				o.getters = make(map[string]JsPropertyGetter)
			}
			o.getters[name] = getter
		}

		if setter != nil {
			if o.setters == nil {
				o.setters = make(map[string]JsPropertySetter)
			}
			o.setters[name] = setter
		}

		return true
	}

	return false
}

//export call_go_obj_func
func call_go_obj_func(op unsafe.Pointer, name *C.char, argc C.uintN, vp *C.jsval) C.JSBool {
	var o = (*Object)(op)

	var argv = make([]*Value, int(argc))

	for i := 0; i < len(argv); i++ {
		argv[i] = newValue(o.rt, C.JS_GET_ARGV(o.rt.cx, vp, C.int(i)))
	}

	var result, ok = o.funcs[C.GoString(name)](argv)

	if ok {
		C._JS_SET_RVAL(o.rt.cx, vp, result.val)
		return C.JS_TRUE
	}

	return C.JS_FALSE
}

// Define a function into object
// @name     The function name
// @callback The function implement
func (o *Object) DefineFunction(name string, callback JsFunc) bool {
	o.rt.lock()
	defer o.rt.unlock()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if C.JS_DefineFunction(o.rt.cx, o.obj, cname, C.the_go_obj_func_callback, 0, 0) == nil {
		return false
	}

	if o.funcs == nil {
		o.funcs = make(map[string]JsFunc)
	}

	o.funcs[name] = callback

	return true
}
