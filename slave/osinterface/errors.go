package osinterface

import "errors"
import "C"

var (
	ErrorMTabOpen = errors.New("Unable to open mtab file " + mTabFilePath)
)
