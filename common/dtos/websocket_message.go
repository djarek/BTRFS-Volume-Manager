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

//WSMsgRequests MessageType values
const (
	WSMsgAuthenticationRequest            = 1
	WSMsgLogoutRequest                    = 2
	WSMsgReauthenticationRequest          = 3
	WSMsgStorageServerRegistrationRequest = 4
	WSMsgBlockDeviceRescanRequest         = 5
	WSMsgStorageServerListRequest         = 6
	WSMsgBlockDeviceListRequest           = 7
)

//WSMsgResponse MessageType values
const (
	WSMsgError                             = 10000
	WSMsgAuthenticationResponse            = 10001
	WSMsgStorageServerRegistrationResponse = 10004
	WSMsgBlockDeviceRescanResponse         = 10005
	WSMsgStorageServerListResponse         = 10006
	WSMsgBlockDeviceListResponse           = 10007
)

func init() {
	RegisterMessageType(WSMsgAuthenticationRequest, AuthenticationRequest{})
	RegisterMessageType(WSMsgAuthenticationResponse, AuthenticationResponse{})
	RegisterMessageType(WSMsgLogoutRequest, LogoutRequest{})
	RegisterMessageType(WSMsgReauthenticationRequest, ReauthenticationRequest{})

	RegisterMessageType(WSMsgStorageServerRegistrationRequest, StorageServerRegistrationRequest{})
	RegisterMessageType(WSMsgStorageServerRegistrationResponse, StorageServerRegistrationResponse{})

	RegisterMessageType(WSMsgBlockDeviceRescanRequest, BlockDeviceRescanRequest{})
	RegisterMessageType(WSMsgBlockDeviceRescanResponse, BlockDeviceRescanResponse{})

	RegisterMessageType(WSMsgStorageServerListRequest, StorageServerListRequest{})
	RegisterMessageType(WSMsgStorageServerListResponse, StorageServerListResponse{})

	RegisterMessageType(WSMsgBlockDeviceListRequest, BlockDeviceListRequest{})
	RegisterMessageType(WSMsgBlockDeviceListResponse, BlockDeviceListResponse{})

	RegisterMessageType(WSMsgError, Error{})
}

//WebSocketMessage represents a message received from a client or
//ready to be sent to it
type WebSocketMessage struct {
	MessageType WebSocketMessageType `json:"messageType"`
	RequestID   int64                `json:"requestID"`
	Payload     PayloadType          `json:"payload"`
}

/*PayloadType represents a type that is a WebSocketMessage payload. The method
is used to provide some type safety. A pointer receiver is recommended.*/
type PayloadType interface {
	isPayload()
}

//WebSocketMessageMarshaller allows conversion from byte slices to WSMessage structs
//and vice versa.
type WebSocketMessageMarshaller interface {
	Marshal(WebSocketMessage) ([]byte, error)
	Unmarshal([]byte) (WebSocketMessage, error)
}

//JSONMessageMarshaller is the default WebSocketMessageMarshaller - uses JSON as the target format
type JSONMessageMarshaller struct{}

//Marshal encodes a WebSocketMessage as a JSON object
func (j JSONMessageMarshaller) Marshal(msg WebSocketMessage) (buf []byte, err error) {
	buf, err = json.Marshal(msg)
	return
}

//Unmarshal decodes a JSON object to a WebSocketMessage
func (JSONMessageMarshaller) Unmarshal(buf []byte) (msg WebSocketMessage, err error) {
	err = json.Unmarshal(buf, &msg)
	return
}

//BasePayload should be embedded into structs that are supposed to be sent as
//payloads.
type BasePayload struct{}

func (*BasePayload) isPayload() {}

