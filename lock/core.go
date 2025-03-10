package lock

import (
	"sync"
	"time"
)

type Locker interface {
	Lock(opt ...LockOption) Token
	Unlock(token Token, opt ...UnlockOption) bool
}

type RWLocker interface {
	Locker
	RLock(opt ...LockOption) Token
	RUnlock(token Token, opt ...UnlockOption) bool
}

// 创建新的锁对象
func NewLocker() Locker {
	return &locker{
		mutex:    sync.Mutex{},
		write:    0,
		token:    0,
		expireAt: time.Time{},
	}
}

// 创建新的读写锁对象
func NewRWLocker() RWLocker {
	return &rwLocker{
		write:          0,
		writeIntention: 0,
		mutex:          sync.Mutex{},
		token:          0,
		expireAt:       time.Time{},
		readTokens:     map[Token]time.Time{},
	}
}
