package authentication

import (
	"errors"
	"testing"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type asyncSenderCloserMock struct {
	mock.Mock
}

func (a *asyncSenderCloserMock) SendAsync(msg dtos.WebSocketMessage) <-chan error {
	args := a.Called(msg)
	return args.Get(0).(<-chan error)
}

func (a *asyncSenderCloserMock) Close() {
	a.Called()
}

type authMock struct {
	mock.Mock
}

func (a *authMock) Authenticate(r dtos.AuthenticationRequest) error {
	args := a.Called(r)
	return args.Error(0)
}

func TestOnAuthenticationRequestAuthSuccess(t *testing.T) {
	cMock := &asyncSenderCloserMock{}
	aMock := &authMock{}
	ctrl := controller{
		auth: aMock,
	}
	ctx := &request.Context{AsyncSenderCloser: cMock}
	authReq := dtos.AuthenticationRequest{}
	msg := dtos.NewWebSocketMessage(0, &authReq)
	respMsg := dtos.NewWebSocketMessage(0, &dtos.AuthenticationResponse{Result: "auth_ok"})
	newWSMsg = func(r int64, p dtos.PayloadType) dtos.WebSocketMessage {
		assert.EqualValues(t, p, respMsg.Payload)
		return respMsg
	}
	var r <-chan error

	cMock.On("SendAsync", respMsg).Return(r)
	aMock.On("Authenticate", authReq).Return(nil)
	ctrl.onAuthenticationRequest(ctx, msg)

	cMock.AssertExpectations(t)
	aMock.AssertExpectations(t)
}

func TestOnAuthenticationRequestAuthFailure(t *testing.T) {
	cMock := &asyncSenderCloserMock{}
	aMock := &authMock{}
	ctrl := controller{
		auth: aMock,
	}
	ctx := &request.Context{AsyncSenderCloser: cMock}
	authReq := dtos.AuthenticationRequest{}
	msg := dtos.NewWebSocketMessage(0, &authReq)
	respMsg := dtos.NewWebSocketMessage(0, &dtos.AuthenticationResponse{Result: "auth_wrong"})
	newWSMsg = func(r int64, p dtos.PayloadType) dtos.WebSocketMessage {
		assert.EqualValues(t, p, respMsg.Payload)
		return respMsg
	}

	var r <-chan error
	aMock.On("Authenticate", authReq).Return(errors.New(""))
	cMock.On("SendAsync", respMsg).Return(r)
	ctrl.onAuthenticationRequest(ctx, msg)

	cMock.AssertExpectations(t)
	aMock.AssertExpectations(t)
}
