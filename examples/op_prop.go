package main

import js "github.com/lazytiger/monkey"

func assert(c bool) bool {
	if !c {
		panic("assert failed")
	}
	return c
}

type T struct {
	abc int32
}

func main() {
	// Create Script Runtime
	runtime := js.NewRuntime(8 * 1024 * 1024)

	// Create script context
	context := runtime.NewContext()

	// Return Object With Property Getter And Setter From Go
	ok := context.DefineFunction("get_data",
		func(cx *js.Context, args []*js.Value) *js.Value {
			obj := cx.NewObject(&T{123})

			// Define the property 'abc' with getter and setter
			ok := obj.DefineProperty("abc",
				// Init value
				cx.Int(123),
				// T getter callback called each time
				// JavaScript code accesses the property's value
				func(o *js.Object) *js.Value {
					t := o.GoValue().(*T)
					return cx.Int(t.abc)
				},
				// The setter callback is called each time
				// JavaScript code assigns to the property
				func(o *js.Object, val *js.Value) {
					t := o.GoValue().(*T)
					d, ok := val.ToInt()
					assert(ok)
					t.abc = d
				},
				0,
			)

			assert(ok)

			return obj.ToValue()
		},
	)

	assert(ok)

	if value := context.Eval(`
		a = get_data();
		v1 = a.abc;
		a.abc = 456;
		v2 = a.abc;
		[v1, v2, a];
	`); assert(value != nil) {
		// Type check
		assert(value.IsArray())
		array := value.ToArray()
		assert(array != nil)

		// Length check
		assert(array.GetLength() == 3)

		// Check v1
		value1, ok1 := array.GetInt(0)
		assert(ok1)
		assert(value1 == 123)

		// Check v2
		value2, ok2 := array.GetInt(1)
		assert(ok2)
		assert(value2 == 456)

		// Check v3
		obj := array.GetObject(2)
		assert(obj != nil)
		t, ok3 := obj.GoValue().(*T)
		assert(ok3)
		assert(t.abc == 456)
	}
}
