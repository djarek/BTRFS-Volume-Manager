package osinterface

import (
	"errors"
	"syscall"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
)

var mount = func(source string, target string, fstype string, flags uintptr, data string) error {
	return syscall.Mount(source, target, fstype, flags, data)
}

func getBtrfsRootMount(vol dtos.BtrfsVolume) (mountPath string, err error) {
	mount, ok := MountPointCache.FindRootMount(vol.UUID)
	if !ok {
		mountPath, err = MountBtrfsRoot(vol)
		if err != nil {
			return
		}
	} else {
		mountPath = mount.MountPath
	}
	return
}

/*MountBtrfsRoot attempts to mount the specified btrfs volume's root at the
configured path.*/
func MountBtrfsRoot(vol dtos.BtrfsVolume) (rootMountPath string, err error) {
	const errStr = "BTRFS root mount failed: "
	targetPath := rootMountsPath + "/" + string(vol.UUID)
	bds, ok := BlockDeviceCache.FindByUUID(vol.UUID)
	if !ok || len(bds) == 0 {
		return "", errors.New(errStr + "no device present for volume UUID: " + string(vol.UUID))
	}
	dev := bds[0]
	if dev.Type != "btrfs" {
		return "", errors.New(errStr + "not a btrfs device")
	}

	err = mount(dev.Path, targetPath, dev.Type, 0, "subvolid=0")
	if err != nil {
		return "", errors.New(errStr + err.Error())
	}

	err = MountPointCache.Rescan()
	if err != nil {
		return "", errors.New(errStr + err.Error())
	}
	return targetPath, nil
}
