package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/router"
	"github.com/djarek/btrfs-volume-manager/common/wsprotocol"
	"github.com/djarek/btrfs-volume-manager/slave/osinterface"
)

const (
	masterControlURL = "ws://localhost:8080/ws"
	defaultUsername  = "admin"
	defaultPassword  = "admin"
)

func main() {
	r := router.New()
	authController{}.ExportHandlers(r)
	conn, err := wsprotocol.DefaultDialer.Dial(masterControlURL, r)
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
			time.Sleep(time.Millisecond * 1000)
			log.Println("Sending msg")
		}
	}()
	osinterface.BlockDeviceCache.RescanBlockDevs()
	osinterface.MountPointCache.RescanMountPoints()
	vol := dtos.BtrfsVolume{UUID: "e52c00b9-60b2-468a-83cc-e6c652f098f7"}
	_, ok := osinterface.MountPointCache.FindRootMount(vol.UUID)
	if !ok {
		log.Println("lll")
		_, err = osinterface.MountBtrfsRoot(vol)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
	bds, _ := osinterface.MountPointCache.FindByKernelIdentifier("/dev/sdc1")
	log.Println(osinterface.CreateSubVolume(vol, "newvol"))
	log.Println(osinterface.DeleteSubVolume(vol, "newvol"))
	if ok {
		log.Println(bds[0])
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