//AuthenticationRequest represents a request from the client to perform authentication
type AuthenticationRequest struct {
	BasePayload `json:"-"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

/*AuthenticationResponse represents a response to the client indicating whether
authentication succeeded or failed*/
type AuthenticationResponse struct {
	BasePayload `json:"-"`
	Result      string `json:"result"`
	UserDetails string `json:"userDetails"`
}

/*LogoutRequest represents a request from the client to end the session*/
type LogoutRequest struct {
	BasePayload `json:"-"`
}

/*ReauthenticationRequest represents a request from the client to reuse a previous
session.*/
type ReauthenticationRequest struct {
	BasePayload `json:"-"`
	Token       string `json:"token"`
}

/*StorageServerRegistrationRequest represents a request from a storage server to
register it in the server tracker*/
type StorageServerRegistrationRequest struct {
	BasePayload `json:"-"`
	ServerName  string `json:"serverName"`
}

/*StorageServerRegistrationResponse represents a request from a storage server to
register it in the server tracker*/
type StorageServerRegistrationResponse struct {
	BasePayload `json:"-"`
	AssignedID  StorageServerID `json:"assignedID"`
}

/*BlockDeviceRescanRequest represents a request to the slave to perform a scan
of block devices in the system*/
type BlockDeviceRescanRequest struct {
	BasePayload `json:"-"`
	ServerID    StorageServerID `json:"serverID"`
}

/*BlockDeviceRescanResponse represents a response from the slave containing  a list
of all block devices present in the system*/
type BlockDeviceRescanResponse struct {
	BasePayload  `json:"-"`
	BlockDevices []BlockDevice `json:"blockDevices"`
}

/*StorageServerListRequest represents a request from the client to retrieve a list of
all storage servers.*/
type StorageServerListRequest struct {
	BasePayload `json:"-"`
}

/*StorageServerListResponse represents a response to the client with the list of
all storage servers.*/
type StorageServerListResponse struct {
	BasePayload `json:"-"`
	Servers     []StorageServer `json:"servers"`
}

/*BlockDeviceListRequest represents a request from the client to retrieve a list of
all block devices present on the slave.*/
type BlockDeviceListRequest struct {
	BasePayload `json:"-"`
	ServerID    StorageServerID `json:"serverID"`
}

/*BlockDeviceListResponse represents a response to the client with the list of all
block devices present on the slave*/
type BlockDeviceListResponse struct {
	BasePayload  `json:"-"`
	BlockDevices []BlockDevice `json:"blockDevices"`
}

/*Error represents an error that occured in the higher layers and is supposed
to be sent to the client. The subsystem string indicates which entity emitted
the error.*/
type Error struct {
	BasePayload `json:"-"`
	Subsystem   string `json:"subsystem"`
	Details     string `json:"details"`
}

func (e Error) Error() string {
	return e.Subsystem + " error: " + e.Details
}

var unmarshallingTypeMap = make(map[WebSocketMessageType]reflect.Type)
var marshallingTypeMap = make(map[reflect.Type]WebSocketMessageType)

/*RegisterMessageType registers the type for both marshalling and unmarshalling.
This function is NOT thread-safe and should be preferably called in the init()
function of a higher-level package.*/
func RegisterMessageType(typeID WebSocketMessageType, payload interface{}) {
	v := reflect.ValueOf(payload)
	t := reflect.Indirect(v).Type()
	_, found1 := unmarshallingTypeMap[typeID]
	_, found2 := marshallingTypeMap[t]
	if found1 || found2 {
		log.Panicf("Type(ID: %d, %v) already registered!", typeID, payload)
	}

	unmarshallingTypeMap[typeID] = t
	marshallingTypeMap[t] = typeID
}

func getMsgTypeID(payload interface{}) (msgType WebSocketMessageType) {
	v := reflect.Indirect(reflect.ValueOf(payload))
	t := v.Type()
	msgType, found := marshallingTypeMap[t]
	if !found {
		log.Panicf("Unknown payload type (payload:%s)", t.String())
	}
	return
}

func newPayloadType(msgType WebSocketMessageType) (payload PayloadType, err error) {
	payloadType, found := unmarshallingTypeMap[msgType]
	if !found {
		return nil, errors.New("Unknown message type: " + strconv.Itoa(int(msgType)))
	}

	payload = reflect.New(payloadType).Interface().(PayloadType)
	return
}

/*NewWebSocketMessage constructs a WebSocketMessage and sets the appropriate
messageType. If the payload type is not in the typemap the function will panic.*/
func NewWebSocketMessage(requestID int64, p PayloadType) WebSocketMessage {
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
