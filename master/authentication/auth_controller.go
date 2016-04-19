package authentication

import (
	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/router"
)

var newWSMsg = dtos.NewWebSocketMessage

/*controller handles all authentication-related Messages.*/
type controller struct {
	auth AuthService
}

//NewController constructs a new authentication controller
func NewController(a AuthService) router.HandlerExporter {
	return &controller{auth: a}
}

//ExportHandlers adds this Controller's handlers to the router.
func (a *controller) ExportHandlers(adder router.HandlerAdder) {
	adder.AddHandler(dtos.WSMsgAuthenticationRequest, a.onAuthenticationRequest)
	adder.AddHandler(dtos.WSMsgLogoutRequest, a.onLogoutRequest)
	adder.AddHandler(dtos.WSMsgReauthenticationRequest, a.onReauthenticationRequest)
}

func (a *controller) onLogoutRequest(ctx *router.Context, msg dtos.WebSocketMessage) {
	//TODO: clear session and close connection
	ctx.Sender.Close()
}

func (a *controller) onReauthenticationRequest(ctx *router.Context, msg dtos.WebSocketMessage) {
	response := dtos.AuthenticationResponse{
		Result: "auth_ok",
	}
	//TODO: session token validation
	responseMsg := newWSMsg(msg.RequestID, &response)
	ctx.Sender.SendAsync(responseMsg)
}

func (a *controller) onAuthenticationRequest(ctx *router.Context, msg dtos.WebSocketMessage) {
	credentials := msg.Payload.(*dtos.AuthenticationRequest)

	response := dtos.AuthenticationResponse{
		Result: "auth_wrong",
	}
	authErr := a.auth.Authenticate(*credentials)
	if authErr == nil {
		response.Result = "auth_ok"
		response.UserDetails = credentials.Username
	}

	responseMsg := newWSMsg(msg.RequestID, &response)
	ctx.Sender.SendAsync(responseMsg)
	//TODO: Session creation
}
