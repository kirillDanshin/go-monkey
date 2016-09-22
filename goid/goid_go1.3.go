// +build !go1.4

package goid

// Get returns the id of the current goroutine.
func Get() int64
