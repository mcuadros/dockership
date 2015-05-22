module.exports = [
  '$scope', '$http',
  function ($scope, $http) {
    $scope.loadUser = function () {
      $http.get('/rest/user').then(function (res) {
        $scope.user = res.data;
      }, function (msg) {
        $scope.log(msg.data);
      });
    };

    $scope.loadUser()
  }
];