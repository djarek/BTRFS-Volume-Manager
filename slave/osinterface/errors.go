package osinterface

import "errors"
import "C"

var (
	//ErrorMTabOpen indicates that the application was not able to open
	//the mounts file.
	ErrorMTabOpen = errors.New("Unable to open mtab file " + mTabFilePath)
)

//BtrfsCmdError represents an error returned by the btrfs tool
type BtrfsCmdError struct {
	BaseErr string
	Details string
}

func (err BtrfsCmdError) Error() string {
	return "btrfs program error: " + err.BaseErr + "\nDetails" + err.Details
}
