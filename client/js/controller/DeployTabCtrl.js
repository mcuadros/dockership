module.exports = [
  '$scope', '$http', '$rootScope', 'socket', 'ansi2html',
  function ($scope, $http, $rootScope, socket, ansi2html) {
    $scope.log = {};
    $scope.current = 'latest';
    $rootScope.pendingDeployments = 0;

    socket.addHandler('deploy', function (result) {
      var key = result.project + ' ' + result.environment + ' '
        + result.date.slice(0, 16);
      $scope.current = key;

      if ($scope.log[key] == undefined) {
        $rootScope.pendingDeployments++;
        $scope.log[key] = ''
      }
      $scope.log[key] += $scope.ansii ? ansi2html.toHtml(result.log) : result.log;
    });
  }
];