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

	context.DefineFunction("println",
		func(cx *js.Context, args []*js.Value) *js.Value {
			for i := 0; i < len(args); i++ {
				fmt.Print(args[i])
			}
			fmt.Println()
			return cx.Void()
		},
	)

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
