package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"net"
	"os"
)

func main() {
	service := ":1201"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkErrorEcho(err)
	listner, err := net.ListenTCP("tcp", tcpAddr)
	checkErrorEcho(err)
	for {
		conn, err := listner.Accept()
		if err != nil {
			continue
		}
		go handleClientEcho(conn)

	}
}

func handleClientEcho(conn net.Conn) {

	defer conn.Close()

	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		fmt.Println(string(buf[0:]))
		_, err2 := conn.Write(buf[0:n])
		if err2 != nil {
			return
		}
	}
}

func checkErrorEcho(e error) {
	if e != nil {
		logrus.Info("Fatal Error occurred %v", e.Error())
		os.Exit(1)
	}
}
