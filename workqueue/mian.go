package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

var ix int64

type User struct {
	Username string
	Password string
}

func (u *User) Processing() {
	x := rand.Intn(10)
	time.Sleep(time.Duration(x) * time.Second)
	if u.Check() {
		fmt.Println("is ok")
	}
	ix++
}

func (u *User) Check() bool {
	if u.Username == u.Password {
		return true
	}
	return false
}

func main() {

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
		for {
			ticker := time.NewTicker(time.Millisecond * 1)
			<-ticker.C
			fmt.Println(atomic.LoadUint64(&wq.workNum), atomic.LoadUint64(&wq.processingNum), wq.cfg.MaxWork)
		}
	}()

	for i := 0; i < 1000; i++ {
		err := wq.Put(&User{Username: "123", Password: "1213"}, 2*time.Second)
		if err != nil {
			fmt.Println(i, err)
		}
	}

	time.Sleep(20 * time.Second)
	fmt.Println(atomic.LoadUint64(&wq.workNum), atomic.LoadUint64(&wq.processingNum), wq.cfg.MaxWork)
	wq.Stop()

	fmt.Println(atomic.LoadUint64(&wq.workNum), atomic.LoadUint64(&wq.processingNum), wq.cfg.MaxWork)
	fmt.Println(cap(wq.jobsChan), len(wq.jobsChan))
	fmt.Println(ix)
}
