package wsprotocol

import (
	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/gorilla/websocket"
)

//DefaultDialer uses the default marshaller
var DefaultDialer = Dialer{dtos.JSONMessageMarshaller{}}

//Dialer allows establishing and configuring a websocket connection
type Dialer struct {
	m dtos.WebSocketMessageMarshaller
}

//Dial connects to a websocket endpoint and creates a Connection
func (d Dialer) Dial(url string, p RecvMessageParser) (*Connection, error) {
	wsConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	conn := newConnection(wsConn, d.m, p)
	go conn.serve()
	return conn, nil
}
