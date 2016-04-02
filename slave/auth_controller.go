package main

import (
	"log"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/router"
)

type authController struct{}

func (a authController) ExportHandlers(adder router.HandlerAdder) {
	adder.AddHandler(dtos.WSMsgAuthenticationResponse, a.onAuthenticationResponse)
}

func (a authController) onAuthenticationResponse(ctx *router.Context, msg dtos.WebSocketMessage) {
	response := msg.Payload.(*dtos.AuthenticationResponse)

	if response.Result == "auth_ok" {
		log.Println("Authenticated successfully.")
	} else {
		log.Println("Invalid username or password.")
	}
}
