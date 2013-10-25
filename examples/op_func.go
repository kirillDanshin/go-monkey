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
	if value := runtime.Eval("function(a,b){ return a+b; }"); assert(value != nil) {
		// Type check
		assert(value.IsFunction())

		// Call
		value1 := value.Call([]*js.Value{
			runtime.Int(10),
			runtime.Int(20),
		})

		// Result check
		assert(value1 != nil)
		assert(value1.IsNumber())

		if value2, ok2 := value1.ToNumber(); assert(ok2) {
			assert(value2 == 30)
		}
	}

	// Define a function that return an object with function from Go
	ok := runtime.DefineFunction("get_data",
		func(rt *js.Runtime, args []*js.Value) *js.Value {
			obj := rt.NewObject()

			ok := obj.DefineFunction("abc",
				func(rt *js.Runtime, args []*js.Value) *js.Value {
					return rt.Int(100)
				},
			)

			assert(ok)

			return obj.ToValue()
		},
	)

	assert(ok)

	if value := runtime.Eval(`
		a = get_data(); 
		a.abc();
	`); assert(value != nil) {
		assert(value.IsInt())

		if value2, ok2 := value.ToInt(); assert(ok2) {
			assert(value2 == 100)
		}
	}
}
