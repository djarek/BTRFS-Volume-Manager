package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/wsserver"
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

func (a authenticator) Authenticate(addr net.Addr, authMsg []byte) ([]byte, error) {
	var data LoginAndPassword
	err := json.Unmarshal(authMsg, &data)
	if err != nil {
		panic(err)
	}
	usr, err := findByUsername(data.Username)
	if err != nil {
		return []byte("auth_wrong"), errors.New("No such user")
	}
	err = bcrypt.CompareHashAndPassword(
		[]byte(usr.HashedPassword), []byte(data.Password))
	if err == nil {
		return []byte("auth_ok"), nil
	}
	return []byte("auth_wrong"), errors.New("Wrong passsword")
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
	connectionManager := wsserver.NewConnectionManager(dtos.JSONMessageMarshaller{}, messageParser{}, authenticator{})
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
