angular.module("sbAdminApp")
.service("WebsocketService", ["$rootScope", "$q", "Payloads",
function($rootScope, $q, payloads) {
  var nextRequestId = 0;
  var socket = null;
  var connected = false;
  var sentRequests = {};
  var onOpenRequests = [];
  const RECONNECT_DELAY = 1000;//milliseconds

  function onMessage(e) {
    var msg = JSON.parse(e.data);
    if (msg.requestID in sentRequests) {
      var deferred = sentRequests[msg.requestID];
      delete sentRequests[msg.requestID];
      deferred.resolve(msg);
    } else {
      deferred.reject("Unknown requestID");
    }
  };

  function onOpen() {
    connected = true;
    onOpenRequests.forEach(function(onOpenCallback) {
      onOpenCallback();
    });
    onOpenRequests = [];
  }

  function clearRequests(reason) {
    for (var requestID in sentRequests) {
      sentRequests[requestID].reject(reason);
    }
    sentRequests = {};
    onOpenRequests = [];
  };

  function reconnect() {
    if (!connected) {
      socket = new WebSocket("ws://localhost:8080/ws");
      socket.onerror = onError;
      socket.onmessage = onMessage;
      socket.onopen = onOpen;
      socket.onclose = onClose;
    }
  };

  function onClose() {
    connected = false;
    socket = null;
    clearRequests("Connection closed");
    window.setTimeout(reconnect, RECONNECT_DELAY);
    $rootScope.$broadcast("disconnected", {});
  };

  function onError(error) {
    connected = false;
    socket.close();
    socket = null;
    clearRequests("Error: " + error);
    window.setTimeout(reconnect, RECONNECT_DELAY);
    $rootScope.$broadcast("disconnected", {});
    //TODO: Emit notification about the error, close the connection and try
    //reconnecting.
  };

  var Message = function(payloadObject) {
    this.messageType = payloadObject.getMessageType();
    this.requestID = nextRequestId++;
    this.payload = payloadObject;
  };

  this.sendRaw = function(msg, requestID) {
    var deferred = $q.defer();
    sentRequests[msg.requestID] = deferred;
    socket.send(msg);
    return deferred.promise;
  }

  function internalSend(payload, expectedResponse) {
    var deferred = $q.defer();
    var msg = new Message(payload);
    if (expectedResponse !== undefined) {
      sentRequests[msg.requestID] = deferred;
    }
    socket.send(JSON.stringify(msg));

    var promise = deferred.promise
    if (expectedResponse !== undefined) {
      promise = promise.then(function(msg) {
        if (expectedResponse !== undefined && !payloads.isValid(msg, expectedResponse)) {
          return q.reject("Invalid received payload type");
        } else {
          return msg;
        }
      });
    } else {
      promise = $q.when();
    }
    return promise;
  };

  this.send = function(payload, expectedResponse) {
    var deferred = $q.defer();
    if (!connected) {
      onOpenRequests.push(function() {
        deferred.resolve(internalSend(payload, expectedResponse));
      });
      return deferred.promise;
    } else {
      return internalSend(payload, expectedResponse);
    }
  };

  this.reconnect = reconnect;

  this.close = function() {
    socket.close(1000, "Going away");
    connected = false;
  };

  this.payloads = payloads;

  this.reconnect();
}]);

angular.module('sbAdminApp')
.service('Payloads', function() {
  var recvMessageTypes = {
    10000 : "Error",
    10001 : "AuthenticationResponse",
    10005 : "BlockDeviceRescanResponse",
    10006 : "StorageServerListResponse",
    10007 : "BlockDeviceListResponse",
    10008 : "BtrfsVolumeListResponse",
    10009 : "BtrfsSubvolumeListResponse",
    10010 : "BtrfsSubvolumeCreateResponse",
    10011 : "BtrfsSubvolumeDeleteResponse",
    10012 : "BtrfsSubvolumeSnapshotResponse",
  };

  this.isValid = function(msg, expected) {
    return recvMessageTypes[msg.messageType] === expected || expected === "Any";
  }

  this.NewAuthenticationRequest = function(username, password) {
    return {
      username: username,
      password: password,
      getMessageType: function() { return 1; }
    };
  };

  this.NewLogoutRequest = function() {
    return {
      getMessageType: function() { return 2; }
    };
  };

  this.NewReauthenticatonRequest = function(token) {
    return {
      token: token,
      getMessageType: function() { return 3; }
    };
  };

  this.NewBlockDeviceRescanRequest = function(serverID) {
    return {
      serverID: serverID,
      getMessageType: function() { return 5; }
    };
  };

  this.NewStorageServerListRequest = function(serverID) {
    return {
      serverID: serverID,
      getMessageType: function() { return 6; }
    };
  };

  this.NewBlockDeviceListRequest = function(serverID) {
    return {
      serverID: serverID,
      getMessageType: function() { return 7; }
    };
  };

  this.NewBtrfsVolumeListRequest = function(serverID) {
    return {
      serverID: serverID,
      getMessageType: function() { return 8; }
    };
  };

  this.NewBtrfsSubvolumeListRequest = function(serverID, volumeUUID) {
    return {
      serverID: serverID,
      volumeUUID: volumeUUID,
      getMessageType: function() { return 9; }
    };
  };

  this.NewBtrfsSubvolumeCreateRequest = function(serverID, volumeUUID, relativePath) {
    return {
      serverID: serverID,
      volumeUUID: volumeUUID,
      relativePath: relativePath,
      getMessageType: function() { return 10; }
    };
  };

  this.NewBtrfsSubvolumeDeleteRequest = function(serverID, volumeUUID, relativePath) {
    return {
      serverID: serverID,
      volumeUUID: volumeUUID,
      relativePath: relativePath,
      getMessageType: function() { return 11; }
    };
  };

  this.NewBtrfsSubvolumeSnapshotRequest = function(serverID, volumeUUID, relativePath, targetPath) {
    return {
      serverID: serverID,
      volumeUUID: volumeUUID,
      relativePath: relativePath,
      targetPath: targetPath,
      getMessageType: function() { return 12; }
    };
  };
})
