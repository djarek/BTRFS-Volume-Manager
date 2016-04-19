package wsprotocol

import (
	"log"
	"net/http"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

/*Router handles incoming Messages*/
type Router interface {
	OnNewConnection(request.AsyncSenderCloser, <-chan dtos.WebSocketMessage) *request.Context
}

/*ConnectionUpgrader upgrades an incomming HTTP connection to a websocket connection
and constructs necessary structures*/
type ConnectionUpgrader struct {
	router     Router
	marshaller dtos.WebSocketMessageMarshaller
}

//NewConnectionUpgrader creates a valid new instance of a ConnectionManager
func NewConnectionUpgrader(m dtos.WebSocketMessageMarshaller, r Router) *ConnectionUpgrader {
	return &ConnectionUpgrader{
		marshaller: m,
		router:     r,
	}
}

//HandleWSConnection handles the upgrade to a websocket connection and performs
//authentication using the wsserver.WebSocketAuthenticator interface.
//Satisfies the http.HandlerFunc interface.
func (c *ConnectionUpgrader) HandleWSConnection(w http.ResponseWriter, r *http.Request) {
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error when upgrading http connection to websocket protocol: " + err.Error())
		return
	}

	connection, recvChannel := newConnection(wsConnection, c.marshaller)
	c.router.OnNewConnection(connection, recvChannel)
	connection.serve()
}
