package main

import (
	"log"
	"path/filepath"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
	"github.com/djarek/btrfs-volume-manager/common/router"
	"github.com/djarek/btrfs-volume-manager/slave/osinterface"
)

type blockDevController struct{}

func (b *blockDevController) ExportHandlers(adder router.HandlerAdder) {
	adder.AddHandler(dtos.WSMsgBlockDeviceListRequest, b.onBlockDeviceListRequest)
	adder.AddHandler(dtos.WSMsgBlockDeviceRescanRequest, b.onBlockDeviceRescanRequest)
	adder.AddHandler(dtos.WSMsgBtrfsVolumeListRequest, b.onBtrfsVolumeListRequest)
	adder.AddHandler(dtos.WSMsgBtrfsSubvolumeListRequest, b.onBtrfsSubvolumeListRequest)
}

func filterBlockDevices(blockDevs []dtos.BlockDevice) []dtos.BlockDevice {
	var filtered []dtos.BlockDevice
	for _, bd := range blockDevs {
		if len(bd.Type) > 0 {
			filtered = append(filtered, bd)
		}
	}
	return filtered
}

func (b blockDevController) onBlockDeviceListRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	blockDevs := filterBlockDevices(osinterface.BlockDeviceCache.GetAll())
	response := dtos.NewWebSocketMessage(msg.RequestID, &dtos.BlockDeviceListResponse{BlockDevices: blockDevs})
	ctx.SendAsync(response)
}

func (b blockDevController) onBlockDeviceRescanRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	err := osinterface.BlockDeviceCache.Rescan()
	if err != nil {
		//TODO: send error
		return
	}

	blockDevs := filterBlockDevices(osinterface.BlockDeviceCache.GetAll())

	response := dtos.NewWebSocketMessage(msg.RequestID, &dtos.BlockDeviceRescanResponse{BlockDevices: blockDevs})
	ctx.SendAsync(response)
}

func (b blockDevController) onBtrfsVolumeListRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	vols, err := osinterface.ProbeBtrfsVolumes()
	if err != nil {
		//TODO: send error
		return
	}
	response := dtos.NewWebSocketMessage(msg.RequestID, &dtos.BtrfsVolumeListResponse{BtrfsVolumes: vols})
	ctx.SendAsync(response)
}

func (b blockDevController) onBtrfsSubvolumeListRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	request := msg.Payload.(*dtos.BtrfsSubvolumeListRequest)
	mountPath, err := osinterface.GetBtrfsRootMount(dtos.BtrfsVolume{UUID: request.VolumeUUID})
	if err != nil {
		//TODO: send error
		log.Println(err)
		return
	}
	subvols, err := osinterface.ProbeSubVolumes(mountPath)
	if err != nil {
		log.Println(err)
		//TODO: send error
		return
	}

	response := dtos.NewWebSocketMessage(msg.RequestID, &dtos.BtrfsSubvolumeListResponse{Subvolumes: subvols})
	ctx.SendAsync(response)
}

func (b blockDevController) onBtrfsSubvolumeCreateRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	request := msg.Payload.(*dtos.BtrfsSubvolumeCreateRequest)
	vol := dtos.BtrfsVolume{UUID: request.VolumeUUID}
	mountPath, err := osinterface.GetBtrfsRootMount(vol)
	if err != nil {
		//TODO: send error
		log.Println(err)
		return
	}
	volpath := filepath.Join(mountPath, request.RelativePath)
	err = osinterface.CreateSubVolume(vol, volpath)
	if err != nil {
		log.Println(err)
		return
	}
	response := dtos.NewWebSocketMessage(msg.RequestID, &dtos.BtrfsSubvolumeCreateResponse{})
	ctx.SendAsync(response)
}

func (b blockDevController) onBtrfsSubvolumeDeleteRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	request := msg.Payload.(*dtos.BtrfsSubvolumeDeleteRequest)
	vol := dtos.BtrfsVolume{UUID: request.VolumeUUID}
	mountPath, err := osinterface.GetBtrfsRootMount(vol)
	if err != nil {
		//TODO: send error
		log.Println(err)
		return
	}
	volpath := filepath.Join(mountPath, request.RelativePath)
	err = osinterface.DeleteSubVolume(vol, volpath)
	if err != nil {
		log.Println(err)
		return
	}
	response := dtos.NewWebSocketMessage(msg.RequestID, &dtos.BtrfsSubvolumeDeleteResponse{})
	ctx.SendAsync(response)
}

func (b blockDevController) onBtrfsSubvolumeSnapshotRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	request := msg.Payload.(*dtos.BtrfsSubvolumeSnapshotRequest)

	err := osinterface.CreateSnapshot(dtos.BtrfsSubVolume{
		VolumeUUID:   request.VolumeUUID,
		RelativePath: request.RelativePath,
	}, request.TargetPath)
	if err != nil {
		log.Println(err)
		return
	}
	response := dtos.NewWebSocketMessage(msg.RequestID, &dtos.BtrfsSubvolumeSnapshotResponse{})
	ctx.SendAsync(response)
}
