package dtos

import "encoding/json"

//WebSocketMessageType represents the type of the message.
type WebSocketMessageType int32

const (
	//WSMsgError indicates this WebSocketMessage contains an error object
	WSMsgError WebSocketMessageType = iota
	/*WSMsgAuthenticationRequest indicates this WebSocketMessage contains an
	AuthenticationRequest*/
	WSMsgAuthenticationRequest
)

const (
	/*WSMsgAuthenticationResponse indicates this WebSocketMessage contains an
	AuthenticationResponse*/
	WSMsgAuthenticationResponse WebSocketMessageType = iota + 10001
)

//WebSocketMessage represents a message received from a client or
//ready to be sent to it
type WebSocketMessage struct {
	MessageType WebSocketMessageType `json:"messageType"`
	RequestID   int64                `json:"requestID"`
	Payload     *json.RawMessage     `json:"payload"`
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

//AuthenticationRequest represents a request from the client to perform authentication
type AuthenticationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

/*AuthenticationResponse represents a response to the client indicating whether
authentication succeeded or failed*/
type AuthenticationResponse struct {
	Result string `json:"result"`
}
