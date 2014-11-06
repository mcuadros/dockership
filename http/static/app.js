angular.module('dockership', [
    'ui.bootstrap', 'angular-loading-bar', 'ansiToHtml', 'ngAnimate',
    'bd.sockjs', 'headroom'
]);

angular.module('dockership').controller(
    'TabsParentController', function ($scope, $window) {
        var setAllInactive = function() {
            angular.forEach($scope.workspaces, function(workspace) {
                workspace.active = false;
            });
        };

        var addNewWorkspace = function() {
            var id = $scope.workspaces.length + 1;
            $scope.workspaces.push({
                id: id,
                name: "Workspace " + id,
                template: 'DeployContent.html',
                ctrl: 'DeployCtrl',
                active: true
            });
        };

        $scope.workspaces = [{
            name: 'List', active:true, ctrl: 'MainCtrl',
            name: 'Log', active:false, ctrl: 'LogTabCtrl'
        }];

        $scope.addWorkspace = function () {
            setAllInactive();
            addNewWorkspace();
        };
    }
);

angular.module('dockership').controller(
    'LogTabCtrl',
    function ($scope, socket, ansi2html) {
        $scope.log = [];
        socket.setHandler('message', function (e) {
            data = angular.fromJson(e.data);
            $scope.log.unshift(data.result);
        });

        socket.setHandler('open', function (data) {
            console.log(data);
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
    }
);


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

        $scope.openDeploy = function (project, environment) {
             var modalInstance = $modal.open({
                templateUrl: 'DeployContent.html',
                controller: 'DeployCtrl',
                size: 'lg',
                resolve: {
                    project: function () {
                        return project;
                    },
                    environment: function () {
                        return environment;
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
    function ($scope, $http, ansi2html, project, environment, loadStatus) {
        $scope.project = project;
        $scope.environment = environment;
        $scope.log = ""

       /* socket.on('deploy', function (msg) {
            $scope.log += ansi2html.toHtml(msg)
        });
*/
        $http.get('/rest/deploy/' + project.Project.Name + '/' + environment.Name).then(function(res) {
            console.log("done");
        }, function(msg) {
            $scope.log(msg.data);
        });

        $scope.cancel = function () {
            loadStatus()
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

angular.module('dockership').factory('socket', function (socketFactory) {
    return socketFactory({
    url: '/socket'
  });
});

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
