package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strings"
)

type APNSNotificaton struct {
	command    uint8
	id         uint32
	expiry     uint32
	tokenLen   uint16
	devToken   []byte
	payloadLen uint16
	payload    []byte
}

func (self *APNSNotificaton) String() string {
	token := hex.EncodeToString(self.devToken)
	token = strings.ToLower(token)
	return fmt.Sprintf("command=%v; id=%v; expiry=%v; token=%v; payload=%v",
		self.command, self.id, self.expiry, token, string(self.payload))
}

type APNSResponse struct {
	id     uint32
	status uint8
}

type APNSConn struct {
	conn net.Conn
}

func (self *APNSConn) Close() error {
	return self.conn.Close()
}

func (self *APNSConn) ReadNotification() (notif *APNSNotificaton, err error) {
	notif = new(APNSNotificaton)
	conn := self.conn
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

	// devtoken
	notif.devToken = make([]byte, notif.tokenLen)
	n, err := io.ReadFull(self.conn, notif.devToken)
	if err != nil {
		notif = nil
		return
	}
	if n != int(notif.tokenLen) {
		notif = nil
		err = fmt.Errorf("no enough data")
		return
	}

	// payload size
	err = binary.Read(conn, binary.BigEndian, &(notif.payloadLen))
	if err != nil {
		notif = nil
		return
	}

	// payload
	notif.payload = make([]byte, notif.payloadLen)
	n, err = io.ReadFull(conn, notif.payload)
	if err != nil {
		notif = nil
		return
	}
	if n != int(notif.payloadLen) {
		notif = nil
		err = fmt.Errorf("no enough data: payload")
		return
	}
	return
}

func (self *APNSConn) Reply(status *APNSResponse) error {
	var command uint8
	command = 8
	err := binary.Write(self.conn, binary.BigEndian, command)
	if err != nil {
		return err
	}
	err = binary.Write(self.conn, binary.BigEndian, status.status)
	if err != nil {
		return err
	}
	err = binary.Write(self.conn, binary.BigEndian, status.id)
	if err != nil {
		return err
	}
	return nil
}
