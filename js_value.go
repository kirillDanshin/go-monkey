package monkey

/*
#cgo linux  LDFLAGS: -lmozjs185
#cgo darwin LDFLAGS: -lmozjs185

#include "js/jsapi.h"
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

func (v *Value) IsNull() bool {
	if C.JSVAL_IS_NULL(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v *Value) IsVoid() bool {
	if C.JSVAL_IS_VOID(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v *Value) IsInt() bool {
	if C.JSVAL_IS_INT(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v *Value) IsString() bool {
	if C.JSVAL_IS_STRING(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v *Value) IsNumber() bool {
	if C.JSVAL_IS_NUMBER(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v *Value) IsBoolean() bool {
	if C.JSVAL_IS_BOOLEAN(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v *Value) IsObject() bool {
	if C.JSVAL_IS_OBJECT(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v *Value) IsArray() bool {
	if v.IsObject() {
		var obj *C.JSObject
		if C.JS_ValueToObject(v.rt.cx, v.val, &obj) == C.JS_TRUE {
			return C.JS_IsArrayObject(v.rt.cx, obj) == C.JS_TRUE
		}
	}
	return false
}

func (v *Value) IsFunction() bool {
	v.rt.lock()
	defer v.rt.unlock()

	if v.IsObject() && C.JS_ObjectIsFunction(v.rt.cx, C.JSVAL_TO_OBJECT(v.val)) == C.JS_TRUE {
		return true
	}
	return false
}

// Try convert a value to String.
func (v *Value) ToString() string {
	v.rt.lock()
	defer v.rt.unlock()

	cstring := C.JS_EncodeString(v.rt.cx, C.JS_ValueToString(v.rt.cx, v.val))
	gostring := C.GoString(cstring)
	C.JS_free(v.rt.cx, unsafe.Pointer(cstring))
	return gostring
}

// Try convert a value to Int.
func (v *Value) ToInt() (int32, bool) {
	v.rt.lock()
	defer v.rt.unlock()

	var r C.int32
	if C.JS_ValueToInt32(v.rt.cx, v.val, &r) == C.JS_TRUE {
		return int32(r), true
	}
	return 0, false
}

// Try convert a value to Number.
func (v *Value) ToNumber() (float64, bool) {
	v.rt.lock()
	defer v.rt.unlock()

	var r C.jsdouble
	if C.JS_ValueToNumber(v.rt.cx, v.val, &r) == C.JS_TRUE {
		return float64(r), true
	}
	return 0, false
}

// Try convert a value to Boolean.
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

// Try convert a value to Object.
func (v *Value) ToObject() (*Object, bool) {
	v.rt.lock()
	defer v.rt.unlock()

	var obj *C.JSObject
	if C.JS_ValueToObject(v.rt.cx, v.val, &obj) == C.JS_TRUE {
		return newObject(v.rt, obj), true
	}
	return nil, false
}

// Try convert a value to Array.
func (v *Value) ToArray() (*Array, bool) {
	v.rt.lock()
	defer v.rt.unlock()

	var obj *C.JSObject
	if C.JS_ValueToObject(v.rt.cx, v.val, &obj) == C.JS_TRUE {
		if C.JS_IsArrayObject(v.rt.cx, obj) == C.JS_TRUE {
			return newArray(v.rt, obj), true
		}
	}
	return nil, false
}

// !!! This function will make program fault when the value not a really String.
func (v *Value) String() string {
	v.rt.lock()
	defer v.rt.unlock()

	cstring := C.JS_EncodeString(v.rt.cx, C.JSVAL_TO_STRING(v.val))
	gostring := C.GoString(cstring)
	C.JS_free(v.rt.cx, unsafe.Pointer(cstring))

	return gostring
}

// !!! This function will make program fault when the value not a really Int.
func (v *Value) Int() int32 {
	return int32(C.JSVAL_TO_INT(v.val))
}

// !!! This function will make program fault when the value not a really Number.
func (v *Value) Number() float64 {
	return float64(C.JSVAL_TO_DOUBLE(v.val))
}

// !!! This function will make program fault when the value not a really Boolean.
func (v *Value) Boolean() bool {
	if C.JSVAL_TO_BOOLEAN(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

// !!! This function will make program fault when the value not a really Object.
func (v *Value) Object() *Object {
	v.rt.lock()
	defer v.rt.unlock()
	return newObject(v.rt, C.JSVAL_TO_OBJECT(v.val))
}

// !!! This function will make program fault when the value not a really Object.
func (v *Value) Array() *Array {
	v.rt.lock()
	defer v.rt.unlock()
	return newArray(v.rt, C.JSVAL_TO_OBJECT(v.val))
}

func (v *Value) Call(argv []*Value) (*Value, bool) {
	v.rt.lock()
	defer v.rt.unlock()

	argv2 := make([]C.jsval, len(argv))
	for i := 0; i < len(argv); i++ {
		argv2[i] = argv[i].val
	}
	argv3 := unsafe.Pointer(&argv2)
	argv4 := (*reflect.SliceHeader)(argv3).Data
	argv5 := (*C.jsval)(unsafe.Pointer(argv4))

	r := v.rt.Void()
	if C.JS_CallFunctionValue(v.rt.cx, nil, v.val, C.uintN(len(argv)), argv5, &r.val) == C.JS_TRUE {
		return r, true
	}

	return r, false
}

func (v *Value) TypeName() string {
	v.rt.lock()
	defer v.rt.unlock()

	jstype := C.JS_TypeOfValue(v.rt.cx, v.val)
	return C.GoString(C.JS_GetTypeName(v.rt.cx, jstype))
}
