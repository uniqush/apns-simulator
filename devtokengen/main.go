package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"text/template"
)

const (
	defaultTokenLen int = 32
)

type DevtokenGenerator struct {
	TokenLen  int
	nextToken *big.Int
}

func expandBytes(b []byte, minLen int) []byte {
	if len(b) >= minLen {
		return b
	}

	c := make([]byte, minLen)
	copy(c[minLen-len(b):], b)
	return c
}

func (self *DevtokenGenerator) GenToken() []byte {
	if self.TokenLen <= 0 {
		self.TokenLen = defaultTokenLen
	}
	if self.nextToken == nil {
		self.nextToken = big.NewInt(rand.Int63())
	}
	self.nextToken.Add(self.nextToken, big.NewInt(1))
	return expandBytes(self.nextToken.Bytes(), self.TokenLen)
}

func devTokenToString(token []byte) string {
	str := hex.EncodeToString(token)
	return str
}

var flagNrTokens = flag.Int("n", 1, "number of tokens")
var flagTemplate = flag.String("template", "curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.{{.Id}} -d pushservicetype=apns -d devtoken={{.Token}}", "template of the output")

type DeviceToken struct {
	Token string
	Id    int
}

func main() {
	flag.Parse()

	tmpl, err := template.New("tmpl").Parse(*flagTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Template error: %v\n", err)
		return
	}

	g := new(DevtokenGenerator)
	out := os.Stdout
	for i := 0; i < *flagNrTokens; i++ {
		t := devTokenToString(g.GenToken())
		token := &DeviceToken{t, i}
		tmpl.Execute(out, token)
		fmt.Fprintln(out)
	}
}
