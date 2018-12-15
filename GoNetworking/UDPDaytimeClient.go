package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s host:port", os.Args[0])
	}
	service := os.Args[1]
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	checkErrorUdpClient(err)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkErrorUdpClient(err)
	_, err = conn.Write([]byte("Anything"))
	checkErrorUdpClient(err)

	var buf [512]byte
	n, err := conn.Read(buf[0:])
	checkErrorUdpClient(err)
	fmt.Println(string(buf[0:n]))
	os.Exit(0)
}

func checkErrorUdpClient(e error) {
	if e != nil {
		logrus.Info("Fatal Error occurred %v", e.Error())
		os.Exit(1)
	}
}
