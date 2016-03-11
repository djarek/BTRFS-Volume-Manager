package dtos

//BlockDevID represents the identifier field of a BlockDevice entry
type BlockDevID int32

//StorageServerID represents the identifier field of a StorageServer entry
type StorageServerID int32

//VolumeID represents the identifier field of a BtrfsVolume entry
type VolumeID int32

//BlockDevice represents a block device retrieved by blkid probe
type BlockDevice struct {
	ID       BlockDevID
	VolID    VolumeID
	ServerID StorageServerID
	Path     string
	UUID     string
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
	Label    string
}