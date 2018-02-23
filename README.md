![Dockership](https://cdn.rawgit.com/mcuadros/dockership-site/master/static/images/dockership.png)
### [![Latest Stable Version](http://img.shields.io/github/release/mcuadros/dockership.svg?style=flat)](https://github.com/mcuadros/dockership/releases) [![Build Status](http://img.shields.io/travis/mcuadros/dockership.svg?style=flat)](https://travis-ci.org/mcuadros/dockership)

**Dockership** is a tool for easily deploying Docker containers to one or multiple Docker servers.

Why?
----
Nowadays we have many powerful tools for *configuration management* such as [Puppet](http://puppetlabs.com/puppet/what-is-puppet), [Chef](http://www.getchef.com/chef/) and  [Ansible](http://www.ansibleworks.com/) even docker-based deployments tools like [Deis](http://deis.io). These tools are great for medium/big projects, but not optimal for small startups without a DevOps expert, and personal side projects.

With Dockership you can deploy your applications, based on a Docker container, to several servers without learning complex DSLs or hundreds of new terms. Learning new things is great, but when you deploy from time to time, remember how to do it becomes hard to remember.


Overview
--------

The deploy is based on git repositories (currently only supports Github ones) containing the Dockerfile for each project. Dockership handles the building and running process at one or multiple Docker servers, the version control is made through the git commits, being extremely easy.

Dockership comes in two flavours CLI and HTTP, here you can see a screenshot from the HTTP view.

![Projects View](https://raw.githubusercontent.com/mcuadros/dockership-site/master/static/images/screenshots/http-project-view-min.png)

Documentation
-------------

* [Configuration](http://docker.sh/ip/documentation/configuration/)
* [Creating Your Dockerfiles](http://docker.sh/ip/documentation/creating_your_dockerfiles/)
* [Scaffolding Strategies](http://docker.sh/ip/documentation/scaffolding_strategies/)


Example
-------

The configuration is based on an INI-style config file like this
```ini
[Project "corporate-site"]
Repository = git@github.com:example/corporate-site.git
Environment = live
Environment = dev
Port = 0.0.0.0:80:80/tcp


[Environment "live"]
DockerEndPoint = http://live-1.example.com
DockerEndPoint = http://live-2.example.com


[Environment "dev"]
DockerEndPoint = http://dev.example.com
```

License
-------

MIT, see [LICENSE](LICENSE)
