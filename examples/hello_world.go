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
