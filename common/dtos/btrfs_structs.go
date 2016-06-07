package dtos

//BlockDevID represents the identifier field of a BlockDevice entry
type BlockDevID int32

//StorageServerID represents the identifier field of a StorageServer entry
type StorageServerID int32

//VolumeID represents the identifier field of a BtrfsVolume entry
type VolumeID int32

//UUIDType is the string that contains the UUID of a filesystem entity
type UUIDType string

//BlockDevice represents a block device retrieved by blkid probe
type BlockDevice struct {
	ID       BlockDevID      `json:"-"`
	VolID    VolumeID        `json:"-"`
	ServerID StorageServerID `json:"-"`
	Path     string          `json:"path"`
	UUID     UUIDType        `json:"UUID"`
	Type     string          `json:"type"`
}

//StorageServer represents a Network Attached Storage device
type StorageServer struct {
	ID           StorageServerID `json:"id"`
	Name         string          `json:"name"`
	SlaveVersion string          `json:"slaveVersion"`
	OSVersion    string          `json:"osVersion"`
}

//BtrfsVolume represents a filesystem volume which can potentially span over
//multiple devices
type BtrfsVolume struct {
	ID       VolumeID        `json:"-"`
	ServerID StorageServerID `json:"-"`
	UUID     UUIDType        `json:"UUID"`
	Label    string          `json:"label"`
	Devices  []*BlockDevice  `json:"devices"`
}

//MountPoint describes a filesystem mount directory and options
type MountPoint struct {
	Identifier    string
	MountPath     string
	MountType     string
	MountOptions  string
	DumpFrequency int
	FSCKPassNo    int
}

//BtrfsSubVolume represents a subvolume on a btrfs volume
type BtrfsSubVolume struct {
	SubVolID     int
	RelativePath string `json:"relativePath"`
	VolumeUUID   UUIDType
	ParentUUID   UUIDType
}
