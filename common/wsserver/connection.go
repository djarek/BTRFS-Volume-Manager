package wsserver

import (
	"log"
	"time"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/gorilla/websocket"
)

const (
	pongTimeout               = 6 * time.Second
	pingInterval              = (pongTimeout * 9) / 10
	writeTimeout              = 10 * time.Second
	authenticationReadTimeout = writeTimeout
)

//SendCallback is the signature of the function that is called when an async
//write completes
type SendCallback func(error)

//RecvMessageParser specifies the type that is used to resolve received messages
//into appropriate callbacks
type RecvMessageParser interface {
	ParseRecvMsg(dtos.WebSocketMessage) error
}

type sendTask struct {
	callback SendCallback
	payload  []byte
}

//Connection wraps the websocket connection
type Connection struct {
	wsConnection    *websocket.Conn
	authenticated   bool
	marshaller      dtos.WebSocketMessageMarshaller
	parser          RecvMessageParser
	onCloseCallback func()
	pingTicker      *time.Ticker

	writeChannel chan sendTask
}

func newConnection(wsConnection *websocket.Conn, marshaller dtos.WebSocketMessageMarshaller, parser RecvMessageParser) *Connection {
	return &Connection{
		wsConnection:  wsConnection,
		authenticated: false,
		marshaller:    marshaller,
		parser:        parser,

		writeChannel: make(chan sendTask)}
}

func (c *Connection) registerOnCloseCallback(cb func()) {
	c.onCloseCallback = cb
}

//Authenticate performs authentication of the connection using the supplied authenticator.
//It returns nil when the authentication is successful and the connection is ready to be
//served.
func (c *Connection) authenticate(authenticator WebSocketAuthenticator) (err error) {
	c.wsConnection.SetReadDeadline(time.Now().Add(authenticationReadTimeout))
	_, payload, err := c.wsConnection.ReadMessage()
	if err != nil {
		log.Println("Error when reading authentication message: " + err.Error())
		return
	}

	response, err := authenticator.Authenticate(c.wsConnection.RemoteAddr(), payload)
	c.internalWrite(websocket.TextMessage, response)
	if err != nil {
		log.Println("Error when authenticating: " + err.Error())
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
	c.pingTicker = time.NewTicker(pingInterval)
	defer c.Close()

	for {
		select {
		case <-c.pingTicker.C:
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
	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}
	if c.onCloseCallback != nil {
		c.onCloseCallback()
	}
	return c.wsConnection.Close()
}

//SendAsync sends a WebSocketMessage asynchronously and returns immediately if an
//error is encountered or the message write has been enqueued. If an error is
//encountered during the network transfer, the error is passed to the callback
func (c *Connection) SendAsync(msg dtos.WebSocketMessage, callback SendCallback) error {
	webSocketMsgBytes, err := c.marshaller.Marshall(msg)
	if err != nil {
		log.Println("Error when marshalling WebSocketMessage: " + err.Error())
		return err
	}

	c.writeChannel <- sendTask{callback, webSocketMsgBytes}
	return nil
}
