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

	// Return a function object from JavaScript
	if value, ok := runtime.Eval("function(a,b){ return a+b; }"); assert(ok) {
		// Type check
		assert(value.IsFunction())

		// Call
		value1, ok1 := value.Call([]*js.Value{
			runtime.Int(10),
			runtime.Int(20),
		})

		// Result check
		assert(ok1)
		assert(value1.IsNumber())
		assert(value1.Int() == 30)
	}

	// Define a function that return an object with function from Go
	if ok := runtime.DefineFunction("get_data",
		func(rt *js.Runtime, argv []*js.Value) (*js.Value, bool) {
			obj := rt.NewObject()

			ok := obj.DefineFunction("abc", func(rt *js.Runtime, argv []*js.Value) (*js.Value, bool) {
				return rt.Int(100), true
			})

			assert(ok)

			return obj.ToValue(), true
		},
	); assert(ok) {
		if value, ok := runtime.Eval(`
			a = get_data(); 
			a.abc();
		`); assert(ok) {
			assert(value.IsInt())
			assert(value.Int() == 100)
		}
	}
}
