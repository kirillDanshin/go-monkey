What is
=======

This package is SpiderMonkey wrapper for Go.

You can use this package to embed JavaScript into your Go program.

This package just newborn, use in production enviroment at your own risk!

Why make
========

You can found "go-v8" or "gomonkey" project on the github.

Thy have same purpose: Embed JavaScript into Golang program, make it more dynamic.

But I found all of existing projects are not powerful enough.

For example "go-v8" use JSON to pass data between Go and JavaScript, each callback has a significant performance cost.

So I make one by myself.

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
	runtime, err1 := js.NewRuntime(8 * 1024 * 1024)
	if err1 != nil {
		panic(err1)
	}

	// Evaluate script
	if value := runtime.Eval("'Hello ' + 'World!'"); value != nil {
		println(value.ToString())
	}

	// Define a function and call it
	runtime.DefineFunction("println",
		func(rt *js.Runtime, args []*js.Value) *js.Value {
			for i := 0; i < len(args); i++ {
				fmt.Print(args[i])
			}
			fmt.Println()
			return runtime.Void()
		},
	)

	runtime.Eval("println('Hello Function!')")

	// Compile once, run many times
	if script := runtime.Compile(
		"println('Hello Compiler!')",
		"<no name>", 0,
	); script != nil {
		script.Execute()
		script.Execute()
		script.Execute()
	}

	// Error handler
	runtime.SetErrorReporter(func(report *js.ErrorReport) {
		println(fmt.Sprintf(
			"%s:%d: %s",
			report.FileName, report.LineNum, report.Message,
		))
		if report.LineBuf != "" {
			println("\t", report.LineBuf)
		}
	})

	// Trigger an error
	runtime.Eval("abc()")
}
```
This code will output:

```
Hello World!
Hello Function!
Hello Compiler!
Hello Compiler!
Hello Compiler!
Eval():0: ReferenceError: abc is not defined
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
	// Create Script Runtime
	runtime, err1 := js.NewRuntime(8 * 1024 * 1024)
	if err1 != nil {
		panic(err1)
	}

	// String
	if value := runtime.Eval("'abc'"); assert(value != nil) {
		assert(value.IsString())
		assert(value.ToString() == "abc")
	}

	// Int
	if value := runtime.Eval("123456789"); assert(value != nil) {
		assert(value.IsInt())

		if value1, ok1 := value.ToInt(); assert(ok1) {
			assert(value1 == 123456789)
		}
	}

	// Number
	if value := runtime.Eval("12345.6789"); assert(value != nil) {
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
	runtime, err1 := js.NewRuntime(8 * 1024 * 1024)
	if err1 != nil {
		panic(err1)
	}

	// Return an array from JavaScript
	if value := runtime.Eval("[123, 456];"); assert(value != nil) {
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
	if ok := runtime.DefineFunction("get_data",
		func(rt *js.Runtime, args []*js.Value) *js.Value {
			array := rt.NewArray()
			array.SetInt(0, 100)
			array.SetInt(1, 200)
			return array.ToValue()
		},
	); assert(ok) {
		if value := runtime.Eval("get_data()"); assert(value != nil) {
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
	runtime, err1 := js.NewRuntime(8 * 1024 * 1024)
	if err1 != nil {
		panic(err1)
	}

	// Return an object from JavaScript
	if value := runtime.Eval("x={a:123}"); assert(value != nil) {
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
	ok := runtime.DefineFunction("get_data",
		func(rt *js.Runtime, args []*js.Value) *js.Value {
			obj := rt.NewObject()
			obj.SetInt("abc", 100)
			obj.SetInt("def", 200)
			return obj.ToValue()
		},
	)

	assert(ok)

	if value := runtime.Eval("get_data()"); assert(value != nil) {
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

func main() {
	// Create Script Runtime
	runtime, err1 := js.NewRuntime(8 * 1024 * 1024)
	if err1 != nil {
		panic(err1)
	}

	// Return Object With Property Getter And Setter From Go
	ok := runtime.DefineFunction("get_data",
		func(rt *js.Runtime, args []*js.Value) *js.Value {
			obj := rt.NewObject()

			// Define the property 'abc' with getter and setter
			var propValue int32 = 123
			ok := obj.DefineProperty("abc",
				// Init value
				runtime.Int(propValue),
				// T getter callback called each time
				// JavaScript code accesses the property's value
				func(o *js.Object) *js.Value {
					return o.Runtime().Int(propValue)
				},
				// The setter callback is called each time
				// JavaScript code assigns to the property
				func(o *js.Object, val *js.Value) {
					var ok bool
					propValue, ok = val.ToInt()
					assert(ok)
				},
				0,
			)

			assert(ok)

			return obj.ToValue()
		},
	)

	assert(ok)

	if value := runtime.Eval(`
		a = get_data();
		v1 = a.abc;
		a.abc = 456;
		v2 = a.abc;
		[v1, v2];
	`); assert(value != nil) {
		// Type check
		assert(value.IsArray())
		array := value.ToArray()
		assert(array != nil)

		// Length check
		assert(array.GetLength() == 2)

		// Check v1
		value1, ok1 := array.GetInt(0)
		assert(ok1)
		assert(value1 == 123)

		// Check v2
		value2, ok2 := array.GetInt(1)
		assert(ok2)
		assert(value2 == 456)
	}
}
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
	runtime, err1 := js.NewRuntime(8 * 1024 * 1024)
	if err1 != nil {
		panic(err1)
	}

	runtime.DefineFunction("println",
		func(rt *js.Runtime, args []*js.Value) *js.Value {
			for i := 0; i < len(args); i++ {
				fmt.Print(args[i])
			}
			fmt.Println()
			return runtime.Void()
		},
	)

	wg := new(sync.WaitGroup)

	// One runtime instance used by many goroutines
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000; j++ {
				v := runtime.Eval("println('Hello World!')")
				assert(v != nil)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
```
