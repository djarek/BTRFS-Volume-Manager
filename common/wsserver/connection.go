package wsserver

import (
	"time"

	"github.com/BTRFS-Volume-Manager/common/dtos"

	"log"

	"github.com/gorilla/websocket"
)

const (
	pongTimeout               = 60 * time.Second
	pingInterval              = (pongTimeout * 9) / 10
	writeTimeout              = 10 * time.Second
	authenticationReadTimeout = writeTimeout
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

//SendCallback is the signature of the function that is called when an async
//write completes
type SendCallback func(error)

//RecvMessageParser specifies the type that is used to resolve received messages
//into appropriate callbacks
type RecvMessageParser interface {
	ParseRecvMsg(*dtos.WebSocketMessage) error
}

type sendTask struct {
	callback SendCallback
	payload  []byte
}

//Connection wraps the websocket connection
type Connection struct {
	wsConnection  *websocket.Conn
	authenticated bool
	marshaller    dtos.WebSocketMessageMarshaller
	parser        RecvMessageParser

	writeChannel chan sendTask
}

//NewConnection constructs a valid empty Connection object
func NewConnection(wsConnection *websocket.Conn, marshaller dtos.WebSocketMessageMarshaller, parser RecvMessageParser) *Connection {
	return &Connection{
		wsConnection:  wsConnection,
		authenticated: false,
		marshaller:    marshaller,
		parser:        parser,

		writeChannel: make(chan sendTask)}

}

//Authenticate performs authentication of the connection using the supplied authenticator.
//It returns nil when the authentication is successful and the connection is ready to be
//served.
func (c *Connection) Authenticate(authenticator WebSocketAuthenticator) (err error) {
	webSocketMsg := authenticator.GetChallenge(c.wsConnection.RemoteAddr())
	payload, err := c.marshaller.Marshall(webSocketMsg)

	if err != nil {
		log.Println("Error when marshalling authentication challenge: " + err.Error())
		return
	}

	err = c.internalWrite(websocket.TextMessage, payload)
	if err != nil {
		log.Println("Error when sending authentication challenge: " + err.Error())
		return
	}

	c.wsConnection.SetReadDeadline(time.Now().Add(authenticationReadTimeout))
	_, payload, err = c.wsConnection.ReadMessage()
	if err != nil {
		log.Println("Error when reading authentication challenge response: " + err.Error())
		return
	}

	webSocketMsg, err = c.marshaller.Unmarshall(payload)
	if err != nil {
		log.Println("Error when unmarshalling authentication challenge response: " + err.Error())
		return
	}

	err = authenticator.VerifyChallengeResponse(c.wsConnection.RemoteAddr(), webSocketMsg)
	if err != nil {
		log.Println("Error when verifying challenger response: " + err.Error())
	}
	c.authenticated = true
	return
}

//Serve launches the reading and writing loops for this websocket connection.
//It must be called only after authentication is successful.
//Blocks until the reader loop exits, at which point the connection is properly
//closed.
func (c *Connection) Serve() {
	if !c.authenticated {
		log.Panicln("Serve called without valid authentication")
	}
	go c.writerLoop()
	c.readerLoop()
}

func (c *Connection) writerLoop() {
	pingTicker := time.NewTicker(pingInterval)
	defer c.Close()

	for {
		select {
		case <-pingTicker.C:
			err := c.internalWrite(websocket.PingMessage, []byte{})
			if err != nil {
				log.Println("Error when sending websocket ping: " + err.Error())
				return
			}

		case task := <-c.writeChannel:
			err := c.internalWrite(websocket.TextMessage, task.payload)
			task.callback(err)
		}
	}
}

func (c *Connection) internalWrite(msgType int, payload []byte) (err error) {
	err = c.wsConnection.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		return
	}
	err = c.wsConnection.WriteMessage(msgType, payload)
	return
}

func (c *Connection) readerLoop() {
	defer c.Close()
	c.wsConnection.SetPongHandler(func(string) error {
		c.wsConnection.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	for {
		_, msgBytes, err := c.wsConnection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Println("Error when reading websocket message: " + err.Error())
			}
			return
		}

		websocketMsg, err := c.marshaller.Unmarshall(msgBytes)
		if err != nil {
			log.Println("Error when unmarshalling WebSocketMessage (received malformed data?): " + err.Error())
			return
		}
		c.parser.ParseRecvMsg(websocketMsg)
	}
}

//Close closes the underlying connection and does the necessary cleanup
//like removing the connection from the connection manager
func (c *Connection) Close() error {
	if c.wsConnection != nil {
		return c.wsConnection.Close()
	}
	return nil
}

//SendAsync sends a WebSocketMessage asynchronously and returns immediately if an
//error is encountered or the message write has been enqueued. If an error is
//encountered during the network transfer, the error is passed to the callback
func (c *Connection) SendAsync(msg *dtos.WebSocketMessage, callback SendCallback) error {
	websocketMsgBytes, err := c.marshaller.Marshall(msg)
	if err != nil {
		log.Println("Error when marshalling WebSocketMessage: " + err.Error())
		return err
	}

	c.writeChannel <- sendTask{callback, websocketMsgBytes}
	return nil
}
