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
