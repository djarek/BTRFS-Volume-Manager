package wsprotocol

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/gorilla/websocket"
)

const (
	pongTimeout               = 6 * time.Second
	pingInterval              = (pongTimeout * 9) / 10
	writeTimeout              = 10 * time.Second
	authenticationReadTimeout = writeTimeout
	writeChannelSize          = 16
)

var (
	/*ErrSessionAlreadyAssigned indicates that the Connection already has
	  a session assigned (authentication has already been performed)*/
	ErrSessionAlreadyAssigned = errors.New("SessionContext already assigned.")
)

//RecvMessageParser specifies the type that is used to resolve received messages
//into appropriate callbacks
type RecvMessageParser interface {
	ParseRecvMsg(dtos.WebSocketMessage, *Connection) error
}

type outputMessage struct {
	channel     chan<- error
	payload     []byte
	messageType int
}

//SessionContext represents the current user session
type SessionContext interface {
	ReleaseConnection(*Connection)
}

//Connection wraps the websocket connection
type Connection struct {
	wsConnection    *websocket.Conn
	marshaller      dtos.WebSocketMessageMarshaller
	parser          RecvMessageParser
	onCloseCallback func()

	sessionMtx sync.Mutex
	session    SessionContext

	writeChannel chan outputMessage
	closeOnce    sync.Once
}

func newConnection(
	wsConnection *websocket.Conn,
	marshaller dtos.WebSocketMessageMarshaller,
	parser RecvMessageParser,
) *Connection {

	return &Connection{
		wsConnection: wsConnection,
		marshaller:   marshaller,
		parser:       parser,
		session:      nil,

		writeChannel: make(chan outputMessage, writeChannelSize)}
}

//serve launches the reading and writing loops for this websocket connection.
//It must be called only after authentication is successful.
//Blocks until the reader loop exits, at which point the connection is properly
//closed.
func (c *Connection) serve() {
	go c.writerLoop()
	c.readerLoop()
}

func (c *Connection) readerLoop() {
	defer c.Close()
	c.wsConnection.SetPongHandler(func(string) error {
		c.wsConnection.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})
	c.wsConnection.SetPingHandler(func(string) error {
		c.enqueueOutputMessage(outputMessage{
			channel:     nil,
			payload:     []byte{},
			messageType: websocket.PongMessage,
		})
		return nil
	})

	for {
		msgType, msgBytes, err := c.wsConnection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway) {
				log.Println("Error when reading websocket message: " + err.Error())
			}
			return
		}
		if msgType != websocket.TextMessage || len(msgBytes) == 0 {
			return
		}

		websocketMsg, err := c.marshaller.Unmarshall(msgBytes)
		if err != nil {
			log.Println("Error when unmarshalling WebSocketMessage (received malformed data?): " + err.Error())
			log.Println(msgBytes)
			continue
		}
		c.parser.ParseRecvMsg(websocketMsg, c)
	}
}

//GetSession returns the current session in a thread-safe way
func (c *Connection) GetSession() SessionContext {
	c.sessionMtx.Lock()
	c.sessionMtx.Unlock()

	return c.session
}

/*SetSession sets the current session in a thread-safe way. If there is
a session already set, the function will return an error.*/
func (c *Connection) SetSession(s SessionContext) error {
	c.sessionMtx.Lock()
	defer c.sessionMtx.Unlock()

	if c.session != nil {
		return ErrSessionAlreadyAssigned
	}

	c.session = s
	return nil
}

/*Close attempts to send a proper close to the client. If the connection
is in an invalid state, this will fail, however, all the necessary cleanup
will be performed properly anyway.*/
func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		c.enqueueOutputMessage(outputMessage{
			channel:     nil,
			payload:     websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			messageType: websocket.CloseMessage,
		})
	})
}

func (c *Connection) enqueueOutputMessage(msg outputMessage) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("Connection closed")
		}
	}()
	c.writeChannel <- msg
	return
}

/*SendAsync sends a WebSocketMessage asynchronously and returns immediately if an
error is encountered or the message write has been enqueued. If an error is
encountered during the network transfer, the error is passed through the
returned channel. If there is no error, nil is sent on that channel*/
func (c *Connection) SendAsync(msg dtos.WebSocketMessage) (<-chan error, error) {
	msgBytes, err := c.marshaller.Marshall(msg)
	if err != nil {
		log.Println("Error when marshalling WebSocketMessage: " + err.Error())
		return nil, err
	}
	channel := make(chan error, 1)
	return channel, c.enqueueOutputMessage(outputMessage{
		channel:     channel,
		payload:     msgBytes,
		messageType: websocket.TextMessage,
	})
}

func (c *Connection) writerLoop() {
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()
	defer c.internalClose()

	for {
		select {
		case <-pingTicker.C:
			err := c.internalWrite(websocket.PingMessage, []byte{})
			if err != nil {
				log.Println("Error when sending websocket ping: " + err.Error())
				return
			}

		case msg := <-c.writeChannel:
			err := c.sendOutputMessage(msg)
			if err != nil || msg.messageType == websocket.CloseMessage {
				return
			}
		}
	}
}

/*!!!!!! !!!!!!  !!!!!! DANGER !!!!!! !!!!!! !!!!!!
The methods below are only supposed to be called by the writer goroutine
*/

func (c *Connection) flushRemainingTasks() {
	for task := range c.writeChannel {
		/*We ignore the error because we want to make sure the task owners are
		notified about completion.*/
		_ = c.sendOutputMessage(task)
	}
}

func (c *Connection) internalClose() {
	session := c.GetSession()
	if session != nil {
		session.ReleaseConnection(c)
	}

	close(c.writeChannel)
	c.flushRemainingTasks()

	if c.onCloseCallback != nil {
		c.onCloseCallback()
	}
	c.wsConnection.Close()
}

func (c *Connection) sendOutputMessage(msg outputMessage) error {
	err := c.internalWrite(msg.messageType, msg.payload)
	if msg.channel != nil {
		msg.channel <- err
	}
	return err
}

func (c *Connection) internalWrite(msgType int, payload []byte) error {
	deadline := time.Now().Add(writeTimeout)
	err := c.wsConnection.SetWriteDeadline(deadline)
	if err != nil {
		return err
	}
	return c.wsConnection.WriteMessage(msgType, payload)
}
