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
	"sync"
	"sync/atomic"
)

// JavaScript Runtime
type Runtime struct {
	jsrt      *C.JSRuntime
	lockBy    int32
	lockLevel int
	mutex     sync.Mutex
	disposed  int64
}

// Initializes the JavaScript runtime.
// @maxbytes Maximum number of allocated bytes after which garbage collection is run.
func NewRuntime(maxbytes uint32) *Runtime {
	r := new(Runtime)

	r.jsrt = C.JS_NewRuntime(C.uint32(maxbytes))
	if r.jsrt == nil {
		return nil
	}

	runtime.SetFinalizer(r, func(r *Runtime) {
		r.Dispose()
	})

	return r
}

// Free by manual
func (r *Runtime) Dispose() {
	if atomic.CompareAndSwapInt64(&r.disposed, 0, 1) {
		C.JS_DestroyRuntime(r.jsrt)
	}
}

// Because we can't prevent Go execute a JavaScript that execute another JavaScript by call Go defined function.
// Like this: runtime.Eval("eval('1 + 1')")
// So I designed this lock mechanism to let runtime can lock by same goroutine many times.
func (r *Runtime) lock() {
	id := goid.Get()
	if r.lockBy != id {
		r.mutex.Lock()
		r.lockBy = id
	} else {
		r.lockLevel += 1
	}
}

func (r *Runtime) unlock() {
	r.lockLevel -= 1
	if r.lockLevel < 0 {
		r.lockLevel = 0
		r.lockBy = -1
		r.mutex.Unlock()
	}
}
