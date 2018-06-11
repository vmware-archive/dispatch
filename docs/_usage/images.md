---
title: Images and Base Images
---

# Images and Base Images

Before a function can be created, it needs a runtime.  Images and Base Images provide the OS, language runtime and
dependencies required for functions.

## Base Images

Base Images are intended to provide the core language functionaity.  This generally consists of the OS (Photon) and
the language runtime.  Additionally, base images contain code to interact with the FaaS (a small web server).

The Base Images supplied by Dispatch can be found on DockerHub under the
[dispatchframework](https://hub.docker.com/u/dispatchframework/) organization.  These images contain the bare minimum
required to run a function for a given language.  If you have global system or language specific dependencies, you can
simply extend the base images by creating a new Docker image.

```
$ mkdir my-python-base
$ cd my-python-base
```
```
$ cat << EOF > Dockerfile
FROM dispatchframework/python3-base:0.0.7

RUN pip3 install requests
EOF
```
```
$ docker build -t berndtj/my-python-base:0.0.1 .
$ docker push berndtj/my-python-base:0.0.1
```

### Adding a Base Image

Base Images can be added through the `create base-image` command:

```
$ dispatch create base-image my-python-base berndtj/my-python-base:0.0.1 --language python3
```

Or via generic `create` (multiple resources may be created at once):

```
$ cat << EOF > base-images.yaml
kind: BaseImage
name: my-python-base
dockerUrl: berndtj/my-python-base:0.0.1
language: python3
tags:
  - key: role
    value: docs
EOF
```

```
$ dispatch create -f base-images.yaml
```

### Checking Status

Before a Base Image is `READY`, the server will ensure that it can pull the image.  Verify the image is `READY`:

```
$ dispatch get base-image my-python-base
       NAME      |             URL              | STATUS |         CREATED DATE
------------------------------------------------------------------------------------
  my-python-base | berndtj/my-python-base:0.0.1 | READY  | Fri Jun  8 14:31:48 PDT 2018
```

If the base image is in `ERROR` state, check the reason:

```
$ dispatch get base-image does-not-exist --json | jq .reason
[
  "error when pulling image berndtj/does-not-exist:0.0.1: Error response from daemon: pull access denied for berndtj/does-not-exist, repository does not exist or may require 'docker login'"
]

```

## Images

Images represent the function runtime.  Functions usually have dependencies, such as libraries useful for a given task.
While dependencies can be added to the base-images, doing so via images is more managed and controlled.  The distinction
comes down to roles.  Administrators should manage the catalog of base-images, but developers should manage images. Base
images are built externally via Dockerfiles and may contain anything, whereas images are built via the image service
within Dispatch.  Rather than having access to the Dockerfile, developers may manage image contents through manifests.
By managing images more tightly, Dispatch can provide a catalog of software installed on each image.  This is a
compromise to allow developers to easily add dependencies while maintaining some control and security.

### Adding an Image

In order to create an image which includes the popular python data analysis library, Pandas, first simply create a standard
`requirements.txt` file:

```
$ cat << EOF > requirements.txt
pandas==0.23.0
EOF
```

Images can be added through the `create image` command:

```
$ dispatch create image my-pandas my-python-base --runtime-deps requirements.txt
```

Or via generic `create` (multiple resources may be created at once):

```
$ cat << EOF > images.yaml
kind: Image
name: my-pandas
baseImageName: my-python-base
runtimeDependencies:
  manifest: '@requirements.txt'
tags:
  - key: role
    value: docs
EOF
```

```
$ dispatch create -f images.yaml
```