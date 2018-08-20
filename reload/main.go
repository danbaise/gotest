package main

import (
	"errors"
	"flag"
	"fmt"
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

var (
	server   *http.Server
	listener net.Listener
	graceful = flag.Bool("graceful", false, "listen on fd open 3 (internal use only)")
)

var cxt context.Context = context.Background()

func handler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(20 * time.Second)
	w.Write([]byte("hello world233333!!!!"))
}

func main() {
	flag.Parse()
	fmt.Println(os.Getpid())

	http.HandleFunc("/hello", handler)
	server = &http.Server{Addr: ":9999"}

	var err error
	if *graceful {
		log.Print("main: Listening to existing file descriptor 3.")
		// cmd.ExtraFiles: If non-nil, entry i becomes file descriptor 3+i.
		// when we put socket FD at the first entry, it will always be 3(0+3)
		f := os.NewFile(3, "")
		listener, err = net.FileListener(f)
	} else {
		log.Print("main: Listening on a new file descriptor.")
		listener, err = net.Listen("tcp", server.Addr)
	}

	if err != nil {
		log.Fatalf("listener error: %v", err)
	}

	go func() {
		// server.Shutdown() stops Serve() immediately, thus server.Serve() should not be in main goroutine
		err = server.Serve(listener)
		log.Printf("server.Serve err: %v\n", err)
	}()

	signalHandler()
	log.Printf("signal end")
}

func reload() error {
	tl, ok := listener.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not tcp listener")
	}

	f, err := tl.File()
	if err != nil {
		return err
	}

	args := []string{"-graceful"}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// put socket FD at the first entry
	cmd.ExtraFiles = []*os.File{f}
	return cmd.Start()
}

func signalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR2)
	for {
		switch <-ch {
		case syscall.SIGUSR2:
			// reload
			log.Printf("reload")
			err := reload()
			if err != nil {
				log.Fatalf("graceful restart error: %v", err)
			}
			server.Shutdown(cxt)
			log.Printf("graceful reload")
			return
		}
	}
}
