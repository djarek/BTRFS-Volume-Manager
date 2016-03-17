package wsserver

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

//SendCallback is the signature of the function that is called when an async
//write completes
type SendCallback func(error)

//RecvMessageParser specifies the type that is used to resolve received messages
//into appropriate callbacks
type RecvMessageParser interface {
	ParseRecvMsg(dtos.WebSocketMessage) error
}

type sendTask struct {
	callback    SendCallback
	payload     []byte
	messageType int
}

//Connection wraps the websocket connection
type Connection struct {
	wsConnection    *websocket.Conn
	marshaller      dtos.WebSocketMessageMarshaller
	parser          RecvMessageParser
	onCloseCallback func()

	writeChannel chan sendTask
	closeOnce    sync.Once
}

func newConnection(wsConnection *websocket.Conn, marshaller dtos.WebSocketMessageMarshaller, parser RecvMessageParser) *Connection {
	return &Connection{
		wsConnection: wsConnection,
		marshaller:   marshaller,
		parser:       parser,

		writeChannel: make(chan sendTask, writeChannelSize)}
}

//authenticate performs authentication of the connection using the supplied authenticator.
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
	return
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

	for {
		_, msgBytes, err := c.wsConnection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) &&
				websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
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

/*Close attempts to send a proper close to the client. If the connection
is in an invalid state, this will fail, however, all the necessary cleanup
will be performed properly anyway.*/
func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		c.addSendTask(sendTask{
			callback:    func(error) {},
			payload:     websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			messageType: websocket.CloseMessage,
		})
	})
}

func (c *Connection) addSendTask(task sendTask) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("Connection closed")
			task.callback(err)
		}
	}()
	c.writeChannel <- task
	return
}

//SendAsync sends a WebSocketMessage asynchronously and returns immediately if an
//error is encountered or the message write has been enqueued. If an error is
//encountered during the network transfer, the error is passed to the callback
func (c *Connection) SendAsync(msg dtos.WebSocketMessage, callback SendCallback) error {
	msgBytes, err := c.marshaller.Marshall(msg)
	if err != nil {
		log.Println("Error when marshalling WebSocketMessage: " + err.Error())
		return err
	}

	return c.addSendTask(sendTask{
		callback:    callback,
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

		case task := <-c.writeChannel:
			err := c.sendTask(task)
			if err != nil || task.messageType == websocket.CloseMessage {
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
		_ = c.sendTask(task)
	}
}

func (c *Connection) internalClose() {
	close(c.writeChannel)
	c.flushRemainingTasks()

	if c.onCloseCallback != nil {
		c.onCloseCallback()
	}
	c.wsConnection.Close()
}

func (c *Connection) sendTask(task sendTask) error {
	err := c.internalWrite(websocket.TextMessage, task.payload)
	task.callback(err)
	return err
}

func (c *Connection) internalWrite(msgType int, payload []byte) error {
	err := c.wsConnection.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		return err
	}
	return c.wsConnection.WriteMessage(msgType, payload)
}
