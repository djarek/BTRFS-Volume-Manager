package osinterface

/*
#cgo LDFLAGS: -lbtrfs -lblkid
#include "btrfs.h"
#include <mntent.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unsafe"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
)

const (
	mTabFilePath   = "/proc/mounts"
	setmntentFlags = "r"
	btrfsDevType   = "btrfs"
)

var (
	mTabFilePathCString   = C.CString(mTabFilePath)
	setmntentFlagsCString = C.CString(setmntentFlags)

	BlockDevCache = BlockDeviceCache{
		blockDevsByKIdent: make(map[string]*dtos.BlockDevice),
	}

	MountsCache = MountPointCache{
		mountPointByIdent: make(map[string]*dtos.MountPoint),
	}
)

type MountPointCache struct {
	mtx               sync.RWMutex
	mountPointByIdent map[string]*dtos.MountPoint
	mountPoints       []dtos.MountPoint
}

func (mpc *MountPointCache) RescanMountPoints() (err error) {
	mpc.mtx.Lock()
	defer mpc.mtx.Unlock()

	mountPoints, err := probeMountPoints()
	if err != nil {
		return
	}
	mpc.mountPoints = mountPoints
	mpc.mountPointByIdent = make(map[string]*dtos.MountPoint)

	for i, mountPoint := range mpc.mountPoints {
		mpc.mountPointByIdent[mountPoint.Identifier] = &mountPoints[i]
	}
	return
}

func (mpc *MountPointCache) FindByKernelIdentifier(identifier string) (*dtos.MountPoint, bool) {
	mpc.mtx.RLock()
	defer mpc.mtx.RUnlock()
	mp, ok := mpc.mountPointByIdent[identifier]
	return mp, ok
}

type BlockDeviceCache struct {
	mtx               sync.RWMutex
	blockDevsByKIdent map[string]*dtos.BlockDevice
	blockDevs         []dtos.BlockDevice
}

func (bdc *BlockDeviceCache) RescanBlockDevs() (err error) {
	bdc.mtx.Lock()
	defer bdc.mtx.Unlock()

	blockDevs, err := probeBlockDevices()
	if err != nil {
		return
	}

	bdc.blockDevs = blockDevs
	bdc.blockDevsByKIdent = make(map[string]*dtos.BlockDevice)

	for i, blockDev := range bdc.blockDevs {
		bdc.blockDevsByKIdent[blockDev.Path] = &bdc.blockDevs[i]
	}
	return
}

func (bdc *BlockDeviceCache) FindByKernelIdentifier(identifier string) (*dtos.BlockDevice, bool) {
	bdc.mtx.RLock()
	defer bdc.mtx.RUnlock()
	bd, ok := bdc.blockDevsByKIdent[identifier]
	return bd, ok
}

func probeBlockDevices() ([]dtos.BlockDevice, error) {
	var devArray C.struct_block_devices_array
	if C.get_devices(&devArray) != 0 {
		return nil, errors.New("Unable to retrieve device list")
	}
	defer C.block_devices_array_free(devArray)

	devArraySlice := (*[1 << 30]C.struct_block_device)(unsafe.Pointer(devArray.devs))[:devArray.count:devArray.count]

	ret := make([]dtos.BlockDevice, devArray.count)
	for i, dev := range devArraySlice {
		ret[i].Path = C.GoString(dev.dev_name)
		ret[i].UUID = dtos.UUIDType(C.GoString(dev.UUID))
		ret[i].Type = C.GoString(dev._type)
	}
	return ret, nil
}

func probeMountPoints() ([]dtos.MountPoint, error) {
	var mnt C.struct_mntent
	var buf [4096]C.char
	var ret []dtos.MountPoint

	mTab := C.setmntent(mTabFilePathCString, setmntentFlagsCString)
	if mTab == nil {
		return nil, ErrorMTabOpen
	}
	defer C.endmntent(mTab)

	for {
		mntPtr := C.getmntent_r(mTab, &mnt, &buf[0], C.int(len(buf)))
		if mntPtr == nil {
			break
		}

		mountPoint := dtos.MountPoint{
			Identifier:    C.GoString(mntPtr.mnt_fsname),
			MountPath:     C.GoString(mntPtr.mnt_dir),
			MountType:     C.GoString(mntPtr.mnt_type),
			MountOptions:  C.GoString(mntPtr.mnt_opts),
			DumpFrequency: int(mntPtr.mnt_freq),
			FSCKPassNo:    int(mntPtr.mnt_passno),
		}
		ret = append(ret, mountPoint)
	}
	return ret, nil
}

var devMatcher = regexp.MustCompile("path ([a-zA-Z0-9\\/]+)")

func probeBtrfsVolumes() (vols []dtos.BtrfsVolume, err error) {
	output, err := runBtrfsCommand("filesystem", "show", "--all-devices")
	if err != nil {
		return
	}
	volBlocks := strings.Split(output, "Label:")

	volBlocks = volBlocks[1:]
	for _, volBlock := range volBlocks {
		var volume dtos.BtrfsVolume
		_, err = fmt.Sscanf(volBlock, "%s uuid: %s\n", &volume.Label, &volume.UUID)

		foundMatches := devMatcher.FindAllStringSubmatch(volBlock, -1)
		for _, devMatch := range foundMatches {
			dev, ok := BlockDevCache.FindByKernelIdentifier(devMatch[1])
			if ok {
				volume.Devices = append(volume.Devices, dev)
			}
		}
	}
	return
}
