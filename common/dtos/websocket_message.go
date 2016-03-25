package dtos

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strconv"
)

//WebSocketMessageType represents the type of the message.
type WebSocketMessageType int32

const (
	/*WSMsgAuthenticationRequest indicates this WebSocketMessage contains an
	AuthenticationRequest*/
	WSMsgAuthenticationRequest = iota + 1
)

const (
	//WSMsgError indicates this WebSocketMessage contains an error object
	WSMsgError WebSocketMessageType = iota + 10000
	/*WSMsgAuthenticationResponse indicates this WebSocketMessage contains an
	AuthenticationResponse*/
	WSMsgAuthenticationResponse
)

//WebSocketMessage represents a message received from a client or
//ready to be sent to it
type WebSocketMessage struct {
	MessageType WebSocketMessageType `json:"messageType"`
	RequestID   int64                `json:"requestID"`
	Payload     interface{}          `json:"payload"`
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

/*Error represents an error that occured in the higher layers and is supposed
to be sent to the client. The subsystem string indicates which entity emitted
the error.*/
type Error struct {
	Subsystem string `json:"subsystem"`
	Details   string `json:"details"`
}

func (e Error) Error() string {
	return e.Subsystem + " error: " + e.Details
}

/*NewErrorMsg constructs a WebSocketMessage from an error message*/
func NewErrorMsg(subsystem string, err error, requestID int64) WebSocketMessage {
	errStruct := Error{
		Subsystem: subsystem,
		Details:   err.Error(),
	}
	buf, err := json.Marshal(errStruct)
	if err != nil {
		log.Fatalln("Marshalling error message failed: " + err.Error())
	}

	return WebSocketMessage{
		MessageType: WSMsgError,
		Payload:     (*json.RawMessage)(&buf),
		RequestID:   requestID,
	}
}

var typeMap = make(map[WebSocketMessageType]reflect.Type)
var reversedTypeMap = make(map[reflect.Type]WebSocketMessageType)

func init() {
	RegisterMessageType(WSMsgAuthenticationRequest, AuthenticationRequest{})
	RegisterMessageType(WSMsgAuthenticationResponse, AuthenticationResponse{})
	RegisterMessageType(WSMsgError, Error{})
}

/*RegisterMessageType registers the type for both marshalling and unmarshalling.
This function is NOT thread-safe and should be preferably called in the init()
function of a higher-level package.*/
func RegisterMessageType(typeID WebSocketMessageType, payload interface{}) {
	_, found := typeMap[typeID]
	if found {
		log.Panicf("Type(ID: %d, %v) already registered!  ", typeID, payload)
	}
	t := reflect.ValueOf(payload).Type()
	typeMap[typeID] = t
	reversedTypeMap[t] = typeID
}

/*NewWebSocketMessage constructs a WebSocketMessage and sets the appropriate
messageType. If the payload type is not in the typemap the function will panic.*/
func NewWebSocketMessage(requestID int64, p interface{}) WebSocketMessage {
	v := reflect.ValueOf(p)
	msgType, found := reversedTypeMap[v.Type()]
	if !found {
		log.Panicf("Unknown payload type (payload:%s)", v.Type().String())
	}
	return WebSocketMessage{
		MessageType: msgType,
		RequestID:   requestID,
		Payload:     p,
	}
}

/*UnmarshalJSON unmarshals the json string into a websocket message. If the
message type is unknown or the string is malformed an error is returned.*/
func (w *WebSocketMessage) UnmarshalJSON(data []byte) error {
	temp := struct {
		MessageType WebSocketMessageType `json:"messageType"`
		RequestID   int64                `json:"requestID"`
		PayloadData *json.RawMessage     `json:"payload"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	w.RequestID = temp.RequestID
	w.MessageType = temp.MessageType
	payloadType, found := typeMap[temp.MessageType]
	if !found {
		return errors.New("Unknown message type: " + strconv.Itoa(int(w.MessageType)))
	}
	w.Payload = reflect.New(payloadType).Interface()
	return json.Unmarshal([]byte(*temp.PayloadData), w.Payload)
}
