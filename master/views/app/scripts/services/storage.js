"use strict"
angular.module("sbAdminApp")
  .service("StorageService", ["$q", "WebsocketService",
  function($q, WebsocketService) {
    this.sendStorageServerListRequest = function() {
      var req = WebsocketService.payloads.NewStorageServerListRequest();
      return WebsocketService.send(req, "StorageServerListResponse")
    }
  }])
