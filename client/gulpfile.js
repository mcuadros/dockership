var gulp = require('gulp');
var gutil = require('gulp-util');
var browserify = require('browserify');
var watchify = require('watchify');
var source = require('vinyl-source-stream');
var buffer = require('vinyl-buffer');
var uglify = require('gulp-uglify');
var sourcemaps = require('gulp-sourcemaps');
var clean = require('gulp-clean');
var debug = require('gulp-debug');
var compass = require('gulp-compass');
var minifyCSS = require('gulp-minify-css');
var plumber = require('gulp-plumber');
var filter = require('gulp-filter');

var SASS_CACHE = './.sass-cache';
var SCSS_ENTRY = './src/assets/scss/style.scss';
var BUILD_DIR = './../http/static';
var SRC_DIR = './js';
var JS_APP_FILE = 'app.js';

var executeJsPipeline = function (stream) {
  return stream
    .bundle()
    .pipe(source(JS_APP_FILE))
    .pipe(buffer())
    .pipe(sourcemaps.init({loadMaps: true}))
    .pipe(uglify())
    .on('error', gutil.log)
    .pipe(sourcemaps.write('./', {
      sourceRoot: '../src/js'
    }))
    .pipe(gulp.dest(BUILD_DIR));
};

gulp.task('javascript', function () {
  var bundler = browserify({
    entries: SRC_DIR + '/' + JS_APP_FILE,
    debug: true
  });

  return executeJsPipeline(bundler);
});

gulp.task('watch-js', function() {
  var bundler = browserify({
    entries: SRC_DIR + '/' + JS_APP_FILE,
    debug: true
  });
  var watcher  = watchify(bundler);

  watcher
    .on('update', function () {
      var updateStart = Date.now();
      console.log('Updating...');
      executeJsPipeline(watcher).on('end', function () {
        console.log((new Date()).toTimeString().split(' ')[0] + ' Updated!', (Date.now() - updateStart) + 'ms');
      });
    });

  console.log('Building and generating cache for watcher!');
  executeJsPipeline(watcher).on('end', function () {
    console.log((new Date()).toTimeString().split(' ')[0] + ' Finished cache generation!');
  });
  console.log('Watcher running...');

  return watcher;
});

// Compile .scss files to plain .js in tmp/.
gulp.task('css', function () {
  gutil.log(gutil.colors.yellow('Building CSS from SCSS'));
  return gulp.src(SCSS_ENTRY)
    .pipe(sourcemaps.init())
    .pipe(plumber({
      errorHandler: function (error) {
        console.log(error.message);
        this.emit('end');
      }}))
    .pipe(compass({
      config_file: './config.rb',
      css: BUILD_DIR,
      sass: 'src/assets/scss'
    }))
    .on('error', function(err) {
      console.log(err);
    })
    .pipe(minifyCSS())
    .pipe(sourcemaps.write('.'))
    .pipe(gulp.dest(BUILD_DIR));
});

gulp.task('build', ['javascript', 'css'], function () {
  gutil.log(gutil.colors.yellow('Removing temporary files & folders.'));
  return gulp.src(SASS_CACHE)
    .pipe(clean());
});

gulp.task('default', ['build']);