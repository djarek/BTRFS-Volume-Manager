package blockdevices

import (
	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
	"github.com/djarek/btrfs-volume-manager/common/router"
	"github.com/djarek/btrfs-volume-manager/master/storageservers"
)

type controller struct {
	serverTracker storageservers.Tracker
}

/*NewController constructs a new valid controller*/
func NewController(tracker storageservers.Tracker) router.HandlerExporter {
	return &controller{
		serverTracker: tracker,
	}
}

func (c *controller) ExportHandlers(adder router.HandlerAdder) {
	adder.AddHandler(dtos.WSMsgBlockDeviceRescanRequest, c.onBlockDeviceRescanRequest)
	adder.AddHandler(dtos.WSMsgBlockDeviceRescanResponse, router.DefaultResponseHandler)

	adder.AddHandler(dtos.WSMsgBlockDeviceListRequest, c.onBlockDeviceListRequest)
	adder.AddHandler(dtos.WSMsgBlockDeviceListResponse, router.DefaultResponseHandler)

	adder.AddHandler(dtos.WSMsgBtrfsVolumeListRequest, c.onBtrfsListRequest)
	adder.AddHandler(dtos.WSMsgBtrfsVolumeListResponse, router.DefaultResponseHandler)

	adder.AddHandler(dtos.WSMsgBtrfsSubvolumeListRequest, c.onBtrfsSubvolumeListRequest)
	adder.AddHandler(dtos.WSMsgBtrfsSubvolumeListResponse, router.DefaultResponseHandler)

	adder.AddHandler(dtos.WSMsgBtrfsSubvolumeCreateRequest, c.ForwardToSlave)
	adder.AddHandler(dtos.WSMsgBtrfsSubvolumeCreateResponse, router.DefaultResponseHandler)

	adder.AddHandler(dtos.WSMsgBtrfsSubvolumeDeleteRequest, c.ForwardToSlave)
	adder.AddHandler(dtos.WSMsgBtrfsSubvolumeDeleteResponse, router.DefaultResponseHandler)

	adder.AddHandler(dtos.WSMsgBtrfsSubvolumeSnapshotRequest, c.ForwardToSlave)
	adder.AddHandler(dtos.WSMsgBtrfsSubvolumeSnapshotResponse, router.DefaultResponseHandler)
}

type serverVolumeGetter interface {
	GetServerID() dtos.StorageServerID
	GetVolumeUUID() dtos.UUIDType
}

func (c *controller) ForwardToSlave(ctx *request.Context, msg dtos.WebSocketMessage) {
	servVolGetter := msg.Payload.(serverVolumeGetter)
	storageServCtx, ok := c.serverTracker.GetServerContext(servVolGetter.GetServerID())
	if !ok {
		//TODO: unknown storage server, send error
		return
	}

	clientRequestID := msg.RequestID
	requestID, responseChannel := storageServCtx.NewRequest()
	msg.RequestID = requestID
	storageServCtx.SendAsync(msg)
	go func() {
		response, ok := <-responseChannel
		if ok {
			response.RequestID = clientRequestID
			ctx.SendAsync(response)
		}
	}()
}

func (c *controller) onBlockDeviceListRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	blockDevListRequest := msg.Payload.(*dtos.BlockDeviceListRequest)
	storageServCtx, ok := c.serverTracker.GetServerContext(blockDevListRequest.ServerID)
	if !ok {
		//TODO: unknown storage server, send error
		return
	}

	clientRequestID := msg.RequestID
	requestID, responseChannel := storageServCtx.NewRequest()
	msg.RequestID = requestID
	storageServCtx.SendAsync(msg)
	go func() {
		response, ok := <-responseChannel
		if ok {
			response.RequestID = clientRequestID
			ctx.SendAsync(response)
		}
	}()
}

func (c *controller) onBlockDeviceRescanRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	rescanRequest := msg.Payload.(*dtos.BlockDeviceRescanRequest)
	storageServCtx, ok := c.serverTracker.GetServerContext(rescanRequest.ServerID)
	if !ok {
		//TODO: unknown storage server, send error
		return
	}
	clientRequestID := msg.RequestID
	requestID, responseChannel := storageServCtx.NewRequest()
	msg.RequestID = requestID
	storageServCtx.SendAsync(msg)

	go func() {
		response, ok := <-responseChannel
		if ok {
			response.RequestID = clientRequestID
			ctx.SendAsync(response)
		}
	}()
}

func (c *controller) onBtrfsSubvolumeListRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	listRequest := msg.Payload.(*dtos.BtrfsSubvolumeListRequest)
	storageServCtx, ok := c.serverTracker.GetServerContext(listRequest.ServerID)
	if !ok {
		//TODO: unknown storage server, send error
		return
	}
	clientRequestID := msg.RequestID
	requestID, responseChannel := storageServCtx.NewRequest()
	msg.RequestID = requestID
	storageServCtx.SendAsync(msg)

	go func() {
		response, ok := <-responseChannel
		if ok {
			response.RequestID = clientRequestID
			ctx.SendAsync(response)
		}
	}()
}

func (c *controller) onBtrfsListRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	listRequest := msg.Payload.(*dtos.BtrfsVolumeListRequest)
	storageServCtx, ok := c.serverTracker.GetServerContext(listRequest.ServerID)
	if !ok {
		//TODO: unknown storage server, send error
		return
	}
	clientRequestID := msg.RequestID
	requestID, responseChannel := storageServCtx.NewRequest()
	msg.RequestID = requestID
	storageServCtx.SendAsync(msg)

	go func() {
		response, ok := <-responseChannel
		if ok {
			response.RequestID = clientRequestID
			ctx.SendAsync(response)
		}
	}()
}
