package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	weakrand "math/rand"
	"net"
	"time"
)

func handleClient(conn net.Conn) {
	defer conn.Close()
	for {
		var status uint8
		notif, err := readNotification(conn)
		if err != nil {
			fmt.Printf("Got an error: %v\n", err)
			break
		}
		// we don't need a good random generator. Even with mod is ok.
		s := weakrand.Int() % 100

		if s > 50 {
			s -= 50
			s /= 5
		} else {
			s = 0
		}
		status = uint8(s)
		if status == uint8(0) {
			if weakrand.Int()%5 != 0 {
				fmt.Printf("[%v] Got a notifcation: %v. It is successfully processed but will not be replied. I am a bad fruit\n", time.Now(), notif)
				continue
			}
		} else if status > uint8(8) {
			status = 255
			fmt.Printf("[%v] Drop this connection\n", time.Now())
			return
		}
		fmt.Printf("[%v] Got a notifcation: %v. and the status is: %v\n", time.Now(), notif, status)
		replyNotification(conn, status, notif.id)
	}
}

func main() {
	keyFile := "key.pem"
	certFile := "cert.pem"
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Printf("Load Key Error: %v\n", err)
		return
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		Rand:               rand.Reader,
	}
	conn, err := tls.Listen("tcp", "0.0.0.0:8080", config)
	if err != nil {
		fmt.Printf("Listen: %v\n", err)
		return
	}
	for {
		client, err := conn.Accept()
		if err != nil {
			fmt.Printf("Accept Error: %v\n", err)
		}
		fmt.Printf("[%v] Received connection from %v\n", time.Now(), client.RemoteAddr())
		go handleClient(client)
	}
}
