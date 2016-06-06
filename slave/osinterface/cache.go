package osinterface

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
)

var (
	/*BlockDeviceCache contains the block device cache instance.
	  The object is always in a valid state, however, it will be empty
	  Before RescanBlockDevs is called.
	*/
	BlockDeviceCache = blockDeviceCache{
		blockDevsByKIdent: make(map[string]*dtos.BlockDevice),
	}

	/*MountPointCache contains the block device cache instance.
	  The object is always in a valid state, however, it will be empty
	  Before RescanMountPoints is called.
	*/
	MountPointCache = mountPointCache{
		mountPointByIdent: make(map[string]([]*dtos.MountPoint)),
		btrfsRootMounts:   make(map[dtos.UUIDType]*dtos.MountPoint),
	}
)

type mountPointCache struct {
	mtx               sync.RWMutex
	mountPointByIdent map[string][]*dtos.MountPoint
	mountPoints       []dtos.MountPoint
	btrfsRootMounts   map[dtos.UUIDType]*dtos.MountPoint
}

func (mpc *mountPointCache) FindRootMount(UUID dtos.UUIDType) (*dtos.MountPoint, bool) {
	mpc.mtx.RLock()
	defer mpc.mtx.RUnlock()

	mountPoint, ok := mpc.btrfsRootMounts[UUID]
	return mountPoint, ok
}

/*Rescan performs a scan for all present mount points. If the scan fails,
an appropriate error is returned and the cache is not modified. This function is
thread-safe.
*/
func (mpc *mountPointCache) Rescan() (err error) {
	mountPoints, err := probeMountPoints()
	if err != nil {
		return
	}

	mountPointByIdent := make(map[string][]*dtos.MountPoint)
	btrfsRootMounts := make(map[dtos.UUIDType]*dtos.MountPoint)
	for i, mountPoint := range mountPoints {
		mountPointByIdent[mountPoint.Identifier] = append(
			mountPointByIdent[mountPoint.Identifier],
			&mountPoints[i],
		)
		if rootMountsPath == filepath.Dir(mountPoint.MountPath) {
			bd, ok := BlockDeviceCache.FindByKernelIdentifier(mountPoint.Identifier)
			if ok {
				btrfsRootMounts[bd.UUID] = &mountPoints[i]
			} else {
				log.Println("Warning: unable to find block device for mount point: " +
					mountPoint.MountPath)
			}
		}
	}

	mpc.mtx.Lock()
	defer mpc.mtx.Unlock()

	mpc.btrfsRootMounts = btrfsRootMounts
	mpc.mountPoints = mountPoints
	mpc.mountPointByIdent = mountPointByIdent
	return
}

/*FindByKernelIdentifier retrieves a slice of mount points of the device using the given
kernel identifier (for example /dev/sda1). The second return value indicates whether the value
was found or not. This function is thread-safe, however, the cache has to be initialized
first by the RescanMountPoints function. The caller is not allowed to modify the object
pointed to by the returned pointer.
*/
func (mpc *mountPointCache) FindByKernelIdentifier(identifier string) ([]*dtos.MountPoint, bool) {
	mpc.mtx.RLock()
	defer mpc.mtx.RUnlock()
	mp, ok := mpc.mountPointByIdent[identifier]
	return mp, ok
}

type blockDeviceCache struct {
	mtx               sync.RWMutex
	blockDevsByKIdent map[string]*dtos.BlockDevice
	blockDevsByUUID   map[dtos.UUIDType][]*dtos.BlockDevice
	blockDevs         []dtos.BlockDevice
}

/*Rescan performs a scan for all present block devices. If the scan fails,
an appropriate error is returned and the cache is not modified. This function is
thread-safe.
*/
func (bdc *blockDeviceCache) Rescan() (err error) {
	blockDevs, err := probeBlockDevices()
	if err != nil {
		return
	}

	blockDevsByKIdent := make(map[string]*dtos.BlockDevice)
	blockDevsByUUID := make(map[dtos.UUIDType][]*dtos.BlockDevice)
	for i, blockDev := range blockDevs {
		blockDevsByKIdent[blockDev.Path] = &blockDevs[i]
		blockDevsByUUID[blockDev.UUID] = append(
			blockDevsByUUID[blockDev.UUID],
			&blockDevs[i])
	}

	bdc.mtx.Lock()
	defer bdc.mtx.Unlock()

	bdc.blockDevs = blockDevs
	bdc.blockDevsByKIdent = blockDevsByKIdent
	bdc.blockDevsByUUID = blockDevsByUUID
	return
}

/*FindByKernelIdentifier retrieves a pointer to a BlockDevice using the given
kernel identifier (for example /dev/sda1). The second return value indicates whether the
value was found or not. This function is thread-safe, however, the cache has to be
initialized first by the RescanBlockDevs function. The caller is not allowed to
modify the object pointed to by the returned pointer.
*/
func (bdc *blockDeviceCache) FindByKernelIdentifier(identifier string) (*dtos.BlockDevice, bool) {
	bdc.mtx.RLock()
	defer bdc.mtx.RUnlock()
	bd, ok := bdc.blockDevsByKIdent[identifier]
	return bd, ok
}

func (bdc *blockDeviceCache) FindByUUID(UUID dtos.UUIDType) ([]*dtos.BlockDevice, bool) {
	bdc.mtx.RLock()
	defer bdc.mtx.RUnlock()
	bd, ok := bdc.blockDevsByUUID[UUID]
	return bd, ok
}

/*GetAll returns the cached list of all present block devices.*/
func (bdc *blockDeviceCache) GetAll() []dtos.BlockDevice {
	bdc.mtx.RLock()
	defer bdc.mtx.RUnlock()
	var ret []dtos.BlockDevice
	for _, blockDev := range bdc.blockDevs {
		ret = append(ret, blockDev)
	}
	return ret
}
