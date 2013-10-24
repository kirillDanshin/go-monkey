package main

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

	wg := new(sync.WaitGroup)

	// One runtime instance used by many goroutines
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000; j++ {
				_, ok := runtime.Eval("println('Hello World!')")
				assert(ok)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
