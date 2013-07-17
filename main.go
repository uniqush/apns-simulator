package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/uniqush/log"
	"io"
	weakrand "math/rand"
	"net"
	"os"
	"time"
)

type APNSProcessor struct {
	MinProcessingTime time.Duration
	MaxProcessingTime time.Duration

	Simulator APNSSimulator
	Conn      *APNSConn
	Log       log.Logger

	sendChan chan *APNSResponse
}

func NewAPNSProcessor(conn *APNSConn, logger log.Logger) *APNSProcessor {
	ret := new(APNSProcessor)
	ret.Conn = conn
	ret.Log = logger
	ret.sendChan = make(chan *APNSResponse)
	return ret
}

func (self *APNSProcessor) sendingResponses(start chan bool) {
	<-start
	for res := range self.sendChan {
		err := self.Conn.Reply(res)
		if err != nil {
			return
		}
	}
}

func (self *APNSProcessor) Process() {
	if self.sendChan == nil {
		self.sendChan = make(chan *APNSResponse)
	}
	if self.Conn == nil {
		return
	}
	if self.Simulator == nil {
		self.Simulator = &APNSNormalSimulator{0, 0}
	}

	ch := make(chan bool)
	go self.sendingResponses(ch)
	ch <- true

	defer self.Conn.Close()

	for {
		notif, err := self.Conn.ReadNotification()
		if err != nil {
			self.Log.Errorf("Got an error: %v", err)
			return
		}
		self.Log.Infof("Received notification: %v", notif)
		res, err := self.Simulator.Reply(notif)
		if err != nil {
			self.Log.Errorf("Got an error: %v\n", err)
			return
		}

		waitTime := self.MaxProcessingTime
		if self.MaxProcessingTime != self.MinProcessingTime {
			t := weakrand.Int63n(self.MaxProcessingTime.Nanoseconds() - self.MinProcessingTime.Nanoseconds())
			t += self.MinProcessingTime.Nanoseconds()
			waitTime = time.Duration(t)
		}
		go func(res *APNSResponse, wait time.Duration) {
			time.Sleep(wait)
			self.sendChan <- res
		}(res, waitTime)
	}
}

func handleClient(conn net.Conn, writer io.Writer, level int) {
	logger := log.NewLogger(writer, fmt.Sprintf("[%v]", conn.RemoteAddr().String()), level)
	c := &APNSConn{conn}
	proc := NewAPNSProcessor(c, logger)
	proc.Process()
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
		go handleClient(client, os.Stderr, log.LOGLEVEL_DEBUG)
	}
}
