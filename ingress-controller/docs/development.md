# Getting Started

This document explains how to get started with developing for NGINX Ingress controller.
It includes how to build, test, and release ingress controllers.

## Dependencies

The build uses dependencies in the `vendor` directory, which
must be installed before building a binary/image. Occasionally, you
might need to update the dependencies.

This guide requires you to install the [dep](https://github.com/golang/dep) dependency tool.

Check the version of `dep` you are using and make sure it is up to date.

```console
$ dep version
dep:
 version     : devel
 build date  : 
 git hash    : 
 go version  : go1.9
 go compiler : gc
 platform    : linux/amd64
```

If you have an older version of `dep`, you can update it as follows:

```console
$ go get -u github.com/golang/dep
```

This will automatically save the dependencies to the `vendor/` directory.

```console
$ cd $GOPATH/src/github.com/kong/ingress-controller
$ dep ensure
$ dep ensure -update
$ dep prune
```

## Building

All ingress controllers are built through a Makefile. Depending on your
requirements you can build a raw server binary, a local container image,
or push an image to a remote repository.

In order to use your local Docker, you may need to set the following environment variables:

```console
# "gcloud docker" (default) or "docker"
$ export DOCKER=<docker>

# "quay.io/kubernetes-ingress-controller" (default), "index.docker.io", or your own registry
$ export REGISTRY=<your-docker-registry>
```

To find the registry simply run: `docker system info | grep Registry`

### Nginx Controller

Build a raw server binary
```console
$ make build
```

[TODO](https://github.com/kubernetes/ingress-nginx/issues/387): add more specific instructions needed for raw server binary.

Build a local container image

```console
$ TAG=<tag> REGISTRY=$USER/ingress-controller make docker-build
```

Push the container image to a remote repository

```console
$ TAG=<tag> REGISTRY=$USER/ingress-controller make docker-push
```

## Deploying

There are several ways to deploy the ingress controller onto a cluster.
Please check the [deployment guide](../deploy/README.md)

## Testing

To run unit-tests, just run

```console
$ cd $GOPATH/src/github.com/kong/ingress-controller
$ make test
```

If you have access to a Kubernetes cluster, you can also run e2e tests using ginkgo.

```console
$ cd $GOPATH/src/github.com/kong/ingress-controller
$ make e2e-test
```

## Releasing

All Makefiles will produce a release binary, as shown above. To publish this
to a wider Kubernetes user base, push the image to a container registry, like
[gcr.io](https://cloud.google.com/container-registry/). All release images are hosted under `gcr.io/google_containers` and
tagged according to a [semver](http://semver.org/) scheme.

An example release might look like:
```
$ make release
```

Please follow these guidelines to cut a release:

* Update the [release](https://help.github.com/articles/creating-releases/)
page with a short description of the major changes that correspond to a given
image tag.
* Cut a release branch, if appropriate. Release branches follow the format of
`controller-release-version`. Typically, pre-releases are cut from HEAD.
All major feature work is done in HEAD. Specific bug fixes are
cherry-picked into a release branch.
* If you're not confident about the stability of the code,
[tag](https://help.github.com/articles/working-with-tags/) it as alpha or beta.
Typically, a release branch should have stable code.
