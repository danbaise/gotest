package reload

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	timeout = 30 * time.Second
	tag     = "listener"
)

// Smooth 平滑reload结构体
type Smooth struct {
	Addr     string
	Timeout  time.Duration
	listener net.Listener
	Tag      string
	Exec     Executor
}

// Executor 执行者
type Executor interface {
	Run(l net.Listener, args []string)
	Wait()
}

// SetExecutor SetExecutor
func (s *Smooth) SetExecutor(e Executor) {
	s.Exec = e
}

// SetAddr 地址.
func (s *Smooth) SetAddr(addr string) {

	s.Addr = addr
}

// SetTimeout 设置超时时间.
func (s *Smooth) SetTimeout(t ...int) {

	if len(t) > 0 && t[0] > 0 {
		s.Timeout = time.Duration(t[0]) * time.Second
		return
	}

	s.Timeout = timeout

}

// SetTag 设置标签
func (s *Smooth) SetTag(t ...string) {

	if len(t) > 0 && t[0] != "" {
		s.Tag = t[0]
		return
	}

	s.Tag = tag
}

// isMaster 判断是主进程
func (s *Smooth) isMaster() bool {
	_, ok := os.LookupEnv(s.Tag)
	return !ok
}

// GenListener 生成监听者
func (s *Smooth) GenListener() (err error) {

	if s.isMaster() {
		s.listener, err = net.Listen("tcp", s.Addr)
	} else {
		f := os.NewFile(3, "")
		s.listener, err = net.FileListener(f)
	}

	if err != nil {
		return
	}

	// 这里运行程序
	s.Exec.Run(s.listener, os.Args)
	return
}

// genEnv genEnv
func (s *Smooth) genEnv() string {

	return fmt.Sprintf("%s=%v", s.Tag, 1)
}

// Reload Reload
func (s *Smooth) Reload() (err error) {

	var f *os.File
	switch l := s.listener.(type) {
	case *net.TCPListener:
		f, err = l.File()
		if err != nil {
			return
		}
	default:
		return errors.New("listener is not tcp listener")
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{f}
	cmd.Env = []string{s.genEnv()}
	return cmd.Start()
}

// SignalHandler 信号处理
func (s *Smooth) SignalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {

		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			// stop
			signal.Stop(ch)
		case syscall.SIGUSR2:
			// reload
			err := s.Reload()
			if err != nil {
				log.Fatalf("smooth reload error: %v", err)
			}
			over := make(chan struct{}, 1)
			go func(chan struct{}) {
				s.Exec.Wait()
				over <- struct{}{}
			}(over)
			select {
			case <-over:
				os.Exit(1)
				break
			case <-time.After(s.Timeout):
				os.Exit(1)
			}

		}
	}
}

// KeepingPID 保存PID
func (s *Smooth) KeepingPID(PID int) {

	f, err := os.OpenFile("./core/curr.pid", os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		log.Panicf("store pid [%v]", err)
	}

	defer f.Close()
	f.WriteString(strconv.Itoa(PID))
}
