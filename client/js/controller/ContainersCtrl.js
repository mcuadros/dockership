module.exports = [
  '$scope', '$modalInstance', '$http', 'project', 'containers',
  function ($scope, $modalInstance, $http, project, containers) {
    $scope.project = project;
    $scope.containers = containers;

    $scope.cancel = function () {
      $modalInstance.dismiss('cancel');
    };
  }
];