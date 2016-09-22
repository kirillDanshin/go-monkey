// +build go1.4,!go1.5

package goid

import "unsafe"

var pointerSize = unsafe.Sizeof(uintptr(0))

// Backdoor access to runtime·getg().
func getg() uintptr // in goid_go1.4.s

// Get returns the id of the current goroutine.
func Get() int64 {
	// The goid is the 16th field in the G struct where each field is a
	// pointer, uintptr or padded to that size. See runtime.h from the
	// Go sources. I'm not aware of a cleaner way to determine the
	// offset.
	return *(*int64)(unsafe.Pointer(getg() + 16*pointerSize))
}
