package dtos

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebsocketMessageUnmarshalling(t *testing.T) {
	var msgTypeString = strconv.Itoa(WSMsgAuthenticationRequest)
	var validJSON = "{\"messageType\":" + msgTypeString +
		",\"payload\":{\"username\":\"username\", \"password\":\"password\"},\"requestID\":1}"

	var msg WebSocketMessage
	err := json.Unmarshal([]byte(validJSON), &msg)
	assert.Nil(t, err, "JSON string: "+validJSON)

	assert.EqualValues(t, WSMsgAuthenticationRequest, msg.MessageType,
		"messageType should be WSMsgAuthenticationRequest: "+msgTypeString)

	authReq := msg.Payload.(*AuthenticationRequest)
	assert.EqualValues(t, "username", authReq.Username)
	assert.EqualValues(t, "password", authReq.Password)
}

func TestWebsocketMessageNew(t *testing.T) {
	const requestID = 1
	msg := NewWebSocketMessage(requestID, &AuthenticationRequest{})
	assert.EqualValues(t, WSMsgAuthenticationRequest, msg.MessageType,
		"messageType should be WSMsgAuthenticationRequest")

	assert.EqualValues(t, requestID, msg.RequestID, "invalid ")
}

type unregisteredPayload struct{}

func (unregisteredPayload) isPayload() {}

func TestWebsocketMessageNewPanic(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewWebSocketMessage(0, unregisteredPayload{})
	})
}

func TestWebsocketMessageMarshalling(t *testing.T) {
	var msgTypeString = strconv.Itoa(WSMsgAuthenticationRequest)
	var expectedJSON = "{\"messageType\":" + msgTypeString +
		",\"requestID\":1,\"payload\":{\"username\":\"username\",\"password\":\"password\"}}"
	msg := NewWebSocketMessage(1, &AuthenticationRequest{"username", "password"})

	buf, err := json.Marshal(msg)
	assert.Nil(t, err)

	assert.EqualValues(t, expectedJSON, string(buf))
}
