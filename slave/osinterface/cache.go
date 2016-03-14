package osinterface

import (
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
		mountPointByIdent: make(map[string]*dtos.MountPoint),
	}
)

type mountPointCache struct {
	mtx               sync.RWMutex
	mountPointByIdent map[string]*dtos.MountPoint
	mountPoints       []dtos.MountPoint
}

/*RescanMountPoints performs a scan for all present mount points. If the scan fails,
an appropriate error is returned and the cache is not modified. This function is
thread-safe.
*/
func (mpc *mountPointCache) RescanMountPoints() (err error) {
	mountPoints, err := probeMountPoints()
	if err != nil {
		return
	}

	mountPointByIdent := make(map[string]*dtos.MountPoint)
	for i, mountPoint := range mpc.mountPoints {
		mpc.mountPointByIdent[mountPoint.Identifier] = &mountPoints[i]
	}

	mpc.mtx.Lock()
	defer mpc.mtx.Unlock()

	mpc.mountPoints = mountPoints
	mpc.mountPointByIdent = mountPointByIdent

	return
}

/*FindByKernelIdentifier retrieves a pointer to a mount point of the device using the given
kernel identifier (for example /dev/sda1). The second return value indicates whether the value
was found or not. This function is thread-safe, however, the cache has to be initialized
first by the RescanMountPoints function. The caller is not allowed to modify the object
pointed to by the returned pointer.
*/
func (mpc *mountPointCache) FindByKernelIdentifier(identifier string) (*dtos.MountPoint, bool) {
	mpc.mtx.RLock()
	defer mpc.mtx.RUnlock()
	mp, ok := mpc.mountPointByIdent[identifier]
	return mp, ok
}

type blockDeviceCache struct {
	mtx               sync.RWMutex
	blockDevsByKIdent map[string]*dtos.BlockDevice
	blockDevs         []dtos.BlockDevice
}

/*RescanBlockDevs performs a scan for all present block devices. If the scan fails,
an appropriate error is returned and the cache is not modified. This function is
thread-safe.
*/
func (bdc *blockDeviceCache) RescanBlockDevs() (err error) {
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
