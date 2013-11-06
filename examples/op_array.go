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

	// Return an array from JavaScript
	if value := context.Eval("[123, 456];"); assert(value != nil) {
		// Check type
		assert(value.IsArray())
		array := value.ToArray()
		assert(array != nil)

		// Check length
		assert(array.GetLength() == 2)

		// Check first item
		value1, ok1 := array.GetInt(0)
		assert(ok1)
		assert(value1 == 123)

		// Check second item
		value2, ok2 := array.GetInt(1)
		assert(ok2)
		assert(value2 == 456)

		// Set first item
		assert(array.SetInt(0, 789))
		value3, ok3 := array.GetInt(0)
		assert(ok3)
		assert(value3 == 789)

		// Grows
		assert(array.SetLength(3))
		assert(array.GetLength() == 3)
	}

	// Return an array from Go
	if ok := context.DefineFunction("get_data", func(f *js.Func) {
		array := f.Context().NewArray()
		array.SetInt(0, 100)
		array.SetInt(1, 200)
		f.Return(array.ToValue())
	}); assert(ok) {
		if value := context.Eval("get_data()"); assert(value != nil) {
			// Check type
			assert(value.IsArray())
			array := value.ToArray()
			assert(array != nil)

			// Check length
			assert(array.GetLength() == 2)

			// Check first item
			value1, ok1 := array.GetInt(0)
			assert(ok1)
			assert(value1 == 100)

			// Check second item
			value2, ok2 := array.GetInt(1)
			assert(ok2)
			assert(value2 == 200)
		}
	}
}
