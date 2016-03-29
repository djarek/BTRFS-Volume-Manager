package router

import (
	"log"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/wsprotocol"
)

//Context stores session context
type Context struct {
	Sender wsprotocol.AsyncSenderCloser
}

//HandlerFunc represents a function that handles an incoming Message
type HandlerFunc func(*Context, dtos.WebSocketMessage)

type handlerMap map[dtos.WebSocketMessageType]HandlerFunc

//HandlerAdder registers a new Message handler
type HandlerAdder interface {
	AddHandler(dtos.WebSocketMessageType, HandlerFunc)
}

//HandlerExporter adds handler functions to the router.
type HandlerExporter interface {
	ExportHandlers(HandlerAdder)
}

//Router passes the received Messages to registered HandlerFuncs.
type Router struct {
	handlers handlerMap
}

//New constructs a new valid Router
func New() *Router {
	return &Router{make(handlerMap)}
}

//OnNewConnection starts the parsing loop for this connection
func (r *Router) OnNewConnection(c wsprotocol.AsyncSenderCloser, recv <-chan dtos.WebSocketMessage) {
	go r.parsingLoop(&Context{c}, recv)
}

func (r *Router) parsingLoop(ctx *Context, recvChannel <-chan dtos.WebSocketMessage) {
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
