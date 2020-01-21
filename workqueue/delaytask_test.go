/*
@Author : zj
@Time : 2020/1/21
*/
package main

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

var successDelayTaskCount uint64

type delayTaskTest struct {
	Delay uint
}

func (d *delayTaskTest) Do() {
	time.Sleep(time.Second)
//	fmt.Printf("延迟 %d 执行 ，当前时间 %s\n ", d.Delay, time.Now().Format("2006-01-02 15:04:05"))
	atomic.AddUint64(&successDelayTaskCount, 1)
}

func TestDelayTask(t *testing.T) {
	fmt.Println("TestDelayTask")
	delay := NewDelayTask(60).Start()

	for i := 0; i < 5000; i++ {
		num := rand.Int31n(124) + int32(1)
		err := delay.Save(uuid.New().String(), &delayTaskTest{Delay: uint(num)}, uint(num))
		if err != nil {
			fmt.Println(err)
		}
	}
	<-time.After(126 * time.Second)

	c := atomic.LoadUint64(&successDelayTaskCount)
	if c != 5000 {
		fmt.Println(c)
		t.Error("error")
	}

}