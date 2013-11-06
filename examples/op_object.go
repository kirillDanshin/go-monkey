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
	runtime := js.NewRuntime(8 * 1024 * 1024)

	// Create script context
	context := runtime.NewContext()

	// Return an object from JavaScript
	if value := context.Eval("x={a:123}"); assert(value != nil) {
		// Type check
		assert(value.IsObject())
		obj := value.ToObject()

		// Get property 'a'
		value1, ok1 := obj.GetInt("a")
		assert(ok1)
		assert(value1 == 123)

		// Set property 'b'
		assert(obj.SetInt("b", 456))
		value2, ok2 := obj.GetInt("b")
		assert(ok2)
		assert(value2 == 456)
	}

	// Return and object From Go
	ok := context.DefineFunction("get_data", func(f *js.Func) {
		obj := f.Context().NewObject(nil)
		obj.SetInt("abc", 100)
		obj.SetInt("def", 200)
		f.Return(obj.ToValue())
	})

	assert(ok)

	if value := context.Eval("get_data()"); assert(value != nil) {
		// Type check
		assert(value.IsObject())
		obj := value.ToObject()

		// Get property 'abc'
		value1, ok1 := obj.GetInt("abc")
		assert(ok1)
		assert(value1 == 100)

		// Get property 'def'
		value2, ok2 := obj.GetInt("def")
		assert(ok2)
		assert(value2 == 200)
	}
}
