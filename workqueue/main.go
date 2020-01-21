/*
@Author : zj
@Time : 2020/1/20
*/
package main

import (
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	gob.Register(&Us{})
}

type Us struct {
	Delay uint
}

func (u *Us) Do() {
	fmt.Printf("延迟 %d 时间 %s\n", u.Delay, time.Now().Format("2006-01-02 15:04:05"))
}

func main() {
	delay := NewDelayTask(60).Start()
	NewSnapshot().Run(delay.data)

	for i := 0; i < 100; i++ {
		num := rand.Int31n(30) + 1
		err := delay.Save(uuid.New().String(), &Us{Delay: uint(num)}, uint(num))
		if err != nil {
			fmt.Println(err)
		}
	}

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
}
