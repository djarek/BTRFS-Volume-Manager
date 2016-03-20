package main

import (
	"encoding/json"
	"log"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/wsprotocol"
)

func marshalPayload(out **json.RawMessage, v interface{}) {
	buf, err := json.Marshal(v)
	if err != nil {
		log.Fatalln(err)
	}
	*out = (*json.RawMessage)(&buf)
}

/*messageParser parses the received WebSocketMessage and dispatches appropriate
handler functions. Implements the RecvMessageParser interface.
*/
type messageParser struct{}

/*ParseRecvMsg parses the received WebSocketMessage and dispatches appropriate
handler functions. */
func (mp messageParser) ParseRecvMsg(msg dtos.WebSocketMessage, connection *wsprotocol.Connection) (err error) {
	switch msg.MessageType {
	case dtos.WSMsgAuthenticationRequest:
		{
			s := connection.GetSession()
			if s != nil {
				log.Println("Reauthentication of connection not allowed.")
				break
			}
			onAuthenticationRequest(msg, connection)
		}
	}
	return nil
}

var auth = authenticator{}

func onAuthenticationRequest(msg dtos.WebSocketMessage, connection *wsprotocol.Connection) {
	var credentials dtos.AuthenticationRequest
	err := json.Unmarshal(*msg.Payload, &credentials)
	if err != nil {
		log.Println("Error when unmarshalling credentials: " + err.Error())
		return
	}
	response := dtos.AuthenticationResponse{
		Result: "auth_wrong",
	}

	authErr := auth.Authenticate(credentials)
	if authErr == nil {
		response.Result = "auth_ok"
	}
	marshalPayload(&msg.Payload, response)
	msg.MessageType = dtos.WSMsgAuthenticationResponse

	channel, err := connection.SendAsync(msg)
	if err != nil && authErr == nil {
		go func() {
			err := <-channel
			if err == nil {
				//TODO: Create session
			}
		}()
	}
}
