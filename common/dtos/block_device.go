package dtos

//Block device represents a block device retrieved by blkid probe
type BlockDevice struct {
	Path string
	UUID string
	Type string
}
