package main

import js "github.com/realint/monkey"

func assert(c bool) bool {
	if !c {
		panic("assert failed")
	}
	return c
}

type T struct {
	abc int32
}

func main() {
	// Create Script Runtime
	runtime := js.NewRuntime(8 * 1024 * 1024)

	// Create script context
	context := runtime.NewContext()

	// Return Object With Property Getter And Setter From Go
	ok := context.DefineFunction("get_data", func(f *js.Func) {
		cx := f.Context()
		obj := cx.NewObject(&T{123})

		// Define the property 'abc' with getter and setter
		ok := obj.DefineProperty("abc",
			// Init value
			cx.Void(),
			// T getter callback called each time
			// JavaScript code accesses the property's value
			func(g *js.Getter) {
				t := g.Object().GetPrivate().(*T)
				if g.Name() == "abc" {
					g.Return(cx.Int(t.abc))
				} else {
					panic("undefined property " + g.Name())
				}
			},
			// The setter callback is called each time
			// JavaScript code assigns to the property
			func(s *js.Setter) {
				t := s.Object().GetPrivate().(*T)
				if s.Name() == "abc" {
					d, ok := s.Value().ToInt()
					assert(ok)
					t.abc = d
				} else {
					panic("undefined property " + s.Name())
				}
			},
			0,
		)

		assert(ok)

		f.Return(obj.ToValue())
	})

	assert(ok)

	if value := context.Eval(`
		a = get_data();
		v1 = a.abc;
		a.abc = 456;
		v2 = a.abc;
		[v1, v2, a];
	`); assert(value != nil) {
		// Type check
		assert(value.IsArray())
		array := value.ToArray()
		assert(array != nil)

		// Length check
		assert(array.GetLength() == 3)

		// Check v1
		value1, ok1 := array.GetInt(0)
		assert(ok1)
		assert(value1 == 123)

		// Check v2
		value2, ok2 := array.GetInt(1)
		assert(ok2)
		assert(value2 == 456)

		// Check v3
		obj := array.GetObject(2)
		assert(obj != nil)
		t, ok3 := obj.GetPrivate().(*T)
		assert(ok3)
		assert(t.abc == 456)
	}

	if value := context.Eval(`
		var a = {};
		a.field1 = 1;
		a.field2 = "hello";
		a.field3 = 1.2;
		a.field4 = true;
		a.field5 = {};
		a.func1 = function(){};
		a;
	`); assert(value != nil) {
		obj := value.ToObject()
		assert(obj != nil)

		keys := obj.Keys()
		assert(len(keys) == 6)
		assert(keys[0] == "field1")
		assert(keys[1] == "field2")
		assert(keys[2] == "field3")
		assert(keys[3] == "field4")
		assert(keys[4] == "field5")
		assert(keys[5] == "func1")

		keys = obj.GetProperty("field5").ToObject().Keys()
		assert(len(keys) == 0)
	}
}
