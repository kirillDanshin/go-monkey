package monkey

/*
#include "monkey.h"
*/
import "C"
import (
	"runtime"
)

// JavaScript Array
type Array struct {
	cx  *Context
	obj *C.JSObject
}

// See newObject()
func newArray(cx *Context, obj *C.JSObject) *Array {
	result := &Array{cx, obj}

	C.JS_AddObjectRoot(cx.jscx, &result.obj)

	runtime.SetFinalizer(result, func(a *Array) {
		cx.rt.aryDisposeChan <- a
	})

	return result
}

func (a *Array) ToValue() *Value {
	return newValue(a.cx, C.OBJECT_TO_JSVAL(a.obj))
}

func (a *Array) GetLength() int {
	var result int = -1
	a.cx.rt.Use(func() {
		var l C.jsuint
		if C.JS_GetArrayLength(a.cx.jscx, a.obj, &l) == C.JS_TRUE {
			result = int(l)
		}
	})
	return result
}

func (a *Array) SetLength(length int) bool {
	var result bool
	a.cx.rt.Use(func() {
		result = C.JS_SetArrayLength(a.cx.jscx, a.obj, C.jsuint(length)) == C.JS_TRUE
	})
	return result
}

func (a *Array) GetElement(index int) *Value {
	var result *Value
	a.cx.rt.Use(func() {
		var rval C.jsval
		if C.JS_GetElement(a.cx.jscx, a.obj, C.jsint(index), &rval) == C.JS_TRUE {
			result = newValue(a.cx, rval)
		}
	})
	return result
}

func (a *Array) SetElement(index int, v *Value) bool {
	var result bool
	a.cx.rt.Use(func() {
		result = C.JS_SetElement(a.cx.jscx, a.obj, C.jsint(index), &v.val) == C.JS_TRUE
	})
	return result
}

/*
Utilities
*/

func (a *Array) GetInt(index int) (int32, bool) {
	if v := a.GetElement(index); v != nil {
		return v.ToInt()
	}
	return 0, false
}

func (a *Array) SetInt(index int, v int32) bool {
	return a.SetElement(index, a.cx.Int(v))
}

func (a *Array) GetNumber(index int) (float64, bool) {
	if v := a.GetElement(index); v != nil {
		return v.ToNumber()
	}
	return 0, false
}

func (a *Array) SetNumber(index int, v float64) bool {
	return a.SetElement(index, a.cx.Number(v))
}

func (a *Array) GetBoolean(index int) (bool, bool) {
	if v := a.GetElement(index); v != nil {
		return v.ToBoolean()
	}
	return false, false
}

func (a *Array) SetBoolean(index int, v bool) bool {
	return a.SetElement(index, a.cx.Boolean(v))
}

func (a *Array) GetString(index int) (string, bool) {
	if v := a.GetElement(index); v != nil {
		return v.ToString(), true
	}
	return "", false
}

func (a *Array) SetString(index int, v string) bool {
	return a.SetElement(index, a.cx.String(v))
}

func (a *Array) GetObject(index int) *Object {
	if v := a.GetElement(index); v != nil {
		return v.ToObject()
	}
	return nil
}

func (a *Array) SetObject(index int, o *Object) bool {
	return a.SetElement(index, o.ToValue())
}

func (a *Array) GetArray(index int) *Array {
	if v := a.GetElement(index); v != nil {
		return v.ToArray()
	}
	return nil
}

func (a *Array) SetArray(index int, o *Array) bool {
	return a.SetElement(index, o.ToValue())
}
