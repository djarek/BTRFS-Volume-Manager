package osinterface

/*
#cgo LDFLAGS: -lbtrfs -lblkid
#include "btrfs.h"
#include <mntent.h>
*/
import "C"
import (
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
)

/*ProbeBlockDevices retrieves all block devices present on
the device and returns a slice of dtos.BlockDevice structs that
represent them.
*/
func ProbeBlockDevices() []dtos.BlockDevice {
	var devArray C.struct_block_devices_array
	if C.get_devices(&devArray) != 0 {
		return nil
	}
	defer C.block_devices_array_free(devArray)

	devArraySlice := (*[1 << 30]C.struct_block_device)(unsafe.Pointer(devArray.devs))[:devArray.count:devArray.count]

	ret := make([]dtos.BlockDevice, devArray.count)
	for i, dev := range devArraySlice {
		ret[i].Path = C.GoString(dev.dev_name)
		ret[i].UUID = dtos.UUIDType(C.GoString(dev.UUID))
		ret[i].Type = C.GoString(dev._type)
	}

	return ret
}

/*ProbeMountPoints retrieves all mounted filesystems and mount
information and stores this data in a slice of dtos.MountPoint.
*/
func ProbeMountPoints() ([]dtos.MountPoint, error) {
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

func getBtrfsMountPointsMap(mountPoints []dtos.MountPoint) map[string]*dtos.MountPoint {
	mountPointsMap := make(map[string]*dtos.MountPoint)
	for i, mountPoint := range mountPoints {
		if mountPoint.MountType != btrfsDevType {
			continue
		}

		mountPointsMap[mountPoint.Identifier] = &mountPoints[i]
	}
	return mountPointsMap
}

func filterBtrfsVolumeMountPoints(mountPointsMap map[string]*dtos.MountPoint, vol dtos.BtrfsVolume) (mountPoints []dtos.MountPoint) {
	for _, dev := range vol.PresentDevs {
		mountPoint, ok := mountPointsMap[dev.Path]
		if ok {
			mountPoints = append(mountPoints, *mountPoint)
		}
	}
	//TODO: Handle UUID mount point identifiers
	return
}

/*ProbeBtrfsVolumes retrieves information about all Btrfs volumes present on
the system and stores them in a slice of dtos.BtrfsVolumes.
*/
func ProbeBtrfsVolumes(devs []dtos.BlockDevice) (vols []dtos.BtrfsVolume, err error) {
	blockDevMap := make(map[dtos.UUIDType][]dtos.BlockDevice)

	for _, dev := range devs {
		if dev.Type != btrfsDevType {
			continue
		}
		devSlice := blockDevMap[dev.UUID]
		blockDevMap[dev.UUID] = append(devSlice, dev)
	}

	mountPoints, err := ProbeMountPoints()
	if err != nil {
		return
	}
	mountPointsMap := getBtrfsMountPointsMap(mountPoints)

	for UUID, volDevs := range blockDevMap {
		vol := dtos.BtrfsVolume{
			UUID:        UUID,
			PresentDevs: volDevs,
		}
		vol.MountPoints = filterBtrfsVolumeMountPoints(mountPointsMap, vol)

		vols = append(vols, vol)
	}

	return
}
