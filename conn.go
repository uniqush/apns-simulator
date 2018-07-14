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
	priority   uint8
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

func (self *APNSConn) ReadNotification() (*APNSNotificaton, error) {
	notif := new(APNSNotificaton)
	conn := self.conn
	err := binary.Read(conn, binary.BigEndian, &(notif.command))
	if err != nil {
		return nil, err
	}

	switch notif.command {
	case 0:
		return self.processLegacyV0Notification(notif)
	case 1:
		return self.processLegacyV1Notification(notif)
	case 2:
		return self.processBinaryProviderAPINotification(notif)
	default:
		return nil, fmt.Errorf("Unknown Command")
	}
}

// processLegacyNotification processes the rest of the command 1 "Legacy Notification Format" from Apple.
func (self *APNSConn) processLegacyV1Notification(notif *APNSNotificaton) (*APNSNotificaton, error) {
	conn := self.conn
	err := binary.Read(conn, binary.BigEndian, &(notif.id))
	if err != nil {
		return nil, err
	}
	err = binary.Read(conn, binary.BigEndian, &(notif.expiry))
	if err != nil {
		return nil, err
	}
	// Coincidentally, the remainder of v0 is the same as the remainder of v1
	return self.processLegacyV0Notification(notif)
}

func (self *APNSConn) processLegacyV0Notification(notif *APNSNotificaton) (*APNSNotificaton, error) {
	conn := self.conn
	err := binary.Read(conn, binary.BigEndian, &(notif.tokenLen))
	if err != nil {
		return nil, err
	}

	// devtoken
	notif.devToken, err = self.readDeviceToken(notif.tokenLen)
	if err != nil {
		return nil, err
	}

	// payload size
	err = binary.Read(conn, binary.BigEndian, &(notif.payloadLen))
	if err != nil {
		return nil, err
	}

	// payload
	notif.payload, err = self.readPayload(notif.payloadLen)
	if err != nil {
		return nil, err
	}
	return notif, nil
}

func (self *APNSConn) readPayload(length uint16) ([]byte, error) {
	return self.readByteSlice(length, "payload")
}

func (self *APNSConn) readDeviceToken(length uint16) ([]byte, error) {
	return self.readByteSlice(length, "deviceToken")
}

func (self *APNSConn) readByteSlice(length uint16, description string) ([]byte, error) {
	payload := make([]byte, length)
	n, err := io.ReadFull(self.conn, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", description, err)
	}
	if n != int(length) {
		return nil, fmt.Errorf("no enough data: %s", description)
	}
	return payload, nil
}

const (
	ITEM_DEVICE_TOKEN            = 1
	ITEM_PAYLOAD                 = 2
	ITEM_NOTIFICATION_IDENTIFIER = 3
	ITEM_EXPIRATION_DATE         = 4
	ITEM_PRIORITY                = 5
)

// processBinaryProviderAPINotification processes the rest of the command 1 "Legacy Notification Format" from Apple.
func (self *APNSConn) processBinaryProviderAPINotification(notif *APNSNotificaton) (*APNSNotificaton, error) {
	conn := self.conn
	var frameLength uint32
	err := binary.Read(conn, binary.BigEndian, &frameLength)
	if err != nil {
		return nil, err
	}

	var didRead map[uint8]bool

	bytesRead := uint32(0)
	for bytesRead < frameLength {
		var itemID uint8
		var itemDataLength uint16
		err = binary.Read(conn, binary.BigEndian, &itemID)
		if err != nil {
			return nil, err
		}
		if _, exists := didRead[itemID]; exists {
			return nil, fmt.Errorf("Cannot read %d twice", itemID)
		}
		err = binary.Read(conn, binary.BigEndian, &itemDataLength)
		if err != nil {
			return nil, err
		}
		// This attempts to report an error as early in reading bytes as possible
		switch itemID {
		case ITEM_DEVICE_TOKEN:
			if itemDataLength == 0 {
				return nil, fmt.Errorf("Too short of a token %d==0", itemDataLength)
			}
			if itemDataLength > 100 {
				return nil, fmt.Errorf("Too long of a token %d>100", itemDataLength)
			}
			notif.devToken, err = self.readDeviceToken(itemDataLength)
			if err != nil {
				return nil, err
			}
		case ITEM_PAYLOAD:
			if itemDataLength > 2048 {
				return nil, fmt.Errorf("Too long of a payload %d>2048", itemDataLength)
			}
			notif.payload, err = self.readPayload(itemDataLength)
			if err != nil {
				return nil, err
			}
		case ITEM_NOTIFICATION_IDENTIFIER:
			if itemDataLength != 4 {
				return nil, fmt.Errorf("Invalid notification identifier item length: %d != 4", itemDataLength)
			}
			err = binary.Read(conn, binary.BigEndian, &notif.id)
			if err != nil {
				return nil, err
			}
		case ITEM_EXPIRATION_DATE:
			if itemDataLength != 4 {
				return nil, fmt.Errorf("Invalid expiry date item length: %d != 4", itemDataLength)
			}
			err = binary.Read(conn, binary.BigEndian, &notif.expiry)
			if err != nil {
				return nil, err
			}
		case ITEM_PRIORITY:
			if itemDataLength != 1 {
				return nil, fmt.Errorf("Invalid prority item length: %d != 1", itemDataLength)
			}
			err = binary.Read(conn, binary.BigEndian, &notif.priority)
			if err != nil {
				return nil, err
			}
			if notif.priority != 5 && notif.priority != 10 {
				return nil, fmt.Errorf("Invalid prority value: %d must be 5 or 10", notif.priority)
			}
		}
	}
	if !didRead[ITEM_DEVICE_TOKEN] {
		return nil, fmt.Errorf("Missing device token")
	}
	if !didRead[ITEM_PAYLOAD] {
		return nil, fmt.Errorf("Missing payload")
	}
	return notif, nil
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
