package main

import (
	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
	"github.com/djarek/btrfs-volume-manager/common/router"
	"github.com/djarek/btrfs-volume-manager/slave/osinterface"
)

type blockDevController struct{}

func (b *blockDevController) ExportHandlers(adder router.HandlerAdder) {
	adder.AddHandler(dtos.WSMsgBlockDeviceListRequest, b.onBlockDeviceRescanRequest)
}

func (b blockDevController) onBlockDeviceRescanRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	blockDevs := osinterface.BlockDeviceCache.GetAll()
	response := dtos.NewWebSocketMessage(msg.RequestID, &dtos.BlockDeviceListResponse{BlockDevices: blockDevs})
	ctx.SendAsync(response)
}
