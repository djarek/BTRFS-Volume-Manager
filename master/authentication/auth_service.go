package authentication

import (
	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/master/models"
	"golang.org/x/crypto/bcrypt"
)

//AuthService is a service that performs user authentication
type AuthService interface {
	Authenticate(dtos.AuthenticationRequest) error
}

type authenticator struct {
	usersRepo userFinder
}

type userFinder interface {
	FindUserByUsername(string) (models.User, error)
}

//NewService constructs a new AuthService
func NewService(f userFinder) AuthService {
	return &authenticator{f}
}

//ErrInvalidUserOrPasswd indicates authentication failed due to invalid credentials
type ErrInvalidUserOrPasswd struct{}

func (ErrInvalidUserOrPasswd) Error() string {
	return "Invalid username or password"
}

func (a authenticator) Authenticate(credentials dtos.AuthenticationRequest) error {
	usr, err := a.usersRepo.FindUserByUsername(credentials.Username)
	if err != nil {
		return ErrInvalidUserOrPasswd{}
	}
	err = bcrypt.CompareHashAndPassword(
		[]byte(usr.HashedPassword), []byte(credentials.Password))
	if err != nil {
		return ErrInvalidUserOrPasswd{}
	}
	return nil
}
