package wsprotocol

import (
	"log"
	"net/http"
	"sync"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

//ConnectionID represents the unique ID of a Connection
type ConnectionID uint64

//ConnectionManager is a thread safe websocket connection & session manager
type ConnectionManager struct {
	connections map[ConnectionID]*Connection
	mtx         sync.RWMutex
	nextCID     ConnectionID

	marshaller dtos.WebSocketMessageMarshaller
	parser     RecvMessageParser
}

//NewConnectionManager creates a valid new instance of a ConnectionManager
func NewConnectionManager(marshaller dtos.WebSocketMessageMarshaller, parser RecvMessageParser) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[ConnectionID]*Connection),
		marshaller:  marshaller,
		parser:      parser}
}

func (cm *ConnectionManager) registerConnection(connection *Connection) ConnectionID {
	cm.mtx.Lock()
	defer cm.mtx.Unlock()

	CID := cm.nextCID
	cm.connections[CID] = connection
	cm.nextCID++
	return CID
}

func (cm *ConnectionManager) unregisterConnection(CID ConnectionID) {
	cm.mtx.Lock()
	defer cm.mtx.Unlock()

	delete(cm.connections, CID)
}

//HandleWSConnection handles the upgrade to a websocket connection and performs
//authentication using the wsserver.WebSocketAuthenticator interface.
//Satisfies the http.HandlerFunc interface.
func (cm *ConnectionManager) HandleWSConnection(w http.ResponseWriter, r *http.Request) {
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error when upgrading http connection to websocket protocol: " + err.Error())
		return
	}

	connection := newConnection(wsConnection, cm.marshaller, cm.parser)
	CID := cm.registerConnection(connection)
	connection.onCloseCallback = func() { cm.unregisterConnection(CID) }
	connection.serve()
}
