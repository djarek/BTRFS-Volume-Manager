angular.module("sbAdminApp")
  .controller("ServerDetailsCtrl", ["$scope", "$stateParams", "StorageService",
  function($scope, $stateParams, StorageService) {
    $scope.serverid = $stateParams.id;
    StorageService.sendBlockDeviceListRequest($stateParams.id).then(function(msg) {
      $scope.blockDevices = msg.payload.blockDevices;
    }).catch(function(err) {
      $scope.errorMsg = err;
    });
    $scope.rescanBtrfsVolumes = function() {
      StorageService.sendBtrfsVolumeListRequest($stateParams.id).then(function(msg) {
        $scope.btrfsVolumes = msg.payload.btrfsVolumes;
      }).catch(function(err) {
        $scope.errorMsg = err;
      });
    };
    $scope.rescanBlockDevices = function() {
      StorageService.sendBlockDeviceRescanRequest($stateParams.id).then(function(msg) {
        $scope.blockDevices = msg.payload.blockDevices;
      }).catch(function(err) {
        $scope.errorMsg = err;
      })
    };

    $scope.rescanBtrfsVolumes();
  }])
