package main

import js "github.com/realint/monkey"

func assert(c bool) bool {
	if !c {
		panic("assert failed")
	}
	return c
}

func main() {
	// Create Script Runtime
	runtime, err1 := js.NewRuntime(8 * 1024 * 1024)
	if err1 != nil {
		panic(err1)
	}

	// Return Object With Property Getter And Setter From Go
	if ok := runtime.DefineFunction("get_data",
		func(rt *js.Runtime, args []*js.Value) *js.Value {
			obj := rt.NewObject()

			// Define the property 'abc' with getter and setter
			var propValue int32 = 123
			ok := obj.DefineProperty("abc",
				// Init value
				runtime.Int(propValue),
				// T getter callback called each time
				// JavaScript code accesses the property's value
				func(o *js.Object) *js.Value {
					return o.Runtime().Int(propValue)
				},
				// The setter callback is called each time
				// JavaScript code assigns to the property
				func(o *js.Object, val *js.Value) bool {
					propValue = val.Int()
					return true
				},
				0,
			)

			assert(ok)

			return obj.ToValue()
		},
	); assert(ok) {
		if value, ok := runtime.Eval(`
			a = get_data();
			v1 = a.abc;
			a.abc = 456;
			v2 = a.abc;
			[v1, v2];
		`); assert(ok) {
			// Type check
			assert(value.IsArray())
			array := value.Array()
			assert(array != nil)

			// Length check
			length, ok := array.GetLength()
			assert(ok)
			assert(length == 2)

			// Get first item
			value1, ok1 := array.GetElement(0)
			assert(ok1)
			assert(value1.IsInt())
			assert(value1.Int() == 123)

			// Get second item
			value2, ok2 := array.GetElement(1)
			assert(ok2)
			assert(value2.IsInt())
			assert(value2.Int() == 456)
		}
	}
}
