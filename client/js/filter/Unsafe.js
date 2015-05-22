module.exports = [
  '$sce',
  function ($sce) {
    return function (val) {
      return $sce.trustAsHtml(val);
    };
  }
];