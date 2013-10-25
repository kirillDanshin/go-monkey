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
	script := runtime.Compile(
		"println('Hello Compiler!')",
		"<no name>", 0,
	)

	script.Execute()
	script.Execute()
	script.Execute()

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
