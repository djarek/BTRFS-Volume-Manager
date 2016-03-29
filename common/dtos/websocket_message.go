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

var unmarshallingTypeMap = make(map[WebSocketMessageType]reflect.Type)
var marshallingTypeMap = make(map[reflect.Type]WebSocketMessageType)

func init() {
	RegisterMessageType(WSMsgAuthenticationRequest, AuthenticationRequest{})
	RegisterMessageType(WSMsgAuthenticationResponse, AuthenticationResponse{})
	RegisterMessageType(WSMsgError, Error{})
}

/*RegisterMessageType registers the type for both marshalling and unmarshalling.
This function is NOT thread-safe and should be preferably called in the init()
function of a higher-level package.*/
func RegisterMessageType(typeID WebSocketMessageType, payload interface{}) {
	t := reflect.ValueOf(payload).Type()
	_, found1 := unmarshallingTypeMap[typeID]
	_, found2 := marshallingTypeMap[t]
	if found1 || found2 {
		log.Panicf("Type(ID: %d, %v) already registered!", typeID, payload)
	}

	unmarshallingTypeMap[typeID] = t
	marshallingTypeMap[t] = typeID
}

func getMsgTypeID(payload interface{}) (msgType WebSocketMessageType) {
	t := reflect.ValueOf(payload).Type()
	msgType, found := marshallingTypeMap[t]
	if !found {
		log.Panicf("Unknown payload type (payload:%s)", t.String())
	}
	return
}

func newPayloadType(msgType WebSocketMessageType) (payload interface{}, err error) {
	payloadType, found := unmarshallingTypeMap[msgType]
	if !found {
		return nil, errors.New("Unknown message type: " + strconv.Itoa(int(msgType)))
	}
	payload = reflect.New(payloadType).Interface()
	return
}

/*NewWebSocketMessage constructs a WebSocketMessage and sets the appropriate
messageType. If the payload type is not in the typemap the function will panic.*/
func NewWebSocketMessage(requestID int64, p interface{}) WebSocketMessage {
	msgType := getMsgTypeID(p)
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
	w.Payload, err = newPayloadType(temp.MessageType)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(*temp.PayloadData), w.Payload)
}
