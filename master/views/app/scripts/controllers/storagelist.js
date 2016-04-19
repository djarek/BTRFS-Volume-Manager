angular.module("sbAdminApp")
  .controller("StorageListCtrl", ["$scope", "StorageService",
  function($scope, StorageService) {
    StorageService.sendStorageServerListRequest().then(function(msg) {
      $scope.servers = msg.payload.servers;
    }).catch(function(err) {
      $scope.errorMsg = err;
    });
  }])
