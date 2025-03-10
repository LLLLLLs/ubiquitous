package patchtime

import (
	"sync/atomic"
	"time"
	_ "unsafe"
)

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64)

//go:linkname runtimeNano runtime.nanotime
func runtimeNano() int64

var offset atomic.Int64

func patchNow() time.Time {
	sec, nsec, _ := now()
	return time.Unix(sec+offset.Load(), int64(nsec))
}
