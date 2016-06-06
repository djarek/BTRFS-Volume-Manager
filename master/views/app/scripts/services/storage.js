"use strict"
angular.module("sbAdminApp")
  .service("StorageService", ["$q", "WebsocketService",
  function($q, WebsocketService) {
    this.sendStorageServerListRequest = function() {
      var req = WebsocketService.payloads.NewStorageServerListRequest();
      return WebsocketService.send(req, "StorageServerListResponse");
    }

    this.sendBlockDeviceListRequest = function(serverID) {
      console.log("lol");
      var req = WebsocketService.payloads.NewBlockDeviceListRequest(serverID);
      return WebsocketService.send(req, "BlockDeviceListResponse");
    }
  }])
