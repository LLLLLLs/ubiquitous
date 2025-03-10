package lock

import "time"

type callback []func()

func (c *callback) append(cb func()) {
	*c = append(*c, cb)
}

func (c *callback) invoke() {
	for i := range *c {
		(*c)[i]()
	}
}

func newConfig(opt ...LockOption) *lockConfig {
	conf := &lockConfig{
		acquireTimeout:  DefaultAcquireTimeout,
		lockHoldTimeout: DefaultHoldLockExpired,
		cb:              new(callback),
	}
	for i := range opt {
		opt[i](conf)
	}
	return conf
}

type lockConfig struct {
	acquireTimeout  time.Duration // 获取锁的超时时间
	lockHoldTimeout time.Duration // 锁持有的超时时间
	cb              *callback     // 回调
}

type LockOption func(*lockConfig)

func WithAcquireTimeout(duration time.Duration) LockOption {
	return func(c *lockConfig) {
		c.acquireTimeout = duration
	}
}

func WithLockHoldTimeout(duration time.Duration) LockOption {
	return func(c *lockConfig) {
		c.lockHoldTimeout = duration
	}
}

func WithLockCallback(cb func()) LockOption {
	return func(c *lockConfig) {
		c.cb.append(cb)
	}
}

func newUnlockConfig(opt ...UnlockOption) *unlockConfig {
	conf := &unlockConfig{
		cb: new(callback),
	}
	for i := range opt {
		opt[i](conf)
	}
	return conf
}

type unlockConfig struct {
	cb *callback // 回调
}

type UnlockOption func(config *unlockConfig)

func WithUnlockCallback(cb func()) UnlockOption {
	return func(c *unlockConfig) {
		c.cb.append(cb)
	}
}
