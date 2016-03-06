package dtos

//WebSocketMessageType represents the type of the message.
type WebSocketMessageType uint8

const (
	//WSMsgError indicates the WebSocketMessage is an error response
	WSMsgError WebSocketMessageType = iota
	//WSMsgRequest indicates the WebSocketMessage is a request.
	//The marshalled request body is in the Payload member.
	WSMsgRequest
	//WSMsgResponse indicates the WebSocketMessage is a response.
	//The marshalled response body is in the Payload member.
	WSMsgResponse
)

//WebSocketMessage represents a message received from a client or
//ready to be sent to it
type WebSocketMessage struct {
	Type    WebSocketMessageType
	Payload []byte
}

//WebSocketMessageMarshaller allows conversion from byte slices to WSMessage structs
//and vice versa.
type WebSocketMessageMarshaller interface {
	Marshall(*WebSocketMessage) ([]byte, error)
	Unmarshall([]byte) (*WebSocketMessage, error)
}
