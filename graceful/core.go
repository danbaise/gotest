package graceful

/*
package main

import (
	"fmt"
	"github.com/danbaise/gotest/graceful"
	"net/http"
	"os"
	"time"
)

const ADDRESS = ":9999"

func handler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(20 * time.Second)
	w.Write([]byte("hello world233333!!!!"))
}
func main() {
	fmt.Println("start PID", os.Getpid())
	http.HandleFunc("/hello", handler)
	server := &http.Server{Addr: ADDRESS}

	conf := &graceful.Conf{Server: server, Timeout: 20 * time.Second}
	graceful.New(conf).Serve()
}
*/

import (
	"errors"
	"golang.org/x/net/context"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const (
	ENV = "GRACEFUL"
)

type Graceful struct {
	*Conf
	Listener net.Listener
}

type Conf struct {
	Server  *http.Server
	Timeout time.Duration
}

func New(c *Conf) *Graceful {
	g := &Graceful{}
	g.Conf = c
	return g
}

func (g *Graceful) Getenv(key string) string {
	return os.Getenv(key)
}

func (g *Graceful) SetENV() error {
	return os.Setenv(ENV, "ON")
}

func (g *Graceful) SetListener() error {
	var err error
	if g.Getenv(ENV) == "ON" {
		log.Print("main: Listening to existing file descriptor 3.")
		// cmd.ExtraFiles: If non-nil, entry i becomes file descriptor 3+i.
		// when we put socket FD at the first entry, it will always be 3(0+3)
		f := os.NewFile(3, "")
		g.Listener, err = net.FileListener(f)
	} else {
		log.Print("main: Listening on a new file descriptor.")
		g.Listener, err = net.Listen("tcp", g.Server.Addr)
	}

	if err != nil {
		return err
	}
	return nil
}

func (g *Graceful) Serve() {
	g.SetListener()
	go func() {
		err := g.Server.Serve(g.Listener)
		log.Printf("server.Serve err: %v\n", err)
	}()
	g.signalHandler()
}

func (g *Graceful) reload() error {
	tl, ok := g.Listener.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not tcp listener")
	}

	f, err := tl.File()
	if err != nil {
		return err
	}
	g.SetENV()
	args := os.Args[1:]
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// put socket FD at the first entry
	cmd.ExtraFiles = []*os.File{f}
	return cmd.Start()
}

func (g *Graceful) signalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR2)
	for {
		switch <-ch {
		case syscall.SIGUSR2:
			// reload
			err := g.reload()
			if err != nil {
				log.Fatalf("graceful restart error: %v", err)
			}
			ctx, _ := context.WithTimeout(context.Background(), g.Timeout)
			g.Server.Shutdown(ctx)
			log.Printf("graceful reload")
			return
		}
	}
}
