package patchtime

import (
	"container/list"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe"
)

//go:linkname startTimer time.startTimer
func startTimer(*runtimeTimer)

//go:linkname stopTimer time.stopTimer
func stopTimer(*runtimeTimer) bool

//go:linkname resetTimer time.resetTimer
func resetTimer(*runtimeTimer, int64) bool

type runtimeTimer struct {
	pp       uintptr
	when     int64
	period   int64
	f        func(any, uintptr) // NOTE: must not be closure
	arg      any
	seq      uintptr
	nextwhen int64
	status   uint32
}

type timer struct {
	id     int64
	when   int64
	period int64
	c      chan time.Time
	tt     interface{} // 关联 time.Time 或 time.Ticker
}

var manager = &timerManager{
	lock:         sync.RWMutex{},
	id:           new(atomic.Int64),
	c:            make(chan time.Time, 1),
	runtimeTimer: nil,
	mTimer:       map[int64]*list.Element{},
	timers:       list.New(),
	ttMapper:     sync.Map{},
}

type timerManager struct {
	lock         sync.RWMutex
	id           *atomic.Int64
	c            chan time.Time
	runtimeTimer *runtimeTimer
	mTimer       map[int64]*list.Element
	timers       *list.List
	ttMapper     sync.Map // 映射 [*time.Timer|*time.Ticker] -> timer.id Reset和Stop时用来寻找对应的timer
}

func (tm *timerManager) refreshWithOffsetChanged() {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	tm.adjustRuntimeTimer()
}

func (tm *timerManager) add(t *timer) {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	tm.addNoLock(t)
}

func (tm *timerManager) addNoLock(t *timer, noAdjust ...bool) {
	node := tm.insert(t)
	tm.mTimer[t.id] = node
	if node != tm.timers.Front() || (len(noAdjust) > 0 && noAdjust[0]) {
		return
	}
	tm.adjustRuntimeTimer()
}

func (tm *timerManager) insert(t *timer) *list.Element {
	if tm.timers.Len() == 0 {
		return tm.timers.PushFront(t)
	}
	for ele := tm.timers.Front(); ele != nil; ele = ele.Next() {
		if t.when < ele.Value.(*timer).when {
			return tm.timers.InsertBefore(t, ele)
		}
	}
	return tm.timers.PushBack(t)
}

func (tm *timerManager) reset(tt interface{}, d time.Duration, period time.Duration) bool {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	_id, has := tm.ttMapper.Load(tt)
	if !has {
		return false
	}
	id := _id.(int64)
	ele, has := tm.mTimer[id]
	if !has {
		tm.ttMapper.Delete(tt)
		return false
	}
	tm.removeNoLock(id, true)
	mt := ele.Value.(*timer)
	mt.when = Now().UnixNano() + int64(d)
	mt.period = int64(period)
	tm.addNoLock(mt, true)
	tm.adjustRuntimeTimer()
	return true
}

func (tm *timerManager) stop(tt interface{}) bool {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	mt, has := tm.ttMapper.LoadAndDelete(tt)
	if !has {
		return false
	}
	return tm.removeNoLock(mt.(int64))
}

func (tm *timerManager) removeNoLock(id int64, noAdjust ...bool) bool {
	ele, has := tm.mTimer[id]
	if !has {
		return false
	}
	root := tm.timers.Front()
	tm.timers.Remove(ele)
	if ele != root || (len(noAdjust) > 0 && noAdjust[0]) {
		return true
	}
	tm.adjustRuntimeTimer()
	return true
}

func (tm *timerManager) adjustRuntimeTimer() {
	if tm.timers.Len() == 0 && tm.runtimeTimer == nil {
		return
	}
	if tm.timers.Len() == 0 && tm.runtimeTimer != nil {
		stopTimer(tm.runtimeTimer)
		tm.runtimeTimer = nil
		return
	}
	if tm.runtimeTimer == nil {
		tm.startRuntimeTimer()
		return
	}
	resetTimer(tm.runtimeTimer, tm.nextWhen())
}

func (tm *timerManager) startRuntimeTimer() {
	tm.runtimeTimer = &runtimeTimer{
		when: tm.nextWhen(),
		f:    sendTime,
		arg:  tm.c,
	}
	startTimer(tm.runtimeTimer)
}

func (tm *timerManager) nextWhen() int64 {
	duration := tm.timers.Front().Value.(*timer).when - time.Now().UnixNano()
	return runtimeNano() + duration
}

func (tm *timerManager) consume() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("consume panic", r, debug.Stack())
			go tm.consume()
		}
	}()
	for {
		select {
		case <-tm.c:
			tm.consumeOnce()
		}
	}
}

func (tm *timerManager) consumeOnce() {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	tm.runtimeTimer = nil
	for {
		if tm.handle() {
			break
		}
	}
	tm.adjustRuntimeTimer()
}

func (tm *timerManager) handle() (finish bool) {
	ele := tm.timers.Front()
	if ele == nil {
		finish = true
		return
	}
	t := ele.Value.(*timer)
	sendTime(t.c, 0)
	tm.removeNoLock(t.id, true)
	if t.period > 0 {
		t.when += t.period
		tm.addNoLock(t, true)
	} else {
		tm.ttMapper.Delete(t.tt)
	}
	finish = tm.timers.Front() == nil || tm.timers.Front().Value.(*timer).when >= Now().UnixNano()
	return
}

func (tm *timerManager) newTimer(d time.Duration) *time.Timer {
	mt := tm.addTimer(d, 0)
	t := &time.Timer{
		C: mt.c,
	}
	tm.ttMapper.Store(t, mt.id)
	mt.tt = t
	return t
}

func (tm *timerManager) newTicker(d time.Duration) *time.Ticker {
	mt := tm.addTimer(d, d)
	t := &time.Ticker{
		C: mt.c,
	}
	tm.ttMapper.Store(t, mt.id)
	mt.tt = t
	return t
}

func (tm *timerManager) addTimer(d time.Duration, period time.Duration) *timer {
	mt := &timer{
		id:     tm.id.Add(1),
		when:   Now().UnixNano() + int64(d),
		period: int64(period),
		c:      make(chan time.Time, 1),
	}
	tm.add(mt)
	return mt
}

func sendTime(c any, _ uintptr) {
	select {
	case c.(chan time.Time) <- Now():
	default:
	}
}
