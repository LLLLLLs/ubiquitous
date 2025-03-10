package lock

import (
	"sync"
	"sync/atomic"
	"time"
)

// 读写锁对象
type rwLocker struct {
	write          int   // 使用int而不是bool值的原因，是为了与read保持类型的一致；
	writeIntention int32 // 写意向
	mutex          sync.Mutex
	token          Token               // token 计数
	expireAt       time.Time           // 写锁超时时间
	readTokens     map[Token]time.Time // 当前持有的所有读锁
}

// 尝试加写锁
// 返回值：加写锁是否成功
func (l *rwLocker) lock(hold time.Duration) Token {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.refresh()
	// 如果已经被锁定，则返回失败
	if l.write == 1 || len(l.readTokens) > 0 {
		return 0
	}
	// 否则，将写锁数量设置为１，并返回成功
	l.write = 1
	l.token++
	l.expireAt = time.Now().Add(hold)
	return l.token
}

func (l *rwLocker) refresh() {
	now := time.Now()
	if l.write == 1 && now.After(l.expireAt) {
		l.write = 0
		l.readTokens = map[Token]time.Time{}
		return
	}
	for tk, expired := range l.readTokens {
		if now.After(expired) {
			delete(l.readTokens, tk)
		}
	}
}

// 写锁定
// timeout:超时毫秒数,timeout<=0则将会死等
// 返回值：
// 成功或失败
// 如果失败，返回上一次成功加锁时的堆栈信息
// 如果失败，返回当前的堆栈信息
func (l *rwLocker) Lock(opt ...LockOption) (token Token) {
	conf := newConfig(opt...)
	atomic.AddInt32(&l.writeIntention, 1)
	defer atomic.AddInt32(&l.writeIntention, -1)
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

// 释放写锁
func (l *rwLocker) Unlock(token Token, opt ...UnlockOption) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.token != token || l.write == 0 {
		return false
	}
	conf := newUnlockConfig(opt...)
	l.write = 0
	conf.cb.invoke()
	return true
}

// 尝试加读锁
// 返回值：加读锁是否成功
func (l *rwLocker) rLock(hold time.Duration) Token {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.refresh()
	// 如果已经被锁或存在写意向的 则获取不到读锁
	if l.write == 1 || atomic.LoadInt32(&l.writeIntention) > 0 {
		return 0
	}
	l.token++
	l.readTokens[l.token] = time.Now().Add(hold)

	return l.token
}

func (l *rwLocker) RLock(opt ...LockOption) (token Token) {
	conf := newConfig(opt...)
	isTimeout := withTimeout(conf.acquireTimeout, func() bool {
		token = l.rLock(conf.lockHoldTimeout)
		return token != 0
	})
	if isTimeout {
		panic(ErrLockTimeout)
	}
	conf.cb.invoke()
	return
}

// 释放读锁
func (l *rwLocker) RUnlock(token Token, opt ...UnlockOption) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	expireAt, ok := l.readTokens[token]
	delete(l.readTokens, token)
	success := ok && time.Now().Before(expireAt)
	if success {
		conf := newUnlockConfig(opt...)
		conf.cb.invoke()
	}
	return success
}
