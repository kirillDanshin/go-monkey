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
