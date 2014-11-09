angular.module('dockership', [
    'ui.bootstrap', 'angular-loading-bar', 'ansiToHtml', 'ngAnimate',
    'bd.sockjs', 'headroom'
]);

angular.module('dockership').controller(
    'LogTabCtrl',
    function ($scope, socket, ansi2html) {
        $scope.level = 4
        $scope.chageLevel = function(level) {
            $scope.level = level;
        };
        $scope.filter = function(line) {
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
        $scope.log = {}
        $scope.current = "latest"
        socket.addHandler('deploy', function (result) {
            var key = result.project + " " + result.environment + " " + result.date.slice(0, 16)
            $scope.current = key;

            if ($scope.log[key] == undefined) {
                $scope.log[key] = ""
            }

            $scope.log[key] += ansi2html.toHtml(result.log)
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

        var envStatus = function(status) {
            if (status == undefined) {
                return ["loading"];
            }

            var total = status.Environment.DockerEndPoints.length;
            var running = status.RunningContainers;
            var revision = status.LastRevisionLabel;
            var outdated = 0;

            if (running.length == 0) {
                return ["down"]
            }

            if (running.length != total) {
                return ["down", "partial"]
            }

            for (var i = running.length - 1; i >= 0; i--) {
                var tmp = running[i].Image.split(':');
                if (revision.slice(0, tmp[1].length) != tmp[1]) {
                    outdated++;
                }
            };

            if (outdated == total) {
                return ["outdated", "partial"]
            }

            if (outdated != 0) {
                return ["outdated"]
            }

            return ["ok"];
        };

        var isDeployable = function(status) {
            if (status == undefined) {
                return null;
            }

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

        socket.addHandler('status', function (result) {
            $scope.processing = false;
            $scope.loaded = true;
            console.log(result);

            angular.forEach(result, function(project, key) {
                angular.forEach(project.Status, function(status, key) {
                    status.IsDeployable = isDeployable(status);
                    status.Status = envStatus(status);
                });
            });

            $scope.status = result;

        });

        socket.addHandler('projects', function (result) {
            $scope.projects = result;
            $scope.loadStatus();
        });

        $scope.openContainers = function (project) {
            $scope.processing = true;
            socket.getContainers(project);
        };

        $scope.openDeploy = function (project, environment) {
            socket.doDeploy(project, environment);
        };

        
        $scope.loaded = false;
        $scope.loadStatus = function() {
            socket.getStatus();
        };
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

angular.module('dockership').directive('myStatus', function () {
    return {
        scope: false,  // this is the default, so you could remove this line
        template: "<div><div another-directive></div>{{button}}</div>",
        replace: true,
        link: function (scope, element, attrs) {
           scope.button = "foo";
           console.log(scope);
        }
    };
});


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
