package monkey

/*
#cgo linux  LDFLAGS: -lmozjs185
#cgo darwin LDFLAGS: -lmozjs185

#include "monkey.h"
*/
import "C"
import (
	"github.com/realint/monkey/goid"
	"runtime"
	"sync/atomic"
)

// JavaScript Runtime
type Runtime struct {
	jsrt      *C.JSRuntime
	goid      int32
	disposed  int64
	workChan  chan jswork
	closeChan chan int
}

type jswork struct {
	callback   func()
	resultChan chan int
}

// Initializes the JavaScript runtime.
// @maxbytes Maximum number of allocated bytes after which garbage collection is run.
func NewRuntime(maxbytes uint32) *Runtime {
	r := new(Runtime)

	r.jsrt = C.JS_NewRuntime(C.uint32(maxbytes))
	if r.jsrt == nil {
		return nil
	}

	r.workChan = make(chan jswork)
	r.closeChan = make(chan int)

	runtime.SetFinalizer(r, func(r *Runtime) {
		r.Dispose()
	})

	go r.workLoop()

	return r
}

func (r *Runtime) workLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	r.goid = goid.Get()
L:
	for {
		select {
		case work := <-r.workChan:
			work.callback()
			work.resultChan <- 1
		case _ = <-r.closeChan:
			break L
		}
	}
}

func (r *Runtime) dowork(callback func()) {
	if goid.Get() == r.goid {
		callback()
	} else {
		work := jswork{
			callback:   callback,
			resultChan: make(chan int),
		}

		r.workChan <- work
		<-work.resultChan
	}
}

// Free by manual
func (r *Runtime) Dispose() {
	if atomic.CompareAndSwapInt64(&r.disposed, 0, 1) {
		r.closeChan <- 1
		C.JS_DestroyRuntime(r.jsrt)
	}
}
