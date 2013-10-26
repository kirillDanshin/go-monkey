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
)

// JavaScript Runtime
type Runtime struct {
	jsrt      *C.JSRuntime
	lockBy    int32
	lockLevel int
	mutex     sync.Mutex
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
		C.JS_DestroyRuntime(r.jsrt)
	})

	return r
}

// Because we can't prevent Go to execute a JavaScript that maybe will execute another JavaScript by invoke Go function.
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
