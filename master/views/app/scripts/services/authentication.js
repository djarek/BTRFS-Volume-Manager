"use strict"
angular.module("sbAdminApp")
  .service("AuthenticationService", ["$rootScope", "$q", "$cookies", "WebsocketService",
  function($rootScope, $q, $cookies, wsService) {
    const SESSION_COOKIE_NAME = "btrfs-volume-manager-session-cookie";
    var userDetails = null;
    var authRequestSent = false;
    $rootScope.$on("disconnected", function(event, data) {
      userDetails = null;
      authRequestSent = false;
    })

    this.getUserDetails = function() {
      return userDetails;
    }

    this.sendLogoutRequest = function() {
      var logoutRequest = wsService.payloads.NewLogoutRequest();
      userDetails = null;
      $cookies.remove(SESSION_COOKIE_NAME);
      return wsService.send(logoutRequest);
    }

    this.isAuthRequestSent = function() {
      return authRequestSent;
    }

    this.isAuthenticated = function() {
      return userDetails != null;
    }

    this.hasSessionCookie = function() {
      return $cookies.getObject(SESSION_COOKIE_NAME) != null
    }

    this.sendReloginRequest = function() {
      var storedDetails = $cookies.getObject(SESSION_COOKIE_NAME);
      var reauthReq = wsService.payloads.NewReauthenticatonRequest(
        storedDetails//.sessionToken
      )
      authRequestSent = true;
      return wsService.send(reauthReq, "AuthenticationResponse").then(function(msg) {
        if (msg.payload.result === "auth_ok") {
          userDetails = storedDetails;
          return true;
        } else {
          authRequestSent = false;
          return false;
        }
      });
    };

    this.sendLoginRequest = function(username, password) {
      var authReq = wsService.payloads.NewAuthenticationRequest(
        username,
        password
      );
      return wsService.send(authReq, "AuthenticationResponse").then(function(msg) {
        if (msg.payload.result === "auth_ok") {
          userDetails = msg.payload.userDetails;
          $cookies.putObject(SESSION_COOKIE_NAME, userDetails);
          return true;
        } else {
          authRequestSent = false;
          return false;
        }
      });
    };
  }]);
