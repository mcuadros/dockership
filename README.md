![Dockership](https://cdn.rawgit.com/mcuadros/dockership-site/master/static/images/dockership.png)
### [![Build Status](https://travis-ci.org/mcuadros/dockership.svg?branch=master)](https://travis-ci.org/mcuadros/dockership)

The deploy is based on git repositories (currently only supports Github ones) containing the Dockerfile for each project. Dockership handles the building and running process at one or multiple docker servers, the  version control is made through the git commits, being extremly easy.

> WARNING: Currently the project is under heavy development and undocumented.

Example
--------

The configuration is based on a INI-style config file like this
```ini
[Project "corporate-site"]
Repository = git@github.com:example/corporate-site.git
Enviroment = live
Enviroment = dev
Port = 0.0.0.0:80:80/tcp


[Enviroment "live"]
DockerEndPoint = http://live-1.example.com
DockerEndPoint = http://live-2.example.com


[Enviroment "dev"]
DockerEndPoint = http://dev.example.com
```


License
-------

MIT, see [LICENSE](LICENSE)

