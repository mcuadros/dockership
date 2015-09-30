module.exports = [
  '$scope', '$http', 'socket', 'dialogs', '$modal',
  function ($scope, $http, socket, dialogs, $modal) {
    'use strict';
    $scope.processing = false;

    socket.addHandler('containers', function (result) {
      $scope.processing = false;
      var modalInstance = $modal.open({
        templateUrl: 'ContainersContent.html',
        controller: 'ContainersCtrl',
        size: 'lg',
        resolve: {
          project: function () {
            return '';
          },
          containers: function () {
            return result;
          }
        }
      });
    });

    var envStatus = function (status) {
      if (status == undefined) {
        return ['loading'];
      }

      var total = status.Environment.DockerEndPoints.length;
      var running = status.RunningContainers;
      var revision = status.LastRevisionLabel;
      var outdated = 0;

      if (running.length == 0) {
        return ['down']
      }

      if (running.length != total) {
        return ['down', 'partial']
      }

      for (var i = running.length - 1; i >= 0; i--) {
        var tmp = running[i].Image.split(':');
        if (tmp.length < 2) {
            continue;
        }

        if (revision.slice(0, tmp[1].length) != tmp[1]) {
          outdated++;
        }
      };

      if (outdated == total) {
        return ['outdated']
      }

      if (outdated != 0) {
        return ['outdated', 'partial']
      }

      return ['ok'];
    };

    socket.addHandler('status', function (result) {
      $scope.processing = false;
      $scope.loaded = true;

      angular.forEach(result, function (project, key) {
        angular.forEach(project.Status, function (status, key) {
          status.Status = envStatus(status);
        });
      });

      $scope.status = result;

    });

    socket.addHandler('projects', function (result) {
      $scope.projects = result;

      angular.forEach(result, function (project, key) {
        $scope.taskStatus[project.Name] = project.TaskStatus;
      });

      $scope.loadStatus();
    });

    $scope.openContainers = function (project) {
      $scope.processing = true;
      socket.getContainers(project);
    };

    $scope.taskStatus = [];
    $scope.openDeploy = function (project, environment) {
      var msg = 'Are you sure want to deploy <b>' + project.Name + '</b> at <b>'
        + environment.Name + '</b>?'    ;
      var dlg = dialogs.confirm('Confirm', msg, {size: 'md'});
      dlg.result.then(function (btn){
        if ($scope.taskStatus[project.Name][environment.Name] == undefined) {
          $scope.taskStatus[project.Name][environment.Name] = {};
        }

        $scope.taskStatus[project.Name][environment.Name]['deploy'] = true;
        socket.doDeploy(project, environment);
      });
    };


    $scope.loaded = false;
    $scope.loadStatus = function () {
      socket.getStatus();
    };
  }
];
