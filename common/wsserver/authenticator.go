package wsserver

import (
	"net"

	"github.com/BTRFS-Volume-Manager/common/dtos"
)

type WebSocketAuthenticator interface {
	GetChallenge(net.Addr) *dtos.WebSocketMessage
	VerifyChallengeResponse(net.Addr, *dtos.WebSocketMessage) error
}
