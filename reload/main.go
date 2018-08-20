package main

import (
	"fmt"
	"github.com/danbaise/gotest/graceful"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"time"
)

const ADDRESS = ":9999"

func handler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(20 * time.Second)
	w.Write([]byte("hello world233333!!!!"))
}
func main() {

	http.HandleFunc("/hello", handler)
	server := &http.Server{Addr: ADDRESS}

	fmt.Println("start...")
	cxt, _ := context.WithTimeout(context.Background(), 30*time.Second)

	graceful := graceful.New()
	graceful.SetConfig(cxt, server, ADDRESS)
	graceful.Run()

	err := server.Serve(graceful.Listener)
	log.Printf("server.Serve err: %v\n", err)

}
