angular.module("sbAdminApp")
  .controller("VolumeDetailsCtrl", ["$scope", "$stateParams", "StorageService",
  function($scope, $stateParams, StorageService) {
    $scope.volid = $stateParams.volid;
    $scope.rescanBtrfsSubvolumes = function() {
      StorageService.sendBtrfsSubvolumeListRequest($stateParams.id, $stateParams.volid).then(function(msg) {
        $scope.subvolumes = msg.payload.subvolumes;
      }).catch(function(err) {
        $scope.errorMsg = err;
      })
    };

    $scope.rescanBtrfsSubvolumes();
  }])
