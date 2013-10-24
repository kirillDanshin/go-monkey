package main

import js "github.com/realint/monkey"

func assert(c bool) bool {
	if !c {
		panic("assert failed")
	}
	return c
}

func main() {
	// Create script runtime
	runtime, err1 := js.NewRuntime(8 * 1024 * 1024)
	if err1 != nil {
		panic(err1)
	}

	// Return an object from JavaScript
	if value, ok := runtime.Eval("x={a:123}"); assert(ok) {
		// Type check
		assert(value.IsObject())
		obj := value.Object()

		// Get property 'a'
		value1, ok1 := obj.GetProperty("a")
		assert(ok1)
		assert(value1.IsInt())
		assert(value1.Int() == 123)

		// Set property 'b'
		assert(obj.SetProperty("b", runtime.Int(456)))
		value2, ok2 := obj.GetProperty("b")
		assert(ok2)
		assert(value2.IsInt())
		assert(value2.Int() == 456)
	}

	// Return and object From Go
	if ok := runtime.DefineFunction("get_data",
		func(rt *js.Runtime, argv []*js.Value) (*js.Value, bool) {
			obj := rt.NewObject()
			obj.SetProperty("abc", rt.Int(100))
			obj.SetProperty("def", rt.Int(200))
			return obj.ToValue(), true
		},
	); assert(ok) {
		if value, ok := runtime.Eval("get_data()"); assert(ok) {
			// Type check
			assert(value.IsObject())
			obj := value.Object()

			// Get property 'abc'
			value1, ok1 := obj.GetProperty("abc")
			assert(ok1)
			assert(value1.IsInt())
			assert(value1.Int() == 100)

			// Get property 'def'
			value2, ok2 := obj.GetProperty("def")
			assert(ok2)
			assert(value2.IsInt())
			assert(value2.Int() == 200)
		}
	}
}
