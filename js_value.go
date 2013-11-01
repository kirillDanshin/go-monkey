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
	cx  *Context
	val C.jsval
}

func newValue(cx *Context, val C.jsval) *Value {
	result := &Value{cx, val}

	C.JS_AddValueRoot(cx.jscx, &result.val)

	runtime.SetFinalizer(result, func(v *Value) {
		C.JS_RemoveValueRoot(v.cx.jscx, &v.val)
	})

	return result
}

func (v *Value) Runtime() *Runtime {
	return v.cx.rt
}

func (v *Value) Context() *Context {
	return v.cx
}

func (v *Value) String() string {
	return v.ToString()
}

func (v *Value) TypeName() string {
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	return C.GoString(C.JS_GetTypeName(v.cx.jscx, C.JS_TypeOfValue(v.cx.jscx, v.val)))
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
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	return v.IsObject() && C.JS_IsArrayObject(
		v.cx.jscx, C.JSVAL_TO_OBJECT(v.val),
	) == C.JS_TRUE
}

func (v *Value) IsFunction() bool {
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	return v.IsObject() && C.JS_ObjectIsFunction(
		v.cx.jscx, C.JSVAL_TO_OBJECT(v.val),
	) == C.JS_TRUE
}

// Convert a value to Int.
func (v *Value) ToInt() (int32, bool) {
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	var r C.int32
	if C.JS_ValueToInt32(v.cx.jscx, v.val, &r) == C.JS_TRUE {
		return int32(r), true
	}
	return 0, false
}

// Convert a value to Number.
func (v *Value) ToNumber() (float64, bool) {
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	var r C.jsdouble
	if C.JS_ValueToNumber(v.cx.jscx, v.val, &r) == C.JS_TRUE {
		return float64(r), true
	}
	return 0, false
}

// Convert a value to Boolean.
func (v *Value) ToBoolean() (bool, bool) {
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	var r C.JSBool
	if C.JS_ValueToBoolean(v.cx.jscx, v.val, &r) == C.JS_TRUE {
		if r == C.JS_TRUE {
			return true, true
		}
		return false, true
	}
	return false, false
}

// Convert a value to String.
func (v *Value) ToString() string {
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	cstring := C.JS_EncodeString(v.cx.jscx, C.JS_ValueToString(v.cx.jscx, v.val))
	gostring := C.GoString(cstring)
	C.JS_free(v.cx.jscx, unsafe.Pointer(cstring))

	return gostring
}

// Convert a value to Object.
func (v *Value) ToObject() *Object {
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	var obj *C.JSObject
	if C.JS_ValueToObject(v.cx.jscx, v.val, &obj) == C.JS_TRUE {
		return newObject(v.cx, obj, nil)
	}

	return nil
}

// Convert a value to Array.
func (v *Value) ToArray() *Array {
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	var obj *C.JSObject
	if C.JS_ValueToObject(v.cx.jscx, v.val, &obj) == C.JS_TRUE {
		if C.JS_IsArrayObject(v.cx.jscx, obj) == C.JS_TRUE {
			return newArray(v.cx, obj)
		}
	}

	return nil
}

// Convert a JavaScript value to Go object
func (v *Value) ToGo() interface{} {
	var ret interface{}

	switch {
	case v.IsBoolean():
		ret, _ = v.ToBoolean()
	case v.IsInt():
		ret, _ = v.ToInt()
	case v.IsNumber():
		ret, _ = v.ToNumber()
	case v.IsString():
		ret = v.String()
	case v.IsObject():
		ret = v.ToObject().ToGo()
	case v.IsArray():
		arr := v.ToArray()
		goArr := make([]interface{}, arr.GetLength())
		for i := 0; i < arr.GetLength(); i++ {
			goArr[i] = arr.GetElement(i).ToGo()
		}
		ret = goArr
	case v.IsNull():
		ret = nil
	default:
		panic("unsupported js type")
	}

	return ret
}

// Call a function value
func (v *Value) Call(argv []*Value) *Value {
	v.cx.rt.lock()
	defer v.cx.rt.unlock()

	argv2 := make([]C.jsval, len(argv))
	for i := 0; i < len(argv); i++ {
		argv2[i] = argv[i].val
	}
	argv3 := unsafe.Pointer(&argv2)
	argv4 := (*reflect.SliceHeader)(argv3).Data
	argv5 := (*C.jsval)(unsafe.Pointer(argv4))

	var rval C.jsval
	if C.JS_CallFunctionValue(v.cx.jscx, nil, v.val, C.uintN(len(argv)), argv5, &rval) == C.JS_TRUE {
		return newValue(v.cx, rval)
	}

	return nil
}
