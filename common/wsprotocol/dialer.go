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
func (d Dialer) Dial(url string, r Router) (*Connection, error) {
	wsConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	conn, recvChannel := newConnection(wsConn, d.m)
	r.OnNewConnection(conn, recvChannel)
	go conn.serve()
	return conn, nil
}
