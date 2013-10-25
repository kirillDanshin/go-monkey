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

	// Define a function and call it
	runtime.DefineFunction("println", func(rt *js.Runtime, args []*js.Value) *js.Value {
		for i := 0; i < len(args); i++ {
			print(args[i].ToString())
		}
		println()
		return runtime.Void()
	})

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
