package storageservers

import (
	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
	"github.com/djarek/btrfs-volume-manager/common/router"
)

const (
	serverDetailsKey = "StorageServerDetails"
)

type storageServerDetails struct {
	ID           dtos.StorageServerID
	name         string
	os           string
	slaveVersion string
}

type controller struct {
	tracker Tracker
}

func (c *controller) ExportHandlers(adder router.HandlerAdder) {
	adder.AddHandler(dtos.WSMsgStorageServerRegistrationRequest, c.onServerRegistrationRequest)
	adder.AddHandler(dtos.WSMsgStorageServerListRequest, c.onServerListRequest)
	adder.AddOnCloseHandler(c.onServerConnectionClose)
}

/*NewController constructs a new valid controller*/
func NewController(t Tracker) router.HandlerExporter {
	return &controller{tracker: t}
}

func (c *controller) onServerRegistrationRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	request := msg.Payload.(*dtos.StorageServerRegistrationRequest)
	ID := c.tracker.RegisterServer(ctx)
	details := storageServerDetails{
		ID:   ID,
		name: request.ServerName,
	}
	ctx.SetSessionData(serverDetailsKey, details)
	responsePayload := &dtos.StorageServerRegistrationResponse{
		AssignedID: ID,
	}
	responseMsg := dtos.NewWebSocketMessage(msg.RequestID, responsePayload)
	ctx.SendAsync(responseMsg)
}

func (c *controller) onServerConnectionClose(ctx *request.Context, msg dtos.WebSocketMessage) {
	detailsInterface, found := ctx.GetSessionData(serverDetailsKey)
	if found {
		c.tracker.RemoveServer(detailsInterface.(storageServerDetails).ID)
	}
}

func (c *controller) onServerListRequest(ctx *request.Context, msg dtos.WebSocketMessage) {
	serverList := c.tracker.GetAllServers()
	var storageServers []dtos.StorageServer
	for _, storageCtx := range serverList {
		detailsInterface, _ := storageCtx.GetSessionData(serverDetailsKey)
		details := detailsInterface.(storageServerDetails)
		serv := dtos.StorageServer{
			ID:           details.ID,
			Name:         details.name, //TODO: replace placeholders
			SlaveVersion: "0.0.1_placeholder",
			OSVersion:    "Ubuntu 16.04_placeholder",
		}
		storageServers = append(storageServers, serv)
	}
	respList := &dtos.StorageServerListResponse{
		Servers: storageServers,
	}
	respMsg := dtos.NewWebSocketMessage(msg.RequestID, respList)
	ctx.SendAsync(respMsg)
}
