/*
@Author : zj
@Time : 2020/1/20
*/
package main

import (
	"errors"
	"sync"
	"time"
)

type DelayTasker interface {
	Save(interface{}, Jober, uint) error
	Stop()
	Delete(int, interface{}) bool
}

type Tasker interface {
	Jober
	CheckTime(int64) bool
	GetUuid() interface{}
	GetRunTime() int64
}

type task struct {
	uuid    interface{}
	mu      sync.Mutex
	runTime int64 //运行时间
	job     Jober
}

func (t *task) Do() {
	t.job.Do()
}

func (t *task) GetRunTime() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.runTime
}

func (t *task) CheckTime(now int64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.runTime == now {
		return true
	}
	return false
}

func (t *task) GetUuid() interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.uuid
}

type delayTask struct {
	data       []sync.Map
	exitChan   chan struct{}
	circleSize int
}

func NewDelayTask(size int) *delayTask {
	d := &delayTask{
		data:       make([]sync.Map, size),
		exitChan:   make(chan struct{}),
		circleSize: size,
	}
	return d
}

func (d *delayTask) Save(uuid interface{}, job Jober, t uint) error {
	if t == 0 {
		return errors.New("t can not be equal to 0")
	}
	curTime := time.Now().Unix()
	rt := curTime + int64(t)
	insertInx := rt % int64(d.circleSize)

	d.data[insertInx].Store(uuid, &task{uuid: uuid, runTime: rt, job: job})
	return nil
}

func (d *delayTask) Stop() {
	close(d.exitChan)
}

func (d *delayTask) Delete(inx int, uuid interface{}) bool {
	if _, ok := d.data[inx].Load(uuid); ok {
		//存在
		d.data[inx].Delete(uuid)
		return true
	}
	return false
}

func (d *delayTask) Start() *delayTask {
	go func() {
		interval := time.Second
		ticker := time.NewTicker(interval)
		for {
			select {
			case t := <-ticker.C:
				now := t.Unix()
				inx := now % int64(d.circleSize)
				f := func(k, v interface{}) bool {
					go func(now int64, t Tasker) {
						if t.CheckTime(now) {
							t.Do()
							d.Delete(int(inx), t.GetUuid())
						}
					}(now, v.(Tasker))
					return true
				}
				d.data[inx].Range(f)

			case <-d.exitChan:
				return
			}
		}
	}()
	return d
}
