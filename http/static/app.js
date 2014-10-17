angular.module('dockership', ['ui.bootstrap', 'angular-loading-bar', 'ngAnimate']);
angular.module('dockership').controller(
    'MainCtrl',
    function ($scope, $http, $modal, $log) {
        'use strict';

        $scope.openContainers = function (project) {
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

        $scope.openDeploy = function (project, enviroment) {
             var modalInstance = $modal.open({
                templateUrl: 'DeployContent.html',
                controller: 'DeployCtrl',
                size: 'lg',
                resolve: {
                    project: function () {
                        return project;
                    },
                    enviroment: function () {
                        return enviroment;
                    }
                }
            });
        };

        $scope.isDeployable = function(status) {
            var running = status.RunningContainers;
            var commit = status.LastCommit;
            for (var i = running.length - 1; i >= 0; i--) {
                var tmp = running[i].Image.split(':');
                if (commit.slice(0, tmp[1].length) == tmp[1]) {
                    return false;
                }
            };

            return true;
        };

        $scope.loadStatus = function() {
            $http.get('/status/').then(function(res) {
                $scope.groups = res.data;
                for (var i in res.data.Errors) {
                    $scope.log(res.data.Errors[i]);
                }
            }, function(msg) {
                $scope.log(msg.data);
            });
        }

        $scope.log = function(msg) {
            $scope.errors.push(msg);
        };

        $scope.loadStatus()
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


angular.module('dockership').controller(
    'DeployCtrl',
    function ($scope, $modalInstance, $http, oboe, project, enviroment) {
        $scope.project = project;
        $scope.enviroment = enviroment;

        $scope.data = [];
        $scope.data = oboe({
            url: '/deploy/' + project.Project.Name + '/' + enviroment.Name,
            pattern: '{msg}',
            pagesize: 1
        });
        $scope.ok = function () {
            $modalInstance.close($scope.selected.item);
        };

        $scope.cancel = function () {
            $modalInstance.dismiss('cancel');
        };
    }
);

angular.module('dockership').service('oboe', [
    'OboeStream',
    function (OboeStream) {
        return function (params) {
            var data = [];
            OboeStream.stream(params, function (node) {
                data.unshift(node);
            });

            return data;
        };
    }
]).factory('OboeStream', [
    '$q',
    function ($q) {
        return {
            stream: function (params, callback) {
                var defer = $q.defer();
                var promise = defer.promise;

                oboe(params).start(function () {
                    defer.resolve();
                }).node(params.pattern || '.', function (node) {
                    promise.then(callback(node));
                    return oboe.drop;
                });
                return promise;
            }
        };
    }
]);
