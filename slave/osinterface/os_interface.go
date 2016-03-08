package osinterface

/*
#cgo CFLAGS: -I ../../libbtrfs/
#cgo LDFLAGS: -L../../libbtrfs/build -lbtrfs -lblkid
#include "btrfs.h"
*/
import "C"
import (
	"unsafe"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
)

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
		ret[i].UUID = C.GoString(dev.UUID)
		ret[i].Type = C.GoString(dev._type)
	}

	return ret
}
