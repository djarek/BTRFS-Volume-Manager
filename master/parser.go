package main

import (
	"log"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/wsprotocol"
)

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
	credentials, ok := msg.Payload.(*dtos.AuthenticationRequest)
	if !ok {
		log.Printf("Invalid payload type (typeID:%d, payload:%v)\n", msg.MessageType,
			msg.Payload)
		return
	}

	response := dtos.AuthenticationResponse{
		Result: "auth_wrong",
	}
	authErr := auth.Authenticate(*credentials)
	if authErr == nil {
		response.Result = "auth_ok"
	}

	responseMsg := dtos.NewWebSocketMessage(msg.RequestID, response)
	channel, err := connection.SendAsync(responseMsg)
	if err != nil && authErr == nil {
		go func() {
			err := <-channel
			if err == nil {
				//TODO: Create session
			}
		}()
	}
}
