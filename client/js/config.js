var components = {
  controller: {
    ContainersCtrl: require('./controller/ContainersCtrl'),
    DeployTabCtrl: require('./controller/DeployTabCtrl'),
    HeaderCtrl: require('./controller/HeaderCtrl'),
    LogTabCtrl: require('./controller/LogTabCtrl'),
    ProjectsCtrl: require('./controller/ProjectsCtrl')
  },
  filter: {
    unsafe: require('./filter/Unsafe')
  },
  factory: {
    socket: require('./factory/Socket')
  }
};

// bootstrap components
var componentType;
var component;
for (componentType in components) {
  for (component in components[componentType]) {
    angular.module('dockership')[componentType](
      component,
      components[componentType][component]
    );
  }
}