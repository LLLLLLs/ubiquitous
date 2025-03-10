package lock

import (
	"errors"
	"time"
)

type Token int

var (
	ErrLockTimeout = errors.New("lock_timeout")

	DefaultAcquireTimeout  = time.Second      // 默认获取锁时超时时间
	DefaultHoldLockExpired = time.Second * 10 // 锁持有的超时时间(超过时间后不再持有)
)

func withTimeout(timeout time.Duration, f func() bool) bool {
	start := time.Now()
	sleep := time.Millisecond
	for {
		if f() {
			return false
		}
		if time.Since(start) > timeout {
			return true
		}
		time.Sleep(sleep)
		sleep *= 2
	}
}
