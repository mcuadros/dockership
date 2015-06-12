module.exports = [
  '$scope', '$rootScope', 'socket', 'Notification',
  function ($scope, $rootScope, socket, Notification) {
    $scope.level = 3;
    $rootScope.pendingLogs = 0;

    $scope.chageLevel = function (level) {
      $scope.level = level;
    };

    $scope.filter = function (line) {
      return line.lvl <= $scope.level;
    };

    $scope.log = [];
    socket.addHandler('log', function (log) {
      $scope.log.unshift(log);
      if (log.lvl < 4) {
        $rootScope.pendingLogs++;
      }

      if (log.lvl < 1) {
        var notification = new Notification('[Critical Error]', {
          body: log.msg,
          icon: '/logo.png',
          delay: 4000
        });
      }
    });

    $scope.params = function (params, first) {
      var strings = [];
      angular.forEach(params, function (value, key) {
        if (key != 't' && key != 'msg' && key != 'lvl' && key != 'revision') {
          this.push('<b>' + key + '</b>: ' + value);
        }

        if (key == 'revision') {
          this.push('<b>' + key + '</b>: ' + value.slice(0,12));
        }
      }, strings);

      if (first) {
        return strings[0].replace(/<[^>]+>/gm, '');
      }

      return strings.join('<br /> ');
    };
  }
];