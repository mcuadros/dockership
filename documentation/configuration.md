---
Title: Configuration
---

Configuration
=============

The **dockership** configuration is based on a INI-formatted config file. 

Dockership will look at `/etc/dockership/dockership.conf` for this config file by default. The `-config` flag may be passed to the `dockershipd` or `dockership` binaries to use a custom config file location.

## Syntax
The config file syntax is based on [git-config](http://git-scm.com/docs/git-config#_syntax), with minor changes.  

The file consists of **sections** and **variables**. A section begins with the name of the section in square brackets and continues until the next section begins. Section names are not case sensitive. Only alphanumeric characters, - and . are allowed in section names. Each variable must belong to some section, which means that there must be a section header before the first setting of a variable.

Sections can be further divided into subsections. To begin a subsection put its name in double quotes, separated by space from the section name, in the section header, like in the example below:

```ini
; Comment line
[section "subsection"]
name = value # Another comment
flag # implicit value for bool is true
```


## Sections

### Global

A miscellaneous of configuration variables used across the whole tool.

* `GithubToken` (mandatory): a Github [personal access token](https://github.com/settings/tokens/new) used in every request to the [Github API](https://developer.github.com/). 

* `UseShortRevisions` (default: true): if is false all the images and containers will be defined using full length revision names, instead the short one.

### HTTP

Configuration of the web server from `dockershipd`. *This section is only required by dockershipd*

Since the authentication is based on a registered [Github Application](https://github.com/settings/applications/new), you should [create it](https://github.com/settings/applications/new) at Github. The *Authorization callback URL* must be filled with `http://<server-addr>/oauth2callback`, this URL should be accessible by anyone.

* `Listen` (default: 0.0.0.0:80): the TCP network address
* `GithubID` (mandatory): the `Client ID` provided by Github
* `GithubSecret` (mandatory): the `Client Secret` provided by Github
* `GithubOrganization` (optional): if given only the members from this Github Organization are allowed to access.
* `GithubRedirectURL` (mandatory): the `Authorization callback URL` configured in Github

### Environment

A environment is a logical group of any number of Docker servers. Dockership supports multiple environments. Each `Environment` is defined as a section with subsection: `[Enviroment "production"]`

* `DockerEndPoint` (mandatory, multiple): [Docker Remote API](https://docs.docker.com/reference/api/docker_remote_api/) address, if dockership and docker are running in the same host you can use `unix:///var/run/docker.sock` if not you should enable remote access at the docker daemon (with -H parameter) and use a TCP endpoint. (eg.: `http://172.17.42.1:4243`)

### Project

`Project` section defines the configuration for every project to be deployed in the environments. The relation between repositories is one-to-one, so the repository should contain the `Dockerfile` and all the files needed to build the Docker image. The Project as Environment is defined as a section with subsection: `[Project "disruptive-app"]`

* `Repository` (mandatory): Github repository SSH clone URL (eg.: `git@github.com:mcuadros/dockership.git`)
* `Branch` (optional): branch to be deployed
* `Dockerfile` (default: Dockerfile): the path to the Dockerfile at the repository.
* `RelatedRepositories` (optional, multiple): SSH clone URL to dependant repositories. (Link to more explanatory document)
* `History` (default: 3): Number to old images you want to keep in each Docker server. 
* `NoCache` (optional): Avoid to use the Docker cache (like --no-cache at `docker build`)
* `Port` (multiple, optional): container port to expose, format: `<host-addr>:<host-port>:<container-port>/<proto>` (like -p at `docker run`)
* `Link` (multiple, optional): creates a Link to other project, when this project is deployed the linked projects are restarted (like -P at `docker run`)
* `GithubToken` (default: Global.GithubToken): the token needed to access this repository, if is different from the global one.
* `Enviroment` (multiple, mandatory): Environment name where this project could be deployed

## Example

### Scenario
#### rest-service project
REST webservice in Python running under a uwsgi+nginx on port 8080

This repository requires the python package `domain`, so we want detect if the rest-service has pending changes to be deployed when the domain has new commits, even when the `rest-service` repository do not have new commits.

#### frontend project
An AngularJS frontend running on a nginx server, with a `reverse_proxy` pointing to the port 8080 at rest-service container, in the path `/rest`.

We want to expose the port 80 at the host server.

### Config file
```ini
[Global]
GithubToken = example-token

[Project "rest-service"]
Repository = git@github.com:company/rest-service.git
Enviroment = live
Enviroment = dev
File = /tmp/container.py
RelatedRepository = git@github.com:company/domain.git

[Project "frontend"]
Repository = git@github.com:company/angular-client.git
Enviroment = live
Enviroment = dev
Port = 0.0.0.0:80:80/tcp
Link = rest-service:backend

[Enviroment "live"]
DockerEndPoint = http://live-1.example.com
DockerEndPoint = http://live-2.example.com

[Enviroment "dev"]
DockerEndPoint = http://dev.example.com
```
