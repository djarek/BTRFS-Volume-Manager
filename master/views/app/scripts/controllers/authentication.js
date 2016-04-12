angular.module("sbAdminApp")
  .controller("AuthCtrl", ["$scope", "$state", "AuthenticationService", function($scope, $state, authService) {
    $scope.onLogoutClick = function() {
      authService.sendLogoutRequest();
      $state.go("login");
    }
    $scope.onLoginClick = function() {
      $scope.errorMessage = null;
      var promise = authService.sendLoginRequest($scope.username, $scope.password);
      promise.then(function(authenticated) {
        if (authenticated) {
          $state.go("dashboard.home");
        } else {
          $scope.errorMessage = "Invalid username or password.";
        }
      }).catch(function(e) {
        $scope.errorMessage = e;
      });
    };
  }]);
