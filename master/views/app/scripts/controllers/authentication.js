angular.module("sbAdminApp")
  .controller("AuthCtrl", ["$rootScope", "$scope", "$state", "AuthenticationService",
  function($rootScope, $scope, $state, authService) {
    $scope.onLogoutClick = function() {
      authService.sendLogoutRequest().then(function(response) {
        $state.go("login");
      });
    }

    function navigateAfterLogin($rootScope) {
      if ($rootScope.toState) {
        $state.go($rootScope.toState, $rootScope.toStateParams);
        delete $rootScope.toState;
        delete $rootScope.toStateParams;
      } else {
        $state.go("dashboard.home");
      }
    }

    $scope.onLoginClick = function() {
      $scope.errorMessage = null;
      if (authService.isAuthRequestSent()) {
        return;
      }
      var promise = authService.sendLoginRequest($scope.username, $scope.password);
      promise.then(function(authenticated) {
        if (authenticated) {
          navigateAfterLogin($rootScope);
        } else {
          $scope.errorMessage = "Invalid username or password.";
        }
      }).catch(function(e) {
        $scope.errorMessage = e;
      });
    };
  }]);
