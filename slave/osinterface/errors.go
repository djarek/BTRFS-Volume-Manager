package osinterface

import "errors"

var (
	//ErrMTabOpen indicates that the application was not able to open
	//the mounts file.
	ErrMTabOpen = errors.New("Unable to open mtab file " + mTabFilePath)
	//ErrBlkidGetCache occurs when the blkid_get_cache function fails
	ErrBlkidGetCache = errors.New("Unable to retrieve blkid cache /etc/blkid/blkid.tab")
)

//BtrfsCmdError represents an error returned by the btrfs tool
type BtrfsCmdError struct {
	BaseErr string
	Details string
}

func (err BtrfsCmdError) Error() string {
	return "btrfs program error: " + err.BaseErr + "\nDetails: " + err.Details
}
