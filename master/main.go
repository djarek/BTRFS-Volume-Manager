package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/wsprotocol"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	connections map[*websocket.Conn]bool
)

type authenticator struct{}

func (a authenticator) Authenticate(credentials dtos.AuthenticationRequest) error {
	usr, err := usersRepo.FindUserByUsername(credentials.Username)
	if err != nil {
		return errors.New("No such user")
	}
	err = bcrypt.CompareHashAndPassword(
		[]byte(usr.HashedPassword), []byte(credentials.Password))
	if err != nil {
		return errors.New("Wrong passsword")
	}
	return nil
}

func main() {
	startDB()
	defer stopDB()

	port := flag.Int("port", 8080, "port to serve on")
	dir := flag.String("directory", "views/", "directory of views")
	flag.Parse()
	connections = make(map[*websocket.Conn]bool)

	fs := http.Dir(*dir)
	fileHandler := http.FileServer(fs)
	http.Handle("/", fileHandler)

	connectionManager := wsprotocol.NewConnectionManager(
		dtos.JSONMessageMarshaller{},
		messageParser{})
	http.HandleFunc("/ws", connectionManager.HandleWSConnection)

	log.Printf("Running on port %d\n", *port)
	addr := fmt.Sprintf("localhost:%d", *port)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		panic(http.ListenAndServe(addr, nil))
	}()
	<-sigs
}
