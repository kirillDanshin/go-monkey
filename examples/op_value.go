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

	// String
	if value, ok := runtime.Eval("'abc'"); assert(ok) {
		assert(value.IsString())
		assert(value.String() == "abc")
	}

	// Int
	if value, ok := runtime.Eval("123456789"); assert(ok) {
		assert(value.IsInt())
		assert(value.Int() == 123456789)
	}

	// Number
	if value, ok := runtime.Eval("12345.6789"); assert(ok) {
		assert(value.IsNumber())
		assert(value.Number() == 12345.6789)
	}
}
