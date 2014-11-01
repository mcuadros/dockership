angular.module('dockership', ['ui.bootstrap', 'angular-loading-bar', 'ngAnimate', 'headroom']);
angular.module('dockership').controller(
    'MainCtrl',
    function ($scope, $http, $modal, $log) {
        'use strict';
        $scope.processing = false;
        $scope.openContainers = function (project) {
            $scope.processing = true;
            $http.get('/rest/containers/' + project.Name).then(function(res) {
                $scope.processing = false;
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
                    },
                    loadStatus: function () {
                        return $scope.loadStatus;
                    }
                }
            });
        };

        $scope.isDeployable = function(status) {
            var running = status.RunningContainers;
            var revision = status.LastRevisionLabel;
            for (var i = running.length - 1; i >= 0; i--) {
                var tmp = running[i].Image.split(':');
                if (revision.slice(0, tmp[1].length) == tmp[1]) {
                    return false;
                }
            };

            return true;
        };

        $scope.loaded = false;
        $scope.loadStatus = function() {
            $http.get('/rest/status/').then(function(res) {
                $scope.groups = res.data;
                for (var i in res.data.Errors) {
                    $scope.log(res.data.Errors[i]);
                }

                $scope.loaded = true;
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
    'HeaderCtrl',
    function ($scope, $http, $log) {
        $scope.loadUser = function() {
            $http.get('/rest/user').then(function(res) {
                $scope.user = res.data;
            }, function(msg) {
                $scope.log(msg.data);
            });
        }

        $scope.loadUser()
    }
);

angular.module('dockership').controller(
    'ContainersCtrl',
    function ($scope, $modalInstance, $http, project, containers) {
        $scope.project = project;
        $scope.containers = containers;

        $scope.cancel = function () {
            $modalInstance.dismiss('cancel');
        };
    }
);


angular.module('dockership').controller(
    'DeployCtrl',
    function ($scope, $modalInstance, $http, oboe, project, enviroment, loadStatus) {
        $scope.project = project;
        $scope.enviroment = enviroment;

        $scope.data = [];
        $scope.data = oboe({
            url: '/rest/deploy/' + project.Project.Name + '/' + enviroment.Name,
            pattern: '{msg}',
        });

        $scope.params = function (params, first) {
            var strings = [];
            angular.forEach(params, function(value, key) {
                if (key != "t" && key != "msg" && key != "lvl" && key != "revision") {
                    this.push('<b>' + key + '</b>: ' + value);
                }

                if (key == "revision") {
                    this.push('<b>' + key + '</b>: ' + value.slice(0,12));
                }
            }, strings);

            if (first) {
                return strings[0].replace(/<[^>]+>/gm, '');
            }

            return strings.join("<br /> ");
        };

        $scope.cancel = function () {
            loadStatus()
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

angular.module('dockership').filter('unsafe', ['$sce', function ($sce) {
    return function (val) {
        return $sce.trustAsHtml(val);
    };
}]);


// update popover template for binding unsafe html
angular.module("template/popover/popover.html", []).run(["$templateCache", function ($templateCache) {
    $templateCache.put("template/popover/popover.html",
      "<div class=\"popover {{placement}}\" ng-class=\"{ in: isOpen(), fade: animation() }\">\n" +
      "  <div class=\"arrow\"></div>\n" +
      "\n" +
      "  <div class=\"popover-inner\">\n" +
      "      <h3 class=\"popover-title\" ng-bind-html=\"title | unsafe\" ng-show=\"title\"></h3>\n" +
      "      <div class=\"popover-content\"ng-bind-html=\"content | unsafe\"></div>\n" +
      "  </div>\n" +
      "</div>\n" +
      "");
}]);
