package main

import (
	"encoding/binary"
	"crypto/rand"
	weakrand "math/rand"
	"crypto/tls"
	"fmt"
	"net"
	"io"
	"time"
)

type apnsNotification struct {
	command    uint8
	id         uint32
	expiry     uint32
	tokenLen   uint16
	devToken   []byte
	payloadLen uint16
	payload    []byte
}

func (self *apnsNotification) String() string {
	return fmt.Sprintf("command=%v\nid=%v\nexpiry=%v\ntoken=%v\n,payload=%v\n",
		self.command, self.id, self.expiry, self.devToken, string(self.payload))
}

func readNotification(conn net.Conn) (notif *apnsNotification, err error) {
	notif = new(apnsNotification)
	err = binary.Read(conn, binary.BigEndian, &(notif.command))
	if err != nil {
		notif = nil
		return
	}

	if notif.command == 1 {
		err = binary.Read(conn, binary.BigEndian, &(notif.id))
		if err != nil {
			notif = nil
			return
		}
		err = binary.Read(conn, binary.BigEndian, &(notif.expiry))
		if err != nil {
			notif = nil
			return
		}
	} else if notif.command != 0 {
		notif = nil
		err = fmt.Errorf("Unkown Command")
		return
	}
	err = binary.Read(conn, binary.BigEndian, &(notif.tokenLen))
	if err != nil {
		notif = nil
		return
	}
	if (notif.tokenLen > 512) {
		notif = nil
		err = fmt.Errorf("Token Length is too large (%v bytes)", notif.tokenLen)
		return
	}
	notif.devToken = make([]byte, notif.tokenLen)
	n, err := conn.Read(notif.devToken)
	if err != nil {
		notif = nil
		return
	}
	if n != int(notif.tokenLen) {
		notif = nil
		err = fmt.Errorf("May be OK. XXX read tokenlen = %v", n)
		return
	}
	err = binary.Read(conn, binary.BigEndian, &(notif.payloadLen))
	if err != nil {
		notif = nil
		return
	}
	if (notif.payloadLen > 2048) {
		notif = nil
		err = fmt.Errorf("payload Length is too large (%v bytes)", notif.payloadLen)
		return
	}
	notif.payload = make([]byte, notif.payloadLen)
	n, err = io.ReadFull(conn, notif.payload)
	if err != nil {
		notif = nil
		return
	}
	if n != int(notif.payloadLen) {
		notif = nil
		err = fmt.Errorf("May be OK. XXX read payload len= %v bytes", n)
		return
	}
	return
}

func replyNotification(conn net.Conn, status uint8, id uint32) error {
	var command uint8
	command = 8
	err := binary.Write(conn, binary.BigEndian, command)
	if err != nil {
		return err
	}
	err = binary.Write(conn, binary.BigEndian, status)
	if err != nil {
		return err
	}
	err = binary.Write(conn, binary.BigEndian, id)
	if err != nil {
		return err
	}
	return nil
}

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
			if weakrand.Int() % 5 != 0 {
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
