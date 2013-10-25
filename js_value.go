package monkey

/*
#include "monkey.h"
*/
import "C"
import (
	"reflect"
	"runtime"
	"unsafe"
)

// JavaScript Value
type Value struct {
	rt  *Runtime
	val C.jsval
}

func newValue(rt *Runtime, val C.jsval) *Value {
	result := &Value{rt, val}

	C.JS_AddValueRoot(rt.cx, &result.val)

	runtime.SetFinalizer(result, func(v *Value) {
		C.JS_RemoveValueRoot(v.rt.cx, &v.val)
	})

	return result
}

func (v *Value) Runtime() *Runtime {
	return v.rt
}

func (v *Value) String() string {
	return v.ToString()
}

func (v *Value) TypeName() string {
	v.rt.lock()
	defer v.rt.unlock()

	return C.GoString(C.JS_GetTypeName(v.rt.cx, C.JS_TypeOfValue(v.rt.cx, v.val)))
}

func (v *Value) IsNull() bool {
	return C.JSVAL_IS_NULL(v.val) == C.JS_TRUE
}

func (v *Value) IsVoid() bool {
	return C.JSVAL_IS_VOID(v.val) == C.JS_TRUE
}

func (v *Value) IsInt() bool {
	return C.JSVAL_IS_INT(v.val) == C.JS_TRUE
}

func (v *Value) IsNumber() bool {
	return C.JSVAL_IS_NUMBER(v.val) == C.JS_TRUE
}

func (v *Value) IsBoolean() bool {
	return C.JSVAL_IS_BOOLEAN(v.val) == C.JS_TRUE
}

func (v *Value) IsString() bool {
	return C.JSVAL_IS_STRING(v.val) == C.JS_TRUE
}

func (v *Value) IsObject() bool {
	return C.JSVAL_IS_OBJECT(v.val) == C.JS_TRUE
}

func (v *Value) IsArray() bool {
	v.rt.lock()
	defer v.rt.unlock()

	return v.IsObject() && C.JS_IsArrayObject(
		v.rt.cx, C.JSVAL_TO_OBJECT(v.val),
	) == C.JS_TRUE
}

func (v *Value) IsFunction() bool {
	v.rt.lock()
	defer v.rt.unlock()

	return v.IsObject() && C.JS_ObjectIsFunction(
		v.rt.cx, C.JSVAL_TO_OBJECT(v.val),
	) == C.JS_TRUE
}

// Convert a value to Int.
func (v *Value) ToInt() (int32, bool) {
	v.rt.lock()
	defer v.rt.unlock()

	var r C.int32
	if C.JS_ValueToInt32(v.rt.cx, v.val, &r) == C.JS_TRUE {
		return int32(r), true
	}
	return 0, false
}

// Convert a value to Number.
func (v *Value) ToNumber() (float64, bool) {
	v.rt.lock()
	defer v.rt.unlock()

	var r C.jsdouble
	if C.JS_ValueToNumber(v.rt.cx, v.val, &r) == C.JS_TRUE {
		return float64(r), true
	}
	return 0, false
}

// Convert a value to Boolean.
func (v *Value) ToBoolean() (bool, bool) {
	v.rt.lock()
	defer v.rt.unlock()

	var r C.JSBool
	if C.JS_ValueToBoolean(v.rt.cx, v.val, &r) == C.JS_TRUE {
		if r == C.JS_TRUE {
			return true, true
		}
		return false, true
	}
	return false, false
}

// Convert a value to String.
func (v *Value) ToString() string {
	v.rt.lock()
	defer v.rt.unlock()

	cstring := C.JS_EncodeString(v.rt.cx, C.JS_ValueToString(v.rt.cx, v.val))
	gostring := C.GoString(cstring)
	C.JS_free(v.rt.cx, unsafe.Pointer(cstring))

	return gostring
}

// Convert a value to Object.
func (v *Value) ToObject() *Object {
	v.rt.lock()
	defer v.rt.unlock()

	var obj *C.JSObject
	if C.JS_ValueToObject(v.rt.cx, v.val, &obj) == C.JS_TRUE {
		return newObject(v.rt, obj)
	}

	return nil
}

// Convert a value to Array.
func (v *Value) ToArray() *Array {
	v.rt.lock()
	defer v.rt.unlock()

	var obj *C.JSObject
	if C.JS_ValueToObject(v.rt.cx, v.val, &obj) == C.JS_TRUE {
		if C.JS_IsArrayObject(v.rt.cx, obj) == C.JS_TRUE {
			return newArray(v.rt, obj)
		}
	}

	return nil
}

// Call a function value
func (v *Value) Call(argv []*Value) *Value {
	v.rt.lock()
	defer v.rt.unlock()

	argv2 := make([]C.jsval, len(argv))
	for i := 0; i < len(argv); i++ {
		argv2[i] = argv[i].val
	}
	argv3 := unsafe.Pointer(&argv2)
	argv4 := (*reflect.SliceHeader)(argv3).Data
	argv5 := (*C.jsval)(unsafe.Pointer(argv4))

	var rval C.jsval
	if C.JS_CallFunctionValue(v.rt.cx, nil, v.val, C.uintN(len(argv)), argv5, &rval) == C.JS_TRUE {
		return newValue(v.rt, rval)
	}

	return nil
}
