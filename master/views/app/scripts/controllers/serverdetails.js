angular.module("sbAdminApp")
  .controller("ServerDetailsCtrl", ["$scope", "StorageService",
  function($scope, StorageService) 
    StorageService.sendBlockDeviceListRequest().then(function(msg) {
      $scope.blockDevices = msg.payload.blockDevices;
    }).catch(function(err) {
      $scope.errorMsg = err;
    });
  }])
