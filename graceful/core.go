package graceful

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
)

const (
	ENV = "GRACEFUL"
)

type Graceful struct {
	*Conf
	Listener net.Listener
	Server   *http.Server
}

type Conf struct {
	Cxt     context.Context
	Address string
}

func New() *Graceful {
	return &Graceful{}
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
		g.Listener, err = net.Listen("tcp", g.Address)
	}

	if err != nil {
		return err
	}
	return nil
}

func (g *Graceful) SetConfig(cxt context.Context, server *http.Server, address string) {
	g.Conf = &Conf{Address: address, Cxt: cxt}
	g.Server = server
}

func (g *Graceful) Run() {
	g.SetListener()
	go g.signalHandler()
}

func (g *Graceful) reload() (error, bool) {
	tl, ok := g.Listener.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not tcp listener"), false
	}

	f, err := tl.File()
	if err != nil {
		return err, false
	}
	g.SetENV()
	args := os.Args[1:]
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// put socket FD at the first entry
	cmd.ExtraFiles = []*os.File{f}
	err = cmd.Start()
	if err != nil {
		log.Fatalf("graceful restart error: %v", err)
		return err, false
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("graceful restart error: %v", err)
		return err, false
	}

	return nil, true
}

func (g *Graceful) signalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR2)
	for {
		switch <-ch {
		case syscall.SIGUSR2:
			// reload
			err, ok := g.reload()
			if !ok {
				log.Fatalf("graceful restart error: %v", err)
				continue
			}
			g.Server.Shutdown(g.Cxt)
			log.Printf("graceful reload")
			return
		}
	}
}
