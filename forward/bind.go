package gateway

import (
	"io"
	"log"
	"math/rand"
	"net"
)

type gw struct {
	Config *Conf
}

var G = new(gw)

func Serv() {
	G.Config = Config()
	listenTCP()
}

func listenTCP() {
	address := G.Config.Get("LocalAddress").(string)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		panic(err)
	}
	l, err := net.ListenTCP("tcp", tcpAddr)
	for {
		sconn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go handler(sconn)
	}
}

func handler(sconn net.Conn) {

	BackendIps := G.Config.Get("BackendIps").([]string)
	n := rand.Intn(len(BackendIps))
	dconn, _ := net.Dial("tcp", BackendIps[n])

	log.Println("后端地址为：" + BackendIps[n])
	exitChan := make(chan bool, 1)
	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(dconn, sconn)
		if err != nil {
			log.Print("发送数据失败:%v\n", err)
		}
		defer func() {
			exitChan <- true
		}()
	}(sconn, dconn, exitChan)

	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(sconn, dconn)
		if err != nil {
			log.Print("接收数据失败:%v\n", err)
		}
		defer func() {
			exitChan <- true
		}()
	}(sconn, dconn, exitChan)

	for i := 0; i < 2; i++ {
		<-exitChan
	}
	defer func() {
		sconn.Close()
		dconn.Close()
		log.Println("链接关闭")
	}()
}
