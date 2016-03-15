package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/wsserver"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
)

const dbName = "btrfs"
const usersCollectionName = "users"

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	connections map[*websocket.Conn]bool
	session     *mgo.Session
	collUsers   *mgo.Collection
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
		return []byte("auth_wrong"), nil
	}
	err = bcrypt.CompareHashAndPassword(
		[]byte(usr.HashedPassword), []byte(data.Password))
	if err == nil {
		return []byte("auth_ok"), nil
	}
	return []byte("auth_wrong"), nil
}

func main() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	collUsers = session.DB(dbName).C(usersCollectionName)

	// Unique index
	index := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = collUsers.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	// Initialize data base if it is empty
	var results []User
	err = collUsers.Find(nil).All(&results)
	if len(results) == 0 {
		initializeDB()
	}

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

	err = http.ListenAndServe(addr, nil)
	log.Fatalln(err.Error())
}
