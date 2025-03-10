package lock

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	DefaultAcquireTimeout = time.Millisecond * 200
	DefaultHoldLockExpired = time.Millisecond * 500
	m.Run()
}

func TestLocker(t *testing.T) {
	t.Run("顺序等待加锁", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewLocker()
		token := lk.Lock()
		ast.NotZero(token)
		go func() {
			token2 := lk.Lock()
			ast.NotZero(token2)
			lk.Unlock(token2)
		}()
		time.Sleep(time.Millisecond * 100)
		lk.Unlock(token)
	})
	t.Run("锁等待超时", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewLocker()
		token := lk.Lock()
		ast.NotZero(token)
		go func() {
			ast.Panics(func() {
				token2 := lk.Lock()
				ast.Zero(token2)
			})
		}()
		time.Sleep(time.Second)
		lk.Unlock(token)
	})
	t.Run("成功获取到超时的锁", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewLocker()
		token := lk.Lock()
		ast.Equal(Token(1), token)
		go func() {
			// 唤醒后再等待100ms成功获取到判定超时的锁
			time.Sleep(time.Millisecond * 400)
			token2 := lk.Lock()
			defer lk.Unlock(token2)
			ast.Equal(Token(2), token2)
		}()
		time.Sleep(time.Second)
		lk.Unlock(token)
	})
	t.Run("死锁", func(t *testing.T) {
		ast := assert.New(t)
		lk1 := NewLocker()
		lk2 := NewLocker()
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			// 获取到lk1的锁，休眠100ms，获取lk2时200ms超时，既第300ms时退出
			defer wg.Done()
			t1 := lk1.Lock()
			ast.Equal(Token(1), t1)
			defer lk1.Unlock(t1)
			time.Sleep(time.Millisecond * 100)
			ast.Panics(func() {
				lk2.Lock()
			})
		}()
		go func() {
			// 获取到lk2的锁，休眠200ms，获取lk1时等待100ms时协程1由于超时退出释放了lk1，最终正常获取到lk1
			defer wg.Done()
			t2 := lk2.Lock()
			ast.Equal(Token(1), t2)
			defer lk2.Unlock(t2)
			time.Sleep(time.Millisecond * 200)
			t1 := lk1.Lock()
			ast.Equal(Token(2), t1)
		}()
		wg.Wait()
	})
	t.Run("回调", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewLocker()
		cb := 0
		tk := lk.Lock(WithLockCallback(func() {
			cb = 1
		}))
		ast.Equal(1, cb)
		lk.Unlock(tk, WithUnlockCallback(func() {
			cb = 2
		}))
		ast.Equal(2, cb)
	})
}

