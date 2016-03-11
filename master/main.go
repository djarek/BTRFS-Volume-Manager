package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"./models"
)

var connections map[*websocket.Conn]bool
var initializeDB = false
var dropDB = false
var session *mgo.Session
var collUsers *mgo.Collection

func authentiaction(loginAndPass []string) bool {
	result := models.User{}
	err := collUsers.Find(bson.M{"username": loginAndPass[0]}).One(&result)
	if err != nil {
		return false
	} else if result.Password == loginAndPass[1] {
		return true
	} else {
		return false
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		data := strings.Split(string(msg), ",")
		if authentiaction(data) {
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
	collUsers = session.DB("btrfs").C("users")

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
	fmt.Println(err.Error())
}
