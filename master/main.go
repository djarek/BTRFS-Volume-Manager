package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/djarek/btrfs-volume-manager/master/authentication"
	"github.com/djarek/btrfs-volume-manager/master/db"
	"github.com/djarek/btrfs-volume-manager/master/storageservers"
	"github.com/djarek/btrfs-volume-manager/master/storageservers/blockdevices"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/router"
	"github.com/djarek/btrfs-volume-manager/common/wsprotocol"
)

func setupAuth(r *router.Router) {
	authService := authentication.NewService(db.UsersRepo)
	authCtrl := authentication.NewController(authService)
	authCtrl.ExportHandlers(r)
}

func setupServerTracker(r *router.Router) {
	tracker := storageservers.NewTracker()
	serverController := storageservers.NewController(tracker)
	blockDevController := blockdevices.NewController(tracker)
	blockDevController.ExportHandlers(r)
	serverController.ExportHandlers(r)
}

func main() {
	db.StartDB()
	defer db.StopDB()

	port := flag.Int("port", 8080, "port to serve on")
	dir := flag.String("directory", "views/app/", "directory of views")
	flag.Parse()

	fs := http.Dir(*dir)
	fileHandler := http.FileServer(fs)
	http.Handle("/", fileHandler)

	wsRouter := router.New()
	setupAuth(wsRouter)
	setupServerTracker(wsRouter)
	connectionManager := wsprotocol.NewConnectionUpgrader(
		dtos.JSONMessageMarshaller{},
		wsRouter)
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
