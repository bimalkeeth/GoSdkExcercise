package main

import (
	"github.com/Sirupsen/logrus"
	"net"
	"os"
	"time"
)

func main() {
	service := ":1200"
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	checkErrorUdp(err)
	conn, err := net.ListenUDP("udp", udpAddr)
	checkErrorUdp(err)
	for {
		handleClientUdp(conn)
	}
}

func handleClientUdp(conn *net.UDPConn) {
	var buff [512]byte
	_, addr, err := conn.ReadFromUDP(buff[0:])
	if err != nil {
		return
	}
	dayTime := time.Now().String()
	_, _ = conn.WriteToUDP([]byte(dayTime), addr)

}

func checkErrorUdp(e error) {
	if e != nil {
		logrus.Info("Fatal Error occurred %v", e.Error())
		os.Exit(1)
	}
}
