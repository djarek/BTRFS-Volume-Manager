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
	ID       BlockDevID
	VolID    VolumeID
	ServerID StorageServerID
	Path     string
	UUID     UUIDType
	Type     string
}

//StorageServer represents a Network Attached Storage device
type StorageServer struct {
	ID   StorageServerID
	Name string
}

//BtrfsVolume represents a filesystem volume which can potentially span over
//multiple devices
type BtrfsVolume struct {
	ID       VolumeID
	ServerID StorageServerID
	UUID     UUIDType
	Label    string
	Devices  []*BlockDevice
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
	RelativePath string
	VolumeUUID   UUIDType
	ParentUUID   UUIDType
}
