package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djarek/btrfs-volume-manager/common/router"
	"github.com/djarek/btrfs-volume-manager/common/wsprotocol"
)

const (
	masterControlURL  = "ws://localhost:8080/ws"
	defaultUsername   = "admin"
	defaultPassword   = "admin"
	defaultServerName = "StorageServer1"
)

func main() {
	r := router.New()
	auth := &authController{}
	auth.ExportHandlers(r)
	bdCtrl := blockDevController{}
	bdCtrl.ExportHandlers(r)
	ctx, err := wsprotocol.DefaultDialer.Dial(masterControlURL, r)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		ctx.Close()
		time.Sleep(time.Second * 1)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	err = auth.sendAuthenticationRequest(ctx, defaultUsername, defaultPassword)
	if err != nil {
		return
	}

	err = auth.sendServerRegistrationRequest(ctx, defaultServerName)
	if err != nil {
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
