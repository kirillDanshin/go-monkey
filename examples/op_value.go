package main

import js "github.com/realint/monkey"

func assert(c bool) bool {
	if !c {
		panic("assert failed")
	}
	return c
}

func main() {
	// Create script Runtime
	runtime := js.NewRuntime(8 * 1024 * 1024)

	// Create script context
	context := runtime.NewContext()

	// String
	if value := context.Eval("'abc'"); assert(value != nil) {
		assert(value.IsString())
		assert(value.ToString() == "abc")
	}

	// Int
	if value := context.Eval("123456789"); assert(value != nil) {
		assert(value.IsInt())

		if value1, ok1 := value.ToInt(); assert(ok1) {
			assert(value1 == 123456789)
		}
	}

	// Number
	if value := context.Eval("12345.6789"); assert(value != nil) {
		assert(value.IsNumber())

		if value1, ok1 := value.ToNumber(); assert(ok1) {
			assert(value1 == 12345.6789)
		}
	}
}
