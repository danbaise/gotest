package main


import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrWorkqueueClosed  = errors.New("working pool closed")
	ErrJobTimeout  = errors.New("add new job timeout")
	ErrLessMinWork = errors.New("not less than minwork")
)

type Jober interface {
	Processing()
}

type Config struct {
	MinWork              uint64
	MaxWork              uint64
	WorkIdleTime         time.Duration
	JobMax               uint64
	DilatationFactor     float64
	DilatationMultiplier uint64
	TickerTime           time.Duration
}

type Workqueue struct {
	cfg           *Config
	jobsChan      chan Jober
	workNum       uint64
	processingNum uint64
	reduceChan    chan struct{}
	exitChan      chan struct{}
	waitGroup     sync.WaitGroup
	isClosed      bool
}

func NewWorkqueue(cfg *Config) *Workqueue {
	return &Workqueue{
		cfg:           cfg,
		jobsChan:      make(chan Jober, cfg.JobMax),
		workNum:       0,
		processingNum: 0,
		exitChan:      make(chan struct{}),
		waitGroup:     sync.WaitGroup{},
		isClosed:      false,
	}
}

func (wq *Workqueue) Stop() {
	close(wq.exitChan)
	wq.isClosed = true
	wq.waitGroup.Wait()
}

func (wq *Workqueue) Start() *Workqueue {
	//初始化worker goroutines 数量
	addWork := func(num uint64) {
		atomic.AddUint64(&wq.workNum, num)
		var i uint64
		for i = 0; i < num; i++ {
			go wq.work()
		}
	}
	addWork(wq.cfg.MinWork)
	ticker := time.NewTicker(wq.cfg.TickerTime)
	go func() {
		for {
			select {
			case <-wq.exitChan:
				return
			case <-ticker.C:
				pn := atomic.LoadUint64(&wq.processingNum)
				wn := atomic.LoadUint64(&wq.workNum)
				// 扩容
				if (float64(pn) / float64(wn)) > wq.cfg.DilatationFactor {
					if (wn * wq.cfg.DilatationMultiplier) <= wq.cfg.MaxWork {
						addWork(wn * (wq.cfg.DilatationMultiplier - 1))
					} else {
						if wq.cfg.MaxWork > wn {
							addWork(wq.cfg.MaxWork - wn)
						}
					}
				}
			}
		}
	}()
	return wq
}

func (wq *Workqueue) SetMaxWork(num uint64) error {
	if num < wq.cfg.MinWork {
		return ErrLessMinWork
	}
	wq.cfg.MaxWork = num
	wn := atomic.LoadUint64(&wq.workNum)
	if num < wn {
		reduceNum := wn - num
		wq.reduceChan = make(chan struct{}, reduceNum)
		var i uint64
		for i = 0; i < reduceNum; i++ {
			wq.reduceChan <- struct{}{}
		}
	}
	return nil
}

func (wq *Workqueue) work() {
	wq.waitGroup.Add(1)
	defer func() {
		atomic.AddUint64(&wq.workNum, ^uint64(0))
		wq.waitGroup.Done()
	}()
	for {
		select {
		case job := <-wq.jobsChan:
			func() {
				atomic.AddUint64(&wq.processingNum, 1)
				defer func() {
					atomic.AddUint64(&wq.processingNum, ^uint64(0))
				}()
				job.Processing()
			}()
		case <-time.After(wq.cfg.WorkIdleTime):
			if wq.checkMinWork() {
				return
			}
		case <-wq.reduceChan:
			if wq.checkMinWork() {
				return
			}
		case <-wq.exitChan:
			return
		}
	}
}

func (wq *Workqueue) checkMinWork() bool {
	if atomic.LoadUint64(&wq.workNum) > wq.cfg.MinWork {
		return true
	}
	return false
}

func (wq *Workqueue) Put(job Jober, timeout time.Duration) error {
	if wq.isClosed {
		return ErrWorkqueueClosed
	}
	select {
	case wq.jobsChan <- job:
		return nil
	case <-time.After(timeout):
		return ErrJobTimeout
	case <-wq.exitChan:
		return ErrWorkqueueClosed
	}
}