angular.module('dockership', [
  'ui.bootstrap',
  'angular-loading-bar',
  'ansiToHtml',
  'ngAnimate',
  'bd.sockjs',
  'headroom',
  'dialogs.main',
  'notification'
]).run(['$templateCache', function ($templateCache) {
  $templateCache.put('template/popover/popover.html',
    '<div class="popover {{placement}}" ng-class="{ in: isOpen(), fade: animation() }">' +
    '  <div class="arrow"></div>' +
    '  <div class="popover-inner">' +
    '      <h3 class="popover-title" ng-bind-html="title | unsafe" ng-show="title"></h3>' +
    '      <div class="popover-content" ng-bind-html="content | unsafe"></div>' +
    '  </div>' +
    '</div>'
  );
}]);

require('./config');
