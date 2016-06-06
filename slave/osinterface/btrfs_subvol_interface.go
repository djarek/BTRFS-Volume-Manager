package osinterface

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
)

const (
	btrfsCmd       = "btrfs"
	rootMountsPath = "/mnt"
)

var runBtrfsCommand = func(options ...string) (outputString string, err error) {
	cmd := exec.Command(btrfsCmd, options...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		//The btrfs tool returned an error or it was not found in the OS
		err = BtrfsCmdError{
			BaseErr: err.Error(),
			Details: stderr.String(),
		}
		return
	}

	outputString = string(output)
	return
}

/*ProbeSubVolumes probes the kernel and retrieves all subvolumes present
in a btrfs volume. The mountPath is the path to any mount point of a volume
(or a path below the mount point).
*/
func ProbeSubVolumes(mountPath string) (subvols []dtos.BtrfsSubVolume, err error) {
	options := "-tqu"
	output, err := runBtrfsCommand("subvolume", "list", mountPath, options)
	if err != nil {
		return
	}

	var subvol dtos.BtrfsSubVolume
	var gen, topLevel int

	//Get lines and skip the first two, because they contain the list header:
	//ID      gen     top level       parent_uuid     uuid    path
	//--      ---     ---------       -----------     ----    ----
	lines := strings.Split(output, "\n")
	lines = lines[2:]
	for _, line := range lines {
		_, err = fmt.Sscanf(line, "%d %d %d %s %s %s",
			&subvol.SubVolID,
			&gen,      //unused by us
			&topLevel, //unused by us
			&subvol.ParentUUID,
			&subvol.VolumeUUID,
			&subvol.RelativePath,
		)
		if err != nil {
			if err != io.EOF {
				return
			}
			err = nil
			return
		}
		subvols = append(subvols, subvol)
	}
	return
}

func runBtrfsSubvolumeCommand(vol dtos.BtrfsVolume, subvolRelativePath string, subCommand string) error {
	mountPath, err := getBtrfsRootMount(vol)
	if err != nil {
		return err
	}
	path := filepath.Join(mountPath, subvolRelativePath)
	_, err = runBtrfsCommand("subvolume", subCommand, path)
	if err != nil {
		return err
	}
	return nil
}

/*CreateSubVolume attempts to create a subvolume at the specified path
(relative to the volume root). If the volume's root cannot be mounted
this function returns an error.*/
func CreateSubVolume(vol dtos.BtrfsVolume, subvolRelativePath string) error {
	err := runBtrfsSubvolumeCommand(vol, subvolRelativePath, "create")
	if err != nil {
		return errors.New("Unable to create subvolume: " + err.Error())
	}
	return nil
}

/*DeleteSubVolume attempts to create a subvolume at the specified path
(relative to the volume root). If the volume's root cannot be mounted
this function returns an error.*/
func DeleteSubVolume(vol dtos.BtrfsVolume, subvolRelativePath string) error {
	err := runBtrfsSubvolumeCommand(vol, subvolRelativePath, "delete")
	if err != nil {
		return errors.New("Unable to delete subvolume: " + err.Error())
	}
	return nil
}

/*CreateSnapshot attempts to create a snapshot of a subvolume at the specified
path (relative to the volume root). If the volume's root is not mounted this
function returns an error.*/
func CreateSnapshot(subvol dtos.BtrfsSubVolume, snapshotRelativePath string) error {
	mountPath, err := getBtrfsRootMount(dtos.BtrfsVolume{UUID: subvol.VolumeUUID})
	if err != nil {
		return err
	}

	path := filepath.Join(mountPath, snapshotRelativePath)
	sourcePath := filepath.Join(mountPath, subvol.RelativePath)
	_, err = runBtrfsCommand("subvolume", "snapshot", sourcePath, path)
	return err
}
