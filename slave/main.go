package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/wsprotocol"
	_ "github.com/djarek/btrfs-volume-manager/slave/osinterface"
)

const (
	masterControlURL = "ws://localhost:8080/ws"
	defaultUsername  = "admin"
	defaultPassword  = "admin"
)

func main() {
	conn, err := wsprotocol.DefaultDialer.Dial(masterControlURL, messageParser{})
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		conn.Close()
		time.Sleep(time.Second * 1)
	}()
	go func() {
		for {
			msg := dtos.WebSocketMessage{}
			conn.SendAsync(msg)
			time.Sleep(time.Millisecond * 1)
			log.Println("Sending msg")
		}
	}()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
