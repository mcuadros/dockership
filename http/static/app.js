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
        $scope.level = 4
        $scope.chageLevel = function(level) {
            $scope.level = level;
        };
        $scope.filter = function(line) {
            console.log($scope.level);
            if (line.lvl <= $scope.level) {
                return true;
            }


            return false;
        };

        $scope.log = [];
        socket.addHandler('log', function (result) {
            $scope.log.unshift(result);
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
    'DeployTabCtrl',
    function ($scope, $http, socket, ansi2html) {
        $scope.log = ""
        socket.addHandler('deploy', function (result) {
            console.log(result);
            $scope.log += ansi2html.toHtml(result.log)
        });
    }
);

angular.module('dockership').controller(
    'MainCtrl',
    function ($scope, $http, socket, $modal, $log) {
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
                        return "";
                    },
                    containers: function () {
                        return result;
                    }
                }
            });
        });

        socket.addHandler('status', function (result) {
            $scope.processing = false;
            $scope.loaded = true;
            $scope.groups = result;
        });

        $scope.openContainers = function (project) {
            $scope.processing = true;
            socket.getContainers(project);
        };

        $scope.openDeploy = function (project, environment) {
            console.log(project, environment);
            socket.doDeploy(project, environment);
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
            socket.getStatus();
        };


        $scope.log = function(msg) {
            $scope.errors.push(msg);
        };

        socket.setHandler('open', function() {
            $scope.loadStatus();
        })
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

angular.module('dockership').factory('socket', function (socketFactory) {
    var socket = socketFactory({
        url: '/socket'
    });

    socket._handlers = {};
    socket.addHandler = function(name, handler) {
        socket._handlers[name] = handler;
    };

    socket.doDeploy = function(project, environment) {
        socket.send(angular.toJson({
            event: 'deploy',
            request: {
                project: project.Name,
                environment: environment.Name
            }
        }))
    };

    socket.getContainers = function(project) {
        socket.send(angular.toJson({
            event: 'containers',
            request: {project: project.Name}
        }))
    };

    socket.getStatus = function(project) {
        socket.send(angular.toJson({
            event: 'status',
            request: {}
        }))
    };

    socket.setHandler('message', function (e) {
        data = angular.fromJson(e.data);
        socket._handlers[data.event](data.result);
    });


    return socket;
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
