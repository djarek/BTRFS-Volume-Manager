package wsserver

import (
	"net"

	"github.com/djarek/BTRFS-Volume-Manager/common/dtos"
)

//WebSocketAuthenticator represents an object used for authentication of a newly
//connected websocket client.
type WebSocketAuthenticator interface {
	GetChallenge(net.Addr) *dtos.WebSocketMessage
	VerifyChallengeResponse(net.Addr, *dtos.WebSocketMessage) error
}
