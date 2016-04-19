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

	"github.com/djarek/btrfs-volume-manager/common/dtos"
)

const (
	mTabFilePath   = "/proc/mounts"
	setmntentFlags = "r"
)

var (
	mTabFilePathCString     = C.CString(mTabFilePath)
	setmntentFlagsCString   = C.CString(setmntentFlags)
	blkidUUIDTagNameCString = C.CString("UUID")
	blkidTypeTagNameCString = C.CString("TYPE")
)

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

var devMatcher = regexp.MustCompile("path (\\/dev\\/[a-zA-Z0-9\\/_]+)")

/*ProbeBtrfsVolumes retrieves the list of all btrfs volumes present on this
server.*/
func ProbeBtrfsVolumes() (vols []dtos.BtrfsVolume, err error) {
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
			dev, ok := BlockDeviceCache.FindByKernelIdentifier(devMatch[1])
			if ok {
				volume.Devices = append(volume.Devices, dev)
			}
		}
		vols = append(vols, volume)
	}
	return
}
