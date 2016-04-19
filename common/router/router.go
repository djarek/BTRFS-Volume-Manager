package router

import (
	"log"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
)

//HandlerFunc represents a function that handles an incoming Message
type HandlerFunc func(*request.Context, dtos.WebSocketMessage)

type handlerMap map[dtos.WebSocketMessageType]HandlerFunc

//HandlerAdder registers a new Message handler
type HandlerAdder interface {
	AddHandler(dtos.WebSocketMessageType, HandlerFunc)
	AddOnCloseHandler(HandlerFunc)
}

//HandlerExporter adds handler functions to the router.
type HandlerExporter interface {
	ExportHandlers(HandlerAdder)
}

//Router passes the received Messages to registered HandlerFuncs.
type Router struct {
	handlers        handlerMap
	onCloseHandlers []HandlerFunc
}

//New constructs a new valid Router
func New() *Router {
	return &Router{handlers: make(handlerMap)}
}

//OnNewConnection starts the parsing loop for this connection
func (r *Router) OnNewConnection(c request.AsyncSenderCloser, recv <-chan dtos.WebSocketMessage) *request.Context {
	ctx := request.NewContext(c)
	go r.parsingLoop(ctx, recv)
	return ctx
}

func (r *Router) parsingLoop(ctx *request.Context, recvChannel <-chan dtos.WebSocketMessage) {
	for msg := range recvChannel {
		h, found := r.handlers[msg.MessageType]
		if !found {
			log.Printf("Unknown message type: %d\n", msg.MessageType)
			//TODO: Send error
			continue
		} else {
			h(ctx, msg)
		}
	}

	for _, h := range r.onCloseHandlers {
		h(ctx, dtos.WebSocketMessage{})
	}
	ctx.OnClose()
}

/*AddHandler registers a handler function for a particular message type.
If the type has already been registered, the function panics.*/
func (r *Router) AddHandler(t dtos.WebSocketMessageType, h HandlerFunc) {
	_, found := r.handlers[t]
	if found {
		log.Panicf("Handler already present(type: %d)\n", t)
	}
	r.handlers[t] = h
}

/*AddOnCloseHandler adds a new on connection close handler */
func (r *Router) AddOnCloseHandler(h HandlerFunc) {
	r.onCloseHandlers = append(r.onCloseHandlers, h)
}

/*DefaultResponseHandler performs the default action when a response is received -
try to find the channel associated with the requestID and send the message to the
handler there*/
func DefaultResponseHandler(ctx *request.Context, msg dtos.WebSocketMessage) {
	responseChannel, found := ctx.GetRequest(msg.RequestID)
	if found {
		responseChannel <- msg
	}
}
