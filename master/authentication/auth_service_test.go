package authentication

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/master/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type finderMock struct {
	mock.Mock
}

func (f *finderMock) FindUserByUsername(username string) (models.User, error) {
	args := f.Called(username)
	return args.Get(0).(models.User), args.Error(1)
}

func TestAuthenticationSuccess(t *testing.T) {
	f := finderMock{}
	a := authenticator{&f}

	authReq := dtos.AuthenticationRequest{Username: "username", Password: "password"}
	hash, err := bcrypt.GenerateFromPassword([]byte(authReq.Password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	user := models.User{
		Username:       authReq.Username,
		HashedPassword: string(hash),
	}

	f.On("FindUserByUsername", authReq.Username).Return(user, nil)
	err = a.Authenticate(authReq)
	assert.NoError(t, err)
	f.AssertExpectations(t)
}

func TestAuthenticationUserNotFound(t *testing.T) {
	f := finderMock{}
	a := authenticator{&f}
	authReq := dtos.AuthenticationRequest{Username: "username", Password: "password"}

	f.On("FindUserByUsername", authReq.Username).Return(models.User{}, errors.New("User not found"))
	err := a.Authenticate(authReq)
	assert.EqualValues(t, ErrInvalidUserOrPasswd{}, err)
	f.AssertExpectations(t)
}

func TestAuthenticationInvalidPassword(t *testing.T) {
	f := finderMock{}
	a := authenticator{&f}
	authReq := dtos.AuthenticationRequest{Username: "username", Password: "password"}
	user := models.User{
		Username:       authReq.Username,
		HashedPassword: "invalid_hash",
	}
	f.On("FindUserByUsername", authReq.Username).Return(user, nil)
	err := a.Authenticate(authReq)
	assert.EqualValues(t, ErrInvalidUserOrPasswd{}, err)
	f.AssertExpectations(t)
}
