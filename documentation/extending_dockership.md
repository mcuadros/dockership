---
Title: Extending Dockership
---

Extending Dockership
====================

Dockership doesn't provide a plugin system. However, it exposes some machine-friendly interfaces you can write programs against to extend its basic functionality.

Web hooks
---------

You can get notifications pushed to your server when some events happen in Dockership. To do this, **add the `WebHook` property to a project** in [your configuration file](https://github.com/mcuadros/dockership/blob/master/documentation/configuration.md#project).

All notifications will be sent as POST requests with a JSON value in the body, thus with an `application/json` content type. will be sent when a deploy succeeds for this project. The request will have a JSON object body with keys `previous_revision`, `current_revision`, `project`, `environment` and `errors`.

Currently, the only notification happens **after a deploy**. The sent JSON value will be an object with:

* `previous_revision`, the identifier (typically, the hash of a commit) of the running revision of the code _before_ the deployment happened.
* `current_revision`, the identifier of the currently running revision, after the deployment.
* `project` name, as defined in [Configuration](https://github.com/mcuadros/dockership/blob/master/documentation/configuration.md#project).
* `environment` name, as defined in [Configuration](https://github.com/mcuadros/dockership/blob/master/documentation/configuration.md#environment).
* `errors`, an array of strings describing the errors that prevented the deployment, if any. If the array is not empty, the deployment failed, and `current_revision` will be the same as `previous_revision`. 

HTTP endpoints
--------------

For pulling information out of Dockership, you can access some resources exposed as JSON values in HTTP endpoints.

* `/rest/projects` is an object containing the projects defined in the configuration indexed by project name. Each entry in the object is the JSON serialization of a [`Project`](http://godoc.org/github.com/mcuadros/dockership/core#Project) value.
* `/rest/status` is an object containing the status of each project indexed by project name. Each entry in the object is the JSON serialization of a [`StatusResult`](http://godoc.org/github.com/mcuadros/dockership/http#StatusResult) value.
* `/rest/status/:project`, `:project` being a placeholder for a project name, is the entry for the desired project in the object given at `/rest/status`.
