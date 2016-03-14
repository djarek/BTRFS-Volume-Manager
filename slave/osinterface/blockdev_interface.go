package osinterface

/*
#cgo LDFLAGS: -lblkid
#include <mntent.h>
#include <blkid/blkid.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
)

const (
	mTabFilePath   = "/proc/mounts"
	setmntentFlags = "r"
	btrfsDevType   = "btrfs"
)

var (
	mTabFilePathCString     = C.CString(mTabFilePath)
	setmntentFlagsCString   = C.CString(setmntentFlags)
	blkidUUIDTagNameCString = C.CString("UUID")
	blkidTypeTagNameCString = C.CString("TYPE")

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
	blockDevs, err := probeBlockDevices()
	if err != nil {
		return
	}

	blockDevsByKIdent := make(map[string]*dtos.BlockDevice)
	for i, blockDev := range bdc.blockDevs {
		bdc.blockDevsByKIdent[blockDev.Path] = &bdc.blockDevs[i]
	}

	bdc.mtx.Lock()
	defer bdc.mtx.Unlock()

	bdc.blockDevs = blockDevs
	bdc.blockDevsByKIdent = blockDevsByKIdent
	return
}

func (bdc *BlockDeviceCache) FindByKernelIdentifier(identifier string) (*dtos.BlockDevice, bool) {
	bdc.mtx.RLock()
	defer bdc.mtx.RUnlock()
	bd, ok := bdc.blockDevsByKIdent[identifier]
	return bd, ok
}

func blkidDevToBlockDev(dev C.blkid_dev) (blockDev dtos.BlockDevice) {
	blockDev.Path = C.GoString(C.blkid_dev_devname(dev))
	blkidTagIterator := C.blkid_tag_iterate_begin(dev)
	var tagValue, tagType *C.char

	for C.blkid_tag_next(blkidTagIterator, &tagType, &tagValue) == 0 {
		if C.strcmp(tagType, blkidUUIDTagNameCString) == 0 {
			blockDev.UUID = dtos.UUIDType(C.GoString(tagValue))
		} else if C.strcmp(tagType, blkidTypeTagNameCString) == 0 {
			blockDev.Type = C.GoString(tagValue)
		}
	}
	return
}

func probeBlockDevices() ([]dtos.BlockDevice, error) {
	var blkidCache C.blkid_cache
	var blkidDevIterator C.blkid_dev_iterate
	var blkidDev C.blkid_dev

	if C.blkid_get_cache(&blkidCache, nil) < 0 {
		return nil, ErrBlkidGetCache
	}
	defer C.blkid_put_cache(blkidCache)

	C.blkid_probe_all(blkidCache)
	blkidDevIterator = C.blkid_dev_iterate_begin(blkidCache)

	var blockDevs []dtos.BlockDevice

	for C.blkid_dev_next(blkidDevIterator, &blkidDev) == 0 {
		blkidDev = C.blkid_verify(blkidCache, blkidDev)

		if blkidDev != nil {
			dev := blkidDevToBlockDev(blkidDev)
			blockDevs = append(blockDevs, dev)
		}
	}

	return blockDevs, nil
}

func probeMountPoints() ([]dtos.MountPoint, error) {
	var mnt C.struct_mntent
	var buf [4096]C.char
	var ret []dtos.MountPoint

	mTab := C.setmntent(mTabFilePathCString, setmntentFlagsCString)
	if mTab == nil {
		return nil, ErrMTabOpen
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
