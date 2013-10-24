package main

import "fmt"
import js "github.com/realint/monkey"

func main() {
	// Create Script Runtime
	runtime, err1 := js.NewRuntime(8 * 1024 * 1024)
	if err1 != nil {
		panic(err1)
	}

	// Evaluate Script
	if value, ok := runtime.Eval("'Hello ' + 'World!'"); ok {
		println(value.ToString())
	}

	// Built-in Function
	runtime.Eval("println('Hello Built-in Function!')")

	// Compile Once, Run Many Times
	if script := runtime.Compile(
		"println('Hello Compiler!')",
		"<no name>", 0,
	); script != nil {
		script.Execute()
		script.Execute()
		script.Execute()
	}

	// Define Function
	if runtime.DefineFunction("add",
		func(argv []*js.Value) (*js.Value, bool) {
			if len(argv) != 2 {
				return runtime.Null(), false
			}
			return runtime.Int(argv[0].Int() + argv[1].Int()), true
		},
	) {
		if value, ok := runtime.Eval("add(100, 200)"); ok {
			println(value.Int())
		}
	}

	// Error Handle
	runtime.SetErrorReporter(func(report *js.ErrorReport) {
		println(fmt.Sprintf(
			"%s:%d: %s",
			report.FileName, report.LineNum, report.Message,
		))
		if report.LineBuf != "" {
			println("\t", report.LineBuf)
		}
	})

	// Trigger An Error
	runtime.Eval("abc()")

	// Say Good Bye
	runtime.Dispose()
}
