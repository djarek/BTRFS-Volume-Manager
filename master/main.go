package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var connections map[*websocket.Conn]bool
var initializeDB = false
var dropDB = false
var session *mgo.Session
var collUsers *mgo.Collection
var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// User model prototype without hashing algorithms implemented yet
type User struct {
	ID           bson.ObjectId `bson:"_id,omitempty"`
	Username     string        `bson:"username,omitempty"`
	Password     string        `bson:"password,omitempty"`
	FirstName    string        `bson:"firstName"`
	SecondName   string        `bson:"secondName"`
	RegisterDate time.Time     `bson:"registerDate"`
}

//StorageServer represents a Network Attached Storage device
type StorageServer struct {
	ID   bson.ObjectId `bson:"_id,omitempty"`
	Name string        `bson:"name"`
}

//BlockDevice represents a block device retrieved by blkid probe
type BlockDevice struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	VolID bson.ObjectId `bson:"volID"` //can be empty
	Path  string        `bson:"path,omitempty"`
	UUID  string        `bson:"uuid,omitempty"`
	Type  string        `bson:"type,omitempty"`
}

//BtrfsVolume represents a filesystem volume which can potentially span over
//multiple devices
type BtrfsVolume struct {
	ID     bson.ObjectId `bson:"_id,omitempty"`
	ServID bson.ObjectId `bson:"servID"` // can be empty
	Label  string        `bson:"label"`
}

func authentication(loginAndPass []string) bool {
	result := User{}
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
		data := strings.Split(string(msg), ",")
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
	log.Fatalln(err.Error())
}
