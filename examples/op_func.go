package main

import js "github.com/lazytiger/monkey"

func assert(c bool) bool {
	if !c {
		panic("assert failed")
	}
	return c
}

func main() {
	// Create script runtime
	runtime := js.NewRuntime(8 * 1024 * 1024)

	// Create script context
	context := runtime.NewContext()

	// Return a function object from JavaScript
	if value := context.Eval("function(a,b){ return a+b; }"); assert(value != nil) {
		// Type check
		assert(value.IsFunction())

		// Call
		value1 := value.Call([]*js.Value{
			context.Int(10),
			context.Int(20),
		})

		// Result check
		assert(value1 != nil)
		assert(value1.IsNumber())

		if value2, ok2 := value1.ToNumber(); assert(ok2) {
			assert(value2 == 30)
		}
	}

	// Define a function that return an object with function from Go
	ok := context.DefineFunction("get_data",
		func(cx *js.Context, args []*js.Value) *js.Value {
			obj := cx.NewObject(nil)

			ok := obj.DefineFunction("abc",
				func(cx *js.Context, args []*js.Value) *js.Value {
					return cx.Int(100)
				},
			)

			assert(ok)

			return obj.ToValue()
		},
	)

	assert(ok)

	if value := context.Eval(`
		a = get_data(); 
		a.abc();
	`); assert(value != nil) {
		assert(value.IsInt())

		if value2, ok2 := value.ToInt(); assert(ok2) {
			assert(value2 == 100)
		}
	}
}
