package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"crypto/subtle"

	"github.com/gorilla/websocket"
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

func authentication(loginAndPass LoginAndPassword) bool {
	usr, err := findByUsername(loginAndPass.Username)
	if err != nil {
		return false
	} else if subtle.ConstantTimeCompare(
		[]byte(usr.Password), []byte(loginAndPass.Password)) == 1 {
		return true
	} else {
		return false
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		log.Println(err)
		return
	}
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var data LoginAndPassword
		err = json.Unmarshal([]byte(msg), &data)
		if err != nil {
			panic(err)
		}
		if authentication(data) {
			conn.WriteMessage(messageType, []byte("true"))
		} else {
			conn.WriteMessage(messageType, []byte("false"))
		}
	}
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

	port := flag.Int("port", 80, "port to serve on")
	dir := flag.String("directory", "views/", "directory of views")
	flag.Parse()

	connections = make(map[*websocket.Conn]bool)

	fs := http.Dir(*dir)
	fileHandler := http.FileServer(fs)
	http.Handle("/", fileHandler)
	http.HandleFunc("/auth", authHandler)

	log.Printf("Running on port %d\n", *port)

	addr := fmt.Sprintf("localhost:%d", *port)

	err = http.ListenAndServe(addr, nil)
	log.Fatalln(err.Error())
}
