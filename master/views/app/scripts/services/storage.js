"use strict"
angular.module("sbAdminApp")
  .service("StorageService", ["$q", "WebsocketService",
  function($q, WebsocketService) {
    this.sendStorageServerListRequest = function() {
      var req = WebsocketService.payloads.NewStorageServerListRequest();
      return WebsocketService.send(req, "StorageServerListResponse");
    };

    this.sendBlockDeviceListRequest = function(serverID) {
      var req = WebsocketService.payloads.NewBlockDeviceListRequest(serverID);
      return WebsocketService.send(req, "BlockDeviceListResponse");
    };

    this.sendBlockDeviceRescanRequest = function(serverID) {
      var req = WebsocketService.payloads.NewBlockDeviceRescanRequest(serverID);
      return WebsocketService.send(req, "BlockDeviceRescanResponse");
    };

    this.sendBtrfsVolumeListRequest = function(serverID) {
      var req = WebsocketService.payloads.NewBtrfsVolumeListRequest(serverID);
      return WebsocketService.send(req, "BtrfsVolumeListResponse");
    };

    this.sendBtrfsSubvolumeListRequest = function(serverID, volumeUUID) {
      var req = WebsocketService.payloads.NewBtrfsSubvolumeListRequest(serverID, volumeUUID);
      return WebsocketService.send(req, "BtrfsSubvolumeListResponse");
    };

    this.sendBtrfsSubvolumeCreateRequest = function(serverID, volumeUUID, relativePath) {
      var req = WebsocketService.payloads.NewBtrfsSubvolumeListRequest(serverID, volumeUUID, relativePath);
      return WebsocketService.send(req, "BtrfsSubvolumeCreateResponse");
    };

    this.sendBtrfsSubvolumeDeleteRequest = function(serverID, volumeUUID, relativePath) {
      var req = WebsocketService.payloads.NewBtrfsSubvolumeListRequest(serverID, volumeUUID, relativePath);
      return WebsocketService.send(req, "BtrfsSubvolumeDeleteResponse");
    };

    this.sendBtrfsSubvolumeSnapshotRequest = function(serverID, volumeUUID, relativePath, targetPath) {
      var req = WebsocketService.payloads.NewBtrfsSubvolumeListRequest(serverID, volumeUUID, relativePath, targetPath);
      return WebsocketService.send(req, "BtrfsSubvolumeSnapshotResponse");
    };
  }])
