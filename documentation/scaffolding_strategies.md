---
Title: Scaffolding Strategies
---

Scaffolding Strategies
======================

**Dockership** suggests self-contained repositories, this means that every repository to be deploy should contain his Dockerfile, another techniques could be used as we explain in this document.

## Self-contained Repositories

This is the simplest way to work with Dockership: Just place a `Dockerfile` at the root of you repository. (This is the default location, you can configure your own). 

In this example we deploy an AngularJS tool using [dockerfile/nginx](https://github.com/dockerfile/nginx), the container should retrieve the code through a git clone and execute all the required command to work properly.

```Dockerfile
FROM dockerfile/nginx

# pre requisites
RUN apt-get install -y npm nodejs
RUN ln -s /usr/bin/nodejs /usr/local/bin/node
RUN npm install -g bower grunt karma

# source code
RUN rm -rf /var/www
RUN git clone git@github.com:example/corporate-site.git /var/www

# post code commands
WORKDIR /var/www
RUN npm install
RUN bower install --allow-root

EXPOSE 80

# boot
CMD ["/usr/sbin/nginx"]
```

This project at the `dockership.ini` can be configured as follows:

```ini
[Project "corporate-site"]
Repository = git@github.com:example/corporate-site.git
Port = 0.0.0.0:80:80/tcp
Enviroment = live
```

## Centralised dockerfiles repository

Maybe you prefer keep all your company's `dockerfiles` at the same repository, keeping away from the development team those files.

Based on this hypothetic scaffolding: 
```
devops
 |_ dokerfiles
     |_ CorporateSiteDokerFile
     |_ InternalSiteDockerfile
```

Our `dockership.ini` will looks like:

```ini
[Project "corporate-site"]
Repository =  git@github.com:example/devops.git
RelatedRepository = git@github.com:example/corporate-site.git
Dockerfile = dockerfiles/CorporateSiteDokerFile
Port = 0.0.0.0:80:80/tcp
Enviroment = live

[Project "internal-site"]
Repository =  git@github.com:example/devops.git
RelatedRepository = git@github.com:example/internal-site.git
Dockerfile = dockerfiles/InternalSiteDockerfile
Port = 0.0.0.0:80:80/tcp
Enviroment = live
```

Using the `RelatedRepository` we can track any change at the deployed project, having the dockerfile in other repository. But we have a caveat, every time a file is changed at the `devops` repository even unrelated to the project, this project will look outdated.
