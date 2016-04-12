angular.module("sbAdminApp")
  .service("AuthenticationService", ["$rootScope", "$q", "WebsocketService", function($rootScope, $q, wsService) {
    var userDetails = null;
    $rootScope.$on("disconnected", function(event, data) {
      userDetails = null;
      console.log("disconnected");
    })
    this.getUserDetails = function() {
      return userDetails;
    }

    this.sendLogoutRequest = function() {
      //TODO: Send logout request
    }

    this.sendLoginRequest = function(username, password) {
      var authReq = wsService.payloads.NewAuthenticationRequest(
        username,
        password
      );
      responsePromise = wsService.send(authReq);
      return responsePromise.then(function(msg) {
        if (wsService.payloads.isValid(msg, "AuthenticationResponse")) {
          if (msg.payload.result === "auth_ok") {
            userDetails = msg.payload.userDetails;
            return true;
          } else {
            return false;
          }
        } else {
          return $q.reject("Invalid received payload type");
        }
      });
    };
  }]);
