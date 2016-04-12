angular.module("sbAdminApp")
.service("WebsocketService", ["$rootScope", "$q", "Payloads", function($rootScope, $q, Payloads) {
  var nextRequestId = 0;
  var socket = null;
  var connected = false;
  var sentRequests = {};

  function onMessage(e) {
    var msg = JSON.parse(e.data);
    if (msg.requestID in sentRequests) {
      deferred = sentRequests[msg.requestID];
      delete sentRequests[msg.requestID];
      deferred.resolve(msg);
    } else {
      deferred.reject("Unknown requestID");
    }
  };

  function clearRequests() {
    //TODO: reject request promises and clear sentRequests
  };

  function onClose() {
    connected = false;
    console.log("onClose");
    socket = null;
    clearRequests();
    $rootScope.$broadcast("disconnected", {});
  };

  function onError(error) {
    connected = false;
    console.log("onError");
    socket.close();
    socket = null;
    clearRequests();
    $rootScope.$emit("disconnected", {});
    //TODO: Emit notification about the error, close the connection and try
    //reconnecting.
  };

  var Message = function(payloadObject) {
    this.messageType = payloadObject.getMessageType();
    this.requestID = nextRequestId++;
    this.payload = payloadObject;
  };

  this.send = function(payload) {
    var deferred = $q.defer();
    if (!connected) {
      deferred.reject("Not connected");
      return deferred.promise;
    }

    var msg = new Message(payload);
    sentRequests[msg.requestID] = deferred;
    socket.send(JSON.stringify(msg));
    return deferred.promise;
  }

  this.reconnect = function() {
    if (!connected) {
      socket = new WebSocket("ws://localhost:8080/ws");
      socket.onerror = onError;
      socket.onmessage = onMessage;
      connected = true;
    }
  };

  this.close = function() {
    socket.close(1000, "Going away");
    connected = false;
  };

  this.payloads = Payloads;

  this.reconnect();
}]);

angular.module('sbAdminApp')
.service('Payloads', function() {
  var recvMessageTypes = {
    10000 : "Error",
    10001 : "AuthenticationResponse"
  };

  this.isValid = function(msg, expected) {
    return recvMessageTypes[msg.messageType] === expected;
  }

  this.NewAuthenticationRequest = function(username, password) {
    return {
      username: username,
      password: password,
      getMessageType: function() {
        return 1;
      }
    };
  };
})
