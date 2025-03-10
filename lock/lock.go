package lock

import (
	"sync"
	"time"
)

// 写锁对象
type locker struct {
	mutex    sync.Mutex
	write    int // 使用int而不是bool值的原因，是为了与RWLocker中的read保持类型的一致；
	token    Token
	expireAt time.Time
}

// 内部锁
// 返回值：
// 加锁是否成功
func (l *locker) lock(hold time.Duration) Token {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 如果已经被锁定，则返回失败
	if l.write == 1 && time.Now().Before(l.expireAt) {
		return 0
	}

	// 否则，将写锁数量设置为１，并返回成功
	l.write = 1
	l.token++
	l.expireAt = time.Now().Add(hold)
	return l.token
}

// 尝试加锁，如果在指定的时间内失败，则会返回失败；否则返回成功
// token 锁标识，释放的时候要带着Token
func (l *locker) Lock(opt ...LockOption) (token Token) {
	conf := newConfig(opt...)
	isTimeout := withTimeout(conf.acquireTimeout, func() bool {
		token = l.lock(conf.lockHoldTimeout)
		return token != 0
	})
	if isTimeout {
		panic(ErrLockTimeout)
	}
	conf.cb.invoke()
	return
}

// 解锁
func (l *locker) Unlock(token Token, opt ...UnlockOption) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.token != token {
		return false
	}
	l.write = 0
	conf := newUnlockConfig(opt...)
	conf.cb.invoke()
	return true
}
