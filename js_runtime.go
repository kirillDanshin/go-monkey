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

var defaultRuntime Runtime

// JavaScript Runtime
type Runtime struct {
	maxbytes       uint32
	jsrt           *C.JSRuntime
	goid           int32
	disposed       int64
	initChan       chan bool
	workChan       chan jswork
	closeChan      chan int
	ctxDisposeChan chan *Context
	objDisposeChan chan *Object
	aryDisposeChan chan *Array
	valDisposeChan chan *Value
	sptDisposeChan chan *Script
}

type jswork struct {
	callback   func()
	resultChan chan int
}

// Initializes the JavaScript runtime.
// @maxbytes Maximum number of allocated bytes after which garbage collection is run.
func NewRuntime(maxbytes uint32) *Runtime {
	r := new(Runtime)
	r.maxbytes = maxbytes
	r.initChan = make(chan bool)

	// make runtime only used in the creator thread
	go r.init()

	initSucceed := <-r.initChan

	if initSucceed {
		return r
	}

	return nil
}

func (r *Runtime) init() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	r.goid = goid.Get()

	r.jsrt = C.JS_NewRuntime(C.uint32(r.maxbytes))
	if r.jsrt == nil {
		r.initChan <- false
		return
	}

	r.workChan = make(chan jswork, 20)
	r.closeChan = make(chan int, 1)
	r.ctxDisposeChan = make(chan *Context, 50)
	r.objDisposeChan = make(chan *Object, 100)
	r.aryDisposeChan = make(chan *Array, 100)
	r.valDisposeChan = make(chan *Value, 100)
	r.sptDisposeChan = make(chan *Script, 100)

	runtime.SetFinalizer(r, func(r *Runtime) {
		r.Dispose()
	})

	r.initChan <- true

L:
	for {
		select {
		case work := <-r.workChan:
			work.callback()
			work.resultChan <- 1
		case ctx := <-r.ctxDisposeChan:
			C.JS_DestroyContext(ctx.jscx)
		case obj := <-r.objDisposeChan:
			C.JS_RemoveObjectRoot(obj.cx.jscx, &obj.obj)
		case ary := <-r.aryDisposeChan:
			C.JS_RemoveObjectRoot(ary.cx.jscx, &ary.obj)
		case val := <-r.valDisposeChan:
			C.JS_RemoveValueRoot(val.cx.jscx, &val.val)
		case spt := <-r.sptDisposeChan:
			C.JS_RemoveObjectRoot(spt.cx.jscx, &spt.obj)
		case _ = <-r.closeChan:
			break L
		}
	}

	C.JS_DestroyRuntime(r.jsrt)
}

// Exeucte the callback in runtime creator thread.
// Use this method to avoid Monkey internal call it many times.
// See the benchmarks in "monkey_test.go".
func (r *Runtime) Use(callback func()) {
	if goid.Get() == r.goid {
		callback()
	} else {
		work := jswork{
			callback:   callback,
			resultChan: make(chan int, 1),
		}

		r.workChan <- work
		<-work.resultChan
	}
}

// Free by manual
func (r *Runtime) Dispose() {
	if atomic.CompareAndSwapInt64(&r.disposed, 0, 1) {
		r.closeChan <- 1
	}
}
