angular.module('dockership', ['ui.bootstrap']);
angular.module('dockership').controller(
    'MainCtrl',
    function ($scope, $http, $modal, $log) {
        'use strict';

        $scope.open = function (project) {
            $http.get('/containers/' + project.Name).then(function(res) {
                var modalInstance = $modal.open({
                  templateUrl: 'ContainersContent.html',
                  controller: 'ContainersCtrl',
                  size: 'lg',
                  resolve: {
                    project: function () {
                        return project;
                    },
                    containers: function () {
                        return res.data;
                    }
                  }
                });
            }, function(msg) {
                $scope.log(msg.data);
            });
        };

        $scope.isDeployable = function(status) {
            var running = status.RunningContainers;
            var commit = status.LastCommit;
            for (var i = running.length - 1; i >= 0; i--) {
                var tmp = running[i].Image.split(':');
                if (commit.slice(0, tmp[0].length+1) == tmp[1]) {
                    return false;
                }
            };

            return true;
        };

        $http.get('/status/').then(function(res) {
            $scope.groups = res.data;
            for (var i in res.data.Errors) {
                $scope.log(res.data.Errors[i]);
            }
        }, function(msg) {
            $scope.log(msg.data);
        });

        $scope.log = function(msg) {
            $scope.errors.push(msg);
        };
    }
);

angular.module('dockership').controller(
    'ContainersCtrl',
    function ($scope, $modalInstance, $http, project, containers) {
        $scope.project = project;
        $scope.containers = containers;

        console.log(containers)
        $scope.ok = function () {
            $modalInstance.close($scope.selected.item);
        };

        $scope.cancel = function () {
            $modalInstance.dismiss('cancel');
        };
    }
);
