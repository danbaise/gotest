/*
@Author : zj
@Time : 2020/1/17
*/
package main

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

var successCount uint64

type User struct {
	Username string
	Password string
}

func (u *User) Processing() {
	x := rand.Intn(2)
	time.Sleep(time.Duration(x) * time.Second)
	atomic.AddUint64(&successCount, 1)
}

func TestExample(t *testing.T) {

	config := &Config{
		MinWork:              2,
		MaxWork:              10,
		JobMax:               50,
		WorkIdleTime:         5 * time.Second,
		DilatationFactor:     0.75,
		DilatationMultiplier: 2,
		TickerTime:           10 * time.Millisecond,
	}
	wq := NewWorkqueue(config).Start()

	go func() {
		time.Sleep(5 * time.Second)
		wq.SetMaxWork(50)
		time.Sleep(5 * time.Second)
		wq.SetMaxWork(30)
		time.Sleep(5 * time.Second)
		wq.SetMaxWork(80)
	}()

	go func() {
		time.Sleep(7 * time.Second)
		wn := atomic.LoadUint64(&wq.workNum)
		if wq.cfg.MaxWork != 50 && (wn >= wq.cfg.MinWork && wn <= wq.cfg.MaxWork) {
			t.Error("SetMaxWork 50 error")
		}

		time.Sleep(7 * time.Second)
		wn = atomic.LoadUint64(&wq.workNum)
		if wq.cfg.MaxWork != 30 && (wn >= wq.cfg.MinWork && wn <= wq.cfg.MaxWork) {
			t.Error("SetMaxWork 30 error")
		}

		time.Sleep(7 * time.Second)
		wn = atomic.LoadUint64(&wq.workNum)
		if wq.cfg.MaxWork != 80 && (wn >= wq.cfg.MinWork && wn <= wq.cfg.MaxWork) {
			t.Error("SetMaxWork 80 error")
		}

	}()

	var errcount uint64
	for i := 0; i < 2000; i++ {
		err := wq.Put(&User{Username: "testusername", Password: "testpassword"}, 2*time.Second)
		if err != nil {
			errcount++
		}
	}

	<-time.After(20 * time.Second)
	successCounts := atomic.LoadUint64(&successCount)

	if (successCounts + errcount) != 2000 {
		t.Error("error")
	}
}

