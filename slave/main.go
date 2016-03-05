package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/djarek/BTRFS-Volume-Manager/common/dtos"
	"github.com/djarek/BTRFS-Volume-Manager/common/wsserver"
)

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

func main() {
	marshaller := &testMarshaller{}
	parser := &testParser{}
	authenticator := &testAuthenticator{}
	cm := wsserver.NewConnectionManager(marshaller, parser, authenticator)
	http.HandleFunc("/ws", cm.HandleWSConnection)
	err := http.ListenAndServe("127.0.0.1:8080", nil)
	panic(err)
}
