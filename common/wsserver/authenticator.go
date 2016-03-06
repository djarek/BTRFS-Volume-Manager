package wsserver

import (
	"net"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
)

//WebSocketAuthenticator represents an object used for authentication of a newly
//connected websocket client.
type WebSocketAuthenticator interface {
	GetChallenge(net.Addr) *dtos.WebSocketMessage
	VerifyChallengeResponse(net.Addr, *dtos.WebSocketMessage) error
}
