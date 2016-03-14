package main

import (
	"fmt"

	"github.com/djarek/btrfs-volume-manager/slave/osinterface"
)

func main() {
	fmt.Print(osinterface.BlockDeviceCache.RescanBlockDevs())
}
