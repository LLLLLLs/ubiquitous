package patchtime

import (
	"fmt"
	"sync"
	"testing"
	"time"
	_ "unsafe"
)

func TestMain(m *testing.M) {
	Patch()
	m.Run()
}

func TestTimer(t *testing.T) {
	c := make(chan struct{})
	r := runtimeTimer{
		when: runtimeNano() + int64(time.Second),
		f:    sendTime,
		arg:  c,
	}
	startTimer(&r)
	fmt.Println(time.Now().String())
	<-c
	fmt.Println(time.Now().String())
}

func TestNewTimer(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second * 1)
		// 1 1 2 2 2
		for i := 0; i < 5; i++ {
			if i == 2 {
				ticker.Reset(time.Second * 2)
			}
			select {
			case <-ticker.C:
				fmt.Println("ticker:", i, runtimeNano())
			}
		}
	}()
	timer1 := time.NewTimer(time.Second * 3)
	timer2 := time.NewTimer(time.Second * 2)
	go func() {
		defer wg.Done()
		<-timer1.C
		fmt.Println("timer1 done", runtimeNano())
		withOffset(2)
	}()
	go func() {
		defer wg.Done()
		<-timer2.C
		fmt.Println("timer2 done", runtimeNano())
		timer1.Reset(time.Second * 4)
	}()
	wg.Wait()
}

func TestOffset(t *testing.T) {
	fmt.Println(runtimeNano())
	timer := time.NewTimer(time.Minute)
	withOffset(59)
	<-timer.C
	fmt.Println(runtimeNano())
}
