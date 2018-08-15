package main

import (
	"context"
	"github.com/danbaise/gotest/reload"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	sm     = reload.Smooth{}
	server = &http.Server{Addr: ":8899"}
	over   chan bool
)

// App App
type App struct {
}

func main() {
	sm.KeepingPID(os.Getpid())
	sm.SetExecutor(&App{})
	sm.SetAddr(server.Addr)
	sm.SetTimeout()
	sm.SetTag("smooth")
	sm.GenListener()
	sm.SignalHandler()
}

// Run Run
func (h *App) Run(l net.Listener, args []string) {
	http.HandleFunc("/hello", handler)
	go func() {
		err := server.Serve(l)
		log.Printf("server.Serve err: %v\n", err)
	}()

}

// Wait Wait
func (h *App) Wait() {
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	server.Shutdown(ctx)
}

func handler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(20 * time.Second)
	w.Write([]byte("This is responce"))
}
