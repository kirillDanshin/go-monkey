package monkey

/*
#include "monkey.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// JavaScript Array
type Array struct {
	rt  *Runtime
	obj *C.JSObject
}

// See newObject()
func newArray(rt *Runtime, obj *C.JSObject) *Array {
	result := &Array{rt, obj}

	C.JS_AddObjectRoot(rt.cx, &result.obj)

	runtime.SetFinalizer(result, func(o *Array) {
		C.JS_RemoveObjectRoot(o.rt.cx, &o.obj)
	})

	return result
}

func (o *Array) ToValue() *Value {
	return newValue(o.rt, C.OBJECT_TO_JSVAL(o.obj))
}

func (o *Array) GetLength() (int, bool) {
	o.rt.lock()
	defer o.rt.unlock()

	var l C.jsuint
	if C.JS_GetArrayLength(o.rt.cx, o.obj, &l) == C.JS_TRUE {
		return int(l), true
	}
	return 0, false
}

func (o *Array) SetLength(length int) bool {
	o.rt.lock()
	defer o.rt.unlock()

	return C.JS_SetArrayLength(o.rt.cx, o.obj, C.jsuint(length)) == C.JS_TRUE
}

func (o *Array) GetElement(index int) (*Value, bool) {
	o.rt.lock()
	defer o.rt.unlock()

	r := o.rt.Void()
	if C.JS_GetElement(o.rt.cx, o.obj, C.jsint(index), &r.val) == C.JS_TRUE {
		return r, true
	}
	return r, false
}

func (o *Array) SetElement(index int, v *Value) bool {
	o.rt.lock()
	defer o.rt.unlock()

	return C.JS_SetElement(o.rt.cx, o.obj, C.jsint(index), &v.val) == C.JS_TRUE
}

func (o *Array) GetProperty(name string) (*Value, bool) {
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

func (o *Array) SetProperty(name string, v *Value) bool {
	o.rt.lock()
	defer o.rt.unlock()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.JS_SetProperty(o.rt.cx, o.obj, cname, &v.val) == C.JS_TRUE
}
