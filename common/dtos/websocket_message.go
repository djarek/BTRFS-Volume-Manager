package dtos

import "encoding/json"

//WebSocketMessageType represents the type of the message.
type WebSocketMessageType int32

const (
	//WSMsgError indicates the WebSocketMessage is an error response
	WSMsgError WebSocketMessageType = iota
	//WSMsgRequestRegisterSlave indicates this is a request for the master to register a new slave
	WSMsgRequestRegisterSlave
	//WSMsgResponseRegisterSlave indicates this is a response to a previous request to register a new slave
	WSMsgResponseRegisterSlave
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
	err = json.Unmarshal(buf, &msg)
	return
}
