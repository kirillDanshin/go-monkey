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
		func(rt *js.Runtime, argv []*js.Value) (*js.Value, bool) {
			obj := rt.NewObject()

			ok := obj.DefineProperty("abc", runtime.Null(),
				func(o *js.Object) (*js.Value, bool) {
					return o.Runtime().Int(100), true
				},
				func(o *js.Object, val *js.Value) bool {
					// must set 200.
					assert(val.IsInt())
					assert(val.Int() == 200)
					return true
				},
				0,
			)

			assert(ok)

			return obj.ToValue(), true
		},
	); assert(ok) {
		if value, ok := runtime.Eval(`
			a = get_data();
			// must set 200, look at the code above.
			a.abc = 200;
			a.abc;
		`); assert(ok) {
			assert(value.IsInt())
			assert(value.Int() == 100)
		}
	}
}
