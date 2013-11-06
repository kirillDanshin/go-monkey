What is
=======

This package is SpiderMonkey wrapper for Go.

You can use this package to embed JavaScript into your Go program.

Why make
========

You can found some project like "go-v8" and "gomonkey" on GitHub.

Thy have same purpose: Embed JavaScript into Go program, make it more dynamic.

But I found all of existing projects are not powerful enough.

For example "go-v8" use JSON to pass data between Go and JavaScript, each callback has a significant performance cost.

And those packages all have thread-safe problem. 

V8 and SpiderMonkey both use thread-local storage. 

Embed them into Go program, you need to make sure each JS runtime used by creator thread only.

So monkey born.

It has a rich API and can be freely used in any multi-goroutine Go program.

Install
=======

You need install SpiderMonkey first.

Mac OS X:

```
brew install spidermonkey
```

Ubuntu:

```
sudo apt-get install libmozjs185-dev
```

Or compile by yourself ([reference](https://developer.mozilla.org/en-US/docs/SpiderMonkey/Build_Documentation)). 

And then install Monkey by "go get" command.

```
go get github.com/realint/monkey
```

Performance
===========

There are some benchmark test in "monkey_test.go".

You can run those test like this:

```
go test -bench="."
```

The benchmark result on my Mac:

```
Benchmark_ADD_IN_JS           200000     13410 ns/op
Benchmark_ADD_BY_JS           200000     15265 ns/op
Benchmark_ADD_BY_GO           100000     24779 ns/op
Benchmark_OOXX_IN_JS           10000    271113 ns/op
Benchmark_OOXX_IN_GO           50000     57133 ns/op
Benchmark_OOXX_BY_GO           20000     81031 ns/op
Benchmark_ADD_IN_JS_IN_USE   1000000      1446 ns/op
Benchmark_ADD_BY_JS_IN_USE   1000000      1427 ns/op
Benchmark_ADD_BY_GO_IN_USE    500000      7139 ns/op
Benchmark_OOXX_IN_JS_IN_USE    10000    262562 ns/op
Benchmark_OOXX_BY_GO_IN_USE    50000     63353 ns/op
```

Examples
========

All the example codes can be found in "examples" folder.

You can run all of the example codes like this:

```
go run examples/hello_world.go
```

Hello World
-----------

The "hello\_world.go" shows what Monkey can do.

```go
package main

import "fmt"
import js "github.com/realint/monkey"

func main() {
	// Create script runtime
	runtime := js.NewRuntime(8 * 1024 * 1024)

	// Create script context
	context := runtime.NewContext()

	// Evaluate script
	value := context.Eval("'Hello ' + 'World!'")
	println(value.ToString())

	// Define a function and call it
	context.DefineFunction("println", func(f *js.Func) {
		for i := 0; i < f.Argc(); i++ {
			fmt.Print(f.Argv(i))
		}
		fmt.Println()
		f.Return(f.Context().Void())
	})
	context.Eval("println('Hello Function!')")

	// Compile once, run many times
	script := context.Compile(
		"println('Hello Compiler!')",
		"<no name>", 0,
	)
	script.Execute()
	script.Execute()
	script.Execute()

	// Error handler
	context.SetErrorReporter(func(report *js.ErrorReport) {
		println(fmt.Sprintf(
			"%s:%d: %s",
			report.FileName, report.LineNum, report.Message,
		))
		if report.LineBuf != "" {
			println("\t", report.LineBuf)
		}
	})
	context.Eval("not_exists()")
}
```
This code will output:

```
Hello World!
Hello Function!
Hello Compiler!
Hello Compiler!
Hello Compiler!
Eval():0: ReferenceError: not_exists is not defined
```

Thread Safe
-----------

The "many\_many.go" shows Monkey is thread safe.

```go
package main

import "fmt"
import "sync"
import "runtime"
import js "github.com/realint/monkey"

func assert(c bool) bool {
	if !c {
		panic("assert failed")
	}
	return c
}

func main() {
	runtime.GOMAXPROCS(20)

	// Create script runtime
	runtime := js.NewRuntime(8 * 1024 * 1024)

	// Create script context
	context := runtime.NewContext()

	context.DefineFunction("println", func(f *js.Func) {
		for i := 0; i < f.Argc(); i++ {
			fmt.Print(f.Argv(i))
		}
		fmt.Println()
		f.Return(f.Context().Void())
	})

	wg := new(sync.WaitGroup)

	// One runtime instance used by many goroutines
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				v := context.Eval("println('Hello World!')")
				assert(v != nil)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
```

Value
-----

The "op_value.go" shows how to convert JS value to Go value.

```go
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
```

Function
--------

The "op_func.go" shows how to play with JS function value.

```go
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
	ok := context.DefineFunction("get_data", func(f *js.Func) {
		obj := f.Context().NewObject(nil)

		ok := obj.DefineFunction("abc",
			func(object *js.Object, name string, args []*js.Value) *js.Value {
				return f.Context().Int(100)
			},
		)

		assert(ok)

		f.Return(obj.ToValue())
	})

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
```

Array
-----

The "op_array.go" shows how to play with JS array.

```go
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
```

Object
------

The "op_object.go" shows how to play with JS object.

```go
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
```

Property
--------

The "op_prop.go" shows how to handle JavaScript object's property in Go.

```go
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
```

