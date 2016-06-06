package main

import (
	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
	"github.com/djarek/btrfs-volume-manager/common/router"
	"github.com/djarek/btrfs-volume-manager/slave/osinterface"
)

type blockDevController struct{}

func (b *blockDevController) ExportHandlers(adder router.HandlerAdder) {
	adder.AddHandler(dtos.WSMsgBlockDeviceListRequest, b.onBlockDeviceListRequest)
	adder.AddHandler(dtos.WSMsgBlockDeviceRescanRequest, b.onBlockDeviceRescanRequest)
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
