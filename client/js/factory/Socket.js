module.exports = [
  'socketFactory',
  function (socketFactory) {
    var socket = socketFactory({
      url: '/socket'
    });

    socket._handlers = {};
    socket.addHandler = function (name, handler) {
      socket._handlers[name] = handler;
    };

    socket.doDeploy = function (project, environment) {
      socket.send(angular.toJson({
        event: 'deploy',
        request: {
          project: project.Name,
          environment: environment.Name
        }
      }))
    };

    socket.getContainers = function (project) {
      socket.send(angular.toJson({
        event: 'containers',
        request: {project: project.Name}
      }))
    };

    socket.getStatus = function (project) {
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
  }
];