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

	// Return Array From JavaScript
	if value, ok := runtime.Eval("[123, 456];"); assert(ok) {
		// Type Check
		assert(value.IsArray())
		array := value.Array()
		assert(array != nil)

		// Length Check
		length, ok := array.GetLength()
		assert(ok)
		assert(length == 2)

		// Get First Item
		value1, ok1 := array.GetElement(0)
		assert(ok1)
		assert(value1.IsInt())
		assert(value1.Int() == 123)

		// Get Second Item
		value2, ok2 := array.GetElement(1)
		assert(ok2)
		assert(value2.IsInt())
		assert(value2.Int() == 456)

		// Set First Item
		assert(array.SetElement(0, runtime.Int(789)))
		value3, ok3 := array.GetElement(0)
		assert(ok3)
		assert(value3.IsInt())
		assert(value3.Int() == 789)

		// Grows
		assert(array.SetLength(3))
		length2, _ := array.GetLength()
		assert(length2 == 3)
	}

	// Return Array From Go
	if ok := runtime.DefineFunction("get_data",
		func(argv []*js.Value) (*js.Value, bool) {
			array := runtime.NewArray()
			array.SetElement(0, runtime.Int(100))
			array.SetElement(1, runtime.Int(200))
			return array.ToValue(), true
		},
	); assert(ok) {
		if value, ok := runtime.Eval("get_data()"); assert(ok) {
			// Type Check
			assert(value.IsArray())
			array := value.Array()
			assert(array != nil)

			// Length Check
			length, ok := array.GetLength()
			assert(ok)
			assert(length == 2)

			// Get First Item
			value1, ok1 := array.GetElement(0)
			assert(ok1)
			assert(value1.IsInt())
			assert(value1.Int() == 100)

			// Get Second Item
			value2, ok2 := array.GetElement(1)
			assert(ok2)
			assert(value2.IsInt())
			assert(value2.Int() == 200)
		}
	}

	runtime.Dispose()
}
