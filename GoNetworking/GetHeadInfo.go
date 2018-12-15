package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s host:port\n", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		logrus.Info("Error in resolving v%", err.Error())
		os.Exit(2)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logrus.Info("Error in connection v%", err.Error())
		os.Exit(2)
	}
	_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	if err != nil {
		logrus.Info("Error in connection write v%", err.Error())
		os.Exit(2)
	}
	result, err := ioutil.ReadAll(conn)
	if err != nil {
		logrus.Info("Error in reading v%", err.Error())
		os.Exit(2)
	}
	fmt.Println(string(result))
	os.Exit(0)
}
