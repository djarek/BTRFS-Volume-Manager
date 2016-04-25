package request

import (
	"sync"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
)

/*AsyncSender allows asynchronous sending of messages*/
type AsyncSender interface {
	SendAsync(msg dtos.WebSocketMessage) <-chan error
}

/*AsyncSenderCloser represents a type that allows asynchronous sending of
messages and closing the connection.*/
type AsyncSenderCloser interface {
	AsyncSender
	Close()
}

type dataMap map[string]interface{}

//Context stores session context
type Context struct {
	AsyncSenderCloser

	data    dataMap
	dataMtx sync.RWMutex

	requestsMtx   sync.Mutex
	requests      map[int64]chan<- dtos.WebSocketMessage
	nextRequestID int64
}

//GetSessionData retrieves a value from this session context
func (c *Context) GetSessionData(key string) (data interface{}, ok bool) {
	c.dataMtx.RLock()
	defer c.dataMtx.RUnlock()

	data, ok = c.data[key]
	return
}

//SetSessionData stores a value in this session context
func (c *Context) SetSessionData(key string, data interface{}) {
	c.dataMtx.Lock()
	defer c.dataMtx.Unlock()

	c.data[key] = data
}

/*NewRequest registers a new request to be sent. The returned channel is used to
receive the incoming response. The ID returned from this function has to be used
as the value for WebSocketMessage.RequestID. */
func (c *Context) NewRequest() (int64, <-chan dtos.WebSocketMessage) {
	responseChannel := make(chan dtos.WebSocketMessage, 1)

	c.requestsMtx.Lock()
	defer c.requestsMtx.Unlock()
	requestID := c.nextRequestID
	c.nextRequestID++
	c.requests[requestID] = responseChannel
	return requestID, responseChannel
}

/*GetRequest retrieves the channel associated with the request identified by the
provided requestID. If the request has not been found or has already been retrieved
the ok return value is set to false.*/
func (c *Context) GetRequest(requestID int64) (channel chan<- dtos.WebSocketMessage, ok bool) {
	c.requestsMtx.Lock()
	defer c.requestsMtx.Unlock()
	channel, ok = c.requests[requestID]
	delete(c.requests, requestID)
	return
}

/*OnClose is called when the underlying connection is closed. It performs the
necessary context cleanup.*/
func (c *Context) OnClose() {
	c.requestsMtx.Lock()
	defer c.requestsMtx.Unlock()

	for _, responseChannel := range c.requests {
		close(responseChannel)
	}
	c.requests = nil
}

//NewContext constructs a new valid Context object
func NewContext(c AsyncSenderCloser) *Context {
	return &Context{
		AsyncSenderCloser: c,
		data:              make(dataMap),
		requests:          make(map[int64]chan<- dtos.WebSocketMessage),
	}
}
