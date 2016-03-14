package dtos

import "encoding/json"

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
	Marshall(WebSocketMessage) ([]byte, error)
	Unmarshall([]byte) (WebSocketMessage, error)
}

//JSONMessageMarshaller is the default WebSocketMessageMarshaller - uses JSON as the target format
type JSONMessageMarshaller struct{}

//Marshall encodes a WebSocketMessage as a JSON object
func (j JSONMessageMarshaller) Marshall(msg WebSocketMessage) (buf []byte, err error) {
	buf, err = json.Marshal(msg)
	return
}

//Unmarshall decodes a JSON object to a WebSocketMessage
func (JSONMessageMarshaller) Unmarshall(buf []byte) (msg WebSocketMessage, err error) {
	err = json.Unmarshal(buf, msg)
	return
}
