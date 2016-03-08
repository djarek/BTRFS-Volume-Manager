package main

/*
#cgo LDFLAGS: -L../libbtrfs/build -lbtrfs -lblkid
*/
import "C"

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/wsserver"
	"github.com/djarek/btrfs-volume-manager/slave/osinterface"
)

var exitChan chan bool

type testMarshaller struct {
}

func (tm testMarshaller) Marshall(wsMsg *dtos.WebSocketMessage) ([]byte, error) {
	return []byte("test"), nil
}

func (tm testMarshaller) Unmarshall(buffer []byte) (*dtos.WebSocketMessage, error) {
	fmt.Println(buffer)
	return &dtos.WebSocketMessage{}, nil
}

type testAuthenticator struct{}

func (ta testAuthenticator) GetChallenge(net.Addr) *dtos.WebSocketMessage {
	return nil
}

func (ta testAuthenticator) VerifyChallengeResponse(net.Addr, *dtos.WebSocketMessage) error {
	return nil
}

type testParser struct{}

func (tp testParser) ParseRecvMsg(*dtos.WebSocketMessage) error {
	return nil
}

func serve() {
	log.Fatalln(http.ListenAndServe("127.0.0.1:8080", nil))
}

func main() {
	marshaller := &testMarshaller{}
	parser := &testParser{}
	authenticator := &testAuthenticator{}
	cm := wsserver.NewConnectionManager(marshaller, parser, authenticator)
	http.HandleFunc("/ws", cm.HandleWSConnection)
	devs := osinterface.ProbeBlockDevices()
	for _, dev := range devs {
		fmt.Println(dev)
	}
	go serve()
	<-exitChan
}