func TestRWLocker(t *testing.T) {
	// 写锁
	t.Run("顺序等待加锁", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewRWLocker()
		token := lk.Lock()
		ast.NotZero(token)
		go func() {
			token2 := lk.Lock()
			ast.NotZero(token2)
			lk.Unlock(token2)
		}()
		time.Sleep(time.Millisecond * 100)
		lk.Unlock(token)
	})
	t.Run("锁等待超时", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewRWLocker()
		token := lk.Lock()
		ast.NotZero(token)
		go func() {
			ast.Panics(func() {
				token2 := lk.Lock()
				ast.Zero(token2)
			})
		}()
		time.Sleep(time.Second)
		lk.Unlock(token)
	})
	t.Run("成功获取到超时的锁", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewRWLocker()
		token := lk.Lock()
		ast.Equal(Token(1), token)
		go func() {
			// 唤醒后再等待100ms成功获取到判定超时的锁
			time.Sleep(time.Millisecond * 400)
			token2 := lk.Lock()
			defer lk.Unlock(token2)
			ast.Equal(Token(2), token2)
		}()
		time.Sleep(time.Second)
		lk.Unlock(token)
	})
	t.Run("死锁", func(t *testing.T) {
		ast := assert.New(t)
		lk1 := NewRWLocker()
		lk2 := NewRWLocker()
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			// 获取到lk1的锁，休眠100ms，获取lk2时200ms超时，既第300ms时退出
			defer wg.Done()
			t1 := lk1.Lock()
			ast.Equal(Token(1), t1)
			defer lk1.Unlock(t1)
			time.Sleep(time.Millisecond * 100)
			ast.Panics(func() {
				lk2.Lock()
			})
		}()
		go func() {
			// 获取到lk2的锁，休眠200ms，获取lk1时等待100ms时协程1由于超时退出释放了lk1，最终正常获取到lk1
			defer wg.Done()
			t2 := lk2.Lock()
			ast.Equal(Token(1), t2)
			defer lk2.Unlock(t2)
			time.Sleep(time.Millisecond * 200)
			t1 := lk1.Lock()
			ast.Equal(Token(2), t1)
		}()
		wg.Wait()
	})
	// 读写锁
	t.Run("同时获取读锁", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewRWLocker()
		t1 := lk.RLock()
		ast.Equal(Token(1), t1)
		go func() {
			t2 := lk.RLock()
			defer lk.RUnlock(t2)
			ast.Equal(Token(2), t2)
		}()
		time.Sleep(time.Second)
		lk.RUnlock(t1)
	})
	t.Run("先读后写", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewRWLocker()
		t1 := lk.RLock()
		ast.Equal(Token(1), t1)
		go func() {
			// 锁获取超时
			ast.Panics(func() {
				lk.Lock()
			})
		}()
		time.Sleep(time.Second)
		lk.RUnlock(t1)
		t2 := lk.Lock()
		ast.Equal(Token(2), t2)
		lk.Unlock(t2)
	})
	t.Run("先写后读", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewRWLocker()
		t1 := lk.Lock()
		ast.Equal(Token(1), t1)
		go func() {
			// 锁获取超时
			ast.Panics(func() {
				lk.RLock()
			})
		}()
		time.Sleep(time.Second)
		lk.Unlock(t1)
		t2 := lk.RLock()
		ast.Equal(Token(2), t2)
		lk.RUnlock(t2)
	})
	t.Run("多个读写锁等待，优先写锁", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewRWLocker()
		t1 := lk.RLock()
		ast.Equal(Token(1), t1)
		wg := sync.WaitGroup{}
		write := func() {
			defer wg.Done()
			t2 := lk.Lock()
			defer lk.Unlock(t2)
			fmt.Println("write lock:", t2)
			// 2-3
			ast.InDelta(2, int(t2), 1)
		}
		read := func() {
			defer wg.Done()
			t2 := lk.RLock()
			defer lk.RUnlock(t2)
			fmt.Println("read lock:", t2)
			// 4-6
			ast.InDelta(5, int(t2), 1)
		}
		// 先尝试获取写锁
		wg.Add(1)
		go write()
		time.Sleep(time.Millisecond * 10)
		// 再尝试获取三个读锁
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go read()
		}
		time.Sleep(time.Millisecond * 10)
		// 最后再尝试获取一个写锁
		wg.Add(1)
		go write()
		time.Sleep(time.Millisecond * 10)
		lk.RUnlock(t1)
		wg.Wait()
		// 结论: 会先获取所有写锁(2次),再获取读锁(3次)
	})
	t.Run("回调", func(t *testing.T) {
		ast := assert.New(t)
		lk := NewRWLocker()
		cb := 0
		tk := lk.Lock(WithLockCallback(func() {
			cb = 1
		}))
		ast.Equal(1, cb)
		lk.Unlock(tk, WithUnlockCallback(func() {
			cb = 2
		}))
		ast.Equal(2, cb)
		tk = lk.RLock(WithLockCallback(func() {
			cb = 3
		}))
		ast.Equal(3, cb)
		lk.RUnlock(tk, WithUnlockCallback(func() {
			cb = 4
		}))
		ast.Equal(4, cb)
	})
}

// BenchmarkSyncLock-12            51267316                22.93 ns/op            8 B/op          1 allocs/op
// BenchmarkLock-12                 5301124               236.6 ns/op            48 B/op          1 allocs/op
// BenchmarkSyncRWLock-12          21413419                54.82 ns/op           24 B/op          1 allocs/op
// BenchmarkRWLock-12               1741074               684.7 ns/op           400 B/op          3 allocs/op
func BenchmarkSyncLock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lk := sync.Mutex{}
		lk.Lock()
		lk.Unlock()
	}
}

func BenchmarkLock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lk := NewLocker()
		t1 := lk.Lock()
		lk.Unlock(t1)
	}
}

func BenchmarkSyncRWLock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lk := sync.RWMutex{}
		lk.RLock()
		lk.RUnlock()
		lk.Lock()
		lk.Unlock()
	}
}

func BenchmarkRWLock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lk := NewRWLocker()
		t1 := lk.RLock()
		lk.RUnlock(t1)
		t2 := lk.Lock()
		lk.RUnlock(t2)
	}
}
