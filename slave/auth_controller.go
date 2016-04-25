package main

import (
	"errors"
	"log"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
	"github.com/djarek/btrfs-volume-manager/common/router"
)

const (
	storageServerIDSessionKey = "StorageServerID"
)

type authController struct{}

type authError struct{}

func (authError) Error() string {
	return "Invalid username or password."
}

func (a *authController) ExportHandlers(adder router.HandlerAdder) {
	adder.AddHandler(dtos.WSMsgAuthenticationResponse, router.DefaultResponseHandler)
	adder.AddHandler(dtos.WSMsgStorageServerRegistrationResponse, router.DefaultResponseHandler)
}

func (a *authController) sendAuthenticationRequest(ctx *request.Context, username string, password string) error {
	requestID, responseChannel := ctx.NewRequest()
	authReq := &dtos.AuthenticationRequest{Username: username, Password: password}
	msg := dtos.NewWebSocketMessage(requestID, authReq)
	errChan := ctx.SendAsync(msg)
	err := <-errChan
	if err != nil {
		return err
	}

	responseMsg, ok := <-responseChannel
	if !ok {
		err = errors.New("Connection closed")
		return err
	}

	response := responseMsg.Payload.(*dtos.AuthenticationResponse)
	if response.Result == "auth_ok" {
		log.Println("Authenticated successfully.")
	} else {
		err = authError{}
		log.Println(err)
	}

	return nil
}

func (a *authController) sendServerRegistrationRequest(ctx *request.Context, serverName string) error {
	requestID, responseChannel := ctx.NewRequest()
	regRequest := &dtos.StorageServerRegistrationRequest{
		ServerName: serverName,
	}

	reqMsg := dtos.NewWebSocketMessage(requestID, regRequest)
	errChan := ctx.SendAsync(reqMsg)
	err := <-errChan
	if err != nil {
		return err
	}

	responseMsg, ok := <-responseChannel
	if !ok {
		err = errors.New("Connection closed")
		return err
	}
	if responseMsg.MessageType == dtos.WSMsgError {
		err := responseMsg.Payload.(*dtos.Error)
		return err
	}

	response := responseMsg.Payload.(*dtos.StorageServerRegistrationResponse)
	ctx.SetSessionData(storageServerIDSessionKey, response.AssignedID)
	log.Println("Storage server registered successfully.")
	return nil
}
