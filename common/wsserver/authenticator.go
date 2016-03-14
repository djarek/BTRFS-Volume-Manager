package wsserver

import "net"

//WebSocketAuthenticator represents an object used for authentication of a newly
//connected websocket client.
type WebSocketAuthenticator interface {
	Authenticate(net.Addr, []byte) ([]byte, error)
}
