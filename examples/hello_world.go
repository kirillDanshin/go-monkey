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
	if value, ok := runtime.Eval("'Hello ' + 'World!'"); ok {
		println(value.ToString())
	}

	// Call built-in function
	runtime.Eval("println('Hello Built-in Function!')")

	// Compile once, run many times
	if script := runtime.Compile(
		"println('Hello Compiler!')",
		"<no name>", 0,
	); script != nil {
		script.Execute()
		script.Execute()
		script.Execute()
	}

	// Define a function
	if runtime.DefineFunction("add",
		func(rt *js.Runtime, argv []*js.Value) (*js.Value, bool) {
			if len(argv) != 2 {
				return rt.Null(), false
			}
			return rt.Int(argv[0].Int() + argv[1].Int()), true
		},
	) {
		// Call the function
		if value, ok := runtime.Eval("add(100, 200)"); ok {
			println(value.Int())
		}
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
