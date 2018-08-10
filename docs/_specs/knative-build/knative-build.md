# Knative Builds and Build Templates

We can replace the way we currently build docker images and functions with Knative build and build templates. This will replace the need for `Docker in Docker` in the `image manager` and `function manager`.

## Proposal

Create two build templates: one for images and one for functions. The build template will utilize `kaniko` to build the docker image and push it to the given registry. We can utilize the existing base images we use now without any changes.

## Images

### ***Image Build Template***

Here is a possible image template:
```yaml
apiVersion: build.knative.dev/v1alpha1
kind: BuildTemplate
metadata:
  name: image-template
  namespace: dispatch
spec:
  parameters:
  - name: DESTINATION
    description: The destination to push the image (image name)
  - name: BASE_IMAGE
    description: The base image which this image is built from
  - name: DOCKERFILE
    description: Path to the Dockerfile to build
    default: Dockerfile
  - name: SYSTEM_PACKAGES_FILE
    description: Path to file with system dependencies
    default: system-packages.txt
  - name: PACKAGES_FILE
    description: Path to file with runtime dependencies
    default: packages.txt

  steps:
  - name: build-and-push
    image: gcr.io/kaniko-project/executor:v0.3.0
    args:
    - --dockerfile=${DOCKERFILE}
    - --destination=${DESTINATION}
    #- --insecure-skip-tls-verify (not yet implemented)
    - --build-arg=BASE_IMAGE=${BASE_IMAGE}
    - --build-arg=SYSTEM_PACKAGES_FILE=${SYSTEM_PACKAGES_FILE}
    - --build-arg=PACKAGES_FILE=${PACKAGES_FILE}
```

This build template accepts multiple arguments that will be used to configure the build. The `DESTINATION` and `BASE_IMAGE` arguments are required. `Kaniko` expects the docker context (`Dockerfile`, `system-packages.txt`, `packages.txt`) to be in the `/workspace` directory, which should be provided by the sources in the knative build.

Insecure registries for kaniko are not supported yet. It is currently [WIP](https://github.com/GoogleContainerTools/kaniko/pull/169).

### ***Image Build Example***

When creating an image, Dispatch will be responsible for using the build template to create a Knative build. Dispatch should use the Knative build api to construct a Knative build object with the build template and send it to kubernetes to start the build.

For example, for a `nodejs` image, it may construct a build object with the given yaml:

```yaml
apiVersion: build.knative.dev/v1alpha1
kind: Build
metadata:
  name: nodejs
  namespace: dispatch
spec:
  #serviceAccountName: docker-reg-auth
  source:
    custom:
      image: dispatchframework/nodejs-base:0.0.12 # base-image
      command: ['cp']
      args: ['${IMAGE_TEMPLATE_DIR}/Dockerfile', '/workspace']
      env:
      - name: 'IMAGE_TEMPLATE_DIR'
        value: /image-template
    custom:
      image: echo # see echo image
      args: ['-n', '${SYSTEM_PACKAGES_CONTENT}', '|',  'base64', '--decode', '>', '/workspace/${SYSTEM_PACKAGES_FILE}']
      env:
      - name: 'SYSTEM_PACKAGES_CONTENT'
        value: cat system-packages.txt | base64 # supplied by image-manager
      - name: 'SYSTEM_PACKAGES_FILE'
        value: system-packages.txt
    custom:
      image: echo # see echo image
      args: ['-n', '${PACKAGES_CONTENT}', '|',  'base64', '--decode', '>', '/workspace/${PACKAGES_FILE}']
      env:
      - name: 'PACKAGES_CONTENT'
        value: cat packages.txt | base64 # supplied by image-manager
      - name: 'PACKAGES_FILE'
        value: packages.txt
  template:
    name: image-template
    arguments:
    - name: DESTINATION
      value: host.docker.internal:5000/575f1f5e-abbb-4056-8879-9bf5694c1f0d:latest
    - name: BASE_IMAGE
      value: dispatchframework/nodejs-base:0.0.12
```

echo image:
```yaml
# echo image

FROM vmware/photon2:20180424

ENTRYPOINT [ "echo" ]
```

The dispatch `image-manager` will be responsible for constructing this knative build object. We utilize the `image-template` to build and push the image. 

Here we have three sources:
* `Dockerfile` - This utilizes the base-image and extracts the image-template Dockerfile and copies it into `/workspace`
* `system-packages.txt` - This utilizes an echo image to base64 decode a system-packages.txt file into `/workspace`. Since this file is expected to be small, we can base64 encode it and pass it as an argument to the container. Otherwise, we would probably need some intermediary storage for this file.
* `packages.txt` - Same as `system-packages.txt`

Note:
* Can't use the docker image label `io.dispatchframework.image.template` (we currently use) to get the image-template directory since doing so would require running `docker inspect`.

Docker registry auth can be easily added by creating a service account with a docker registry secret and specifying it in the build spec. See [here](https://github.com/knative/docs/blob/master/build/auth.md).

This will support the way we currently build images but it can also be easily extended to support additional sources. For example, we could support storing the `packages.txt` file in `git` or `gcs` which already has built in support. We can also easily add support for `s3` by creating a container which pulls the file from `s3` and writes it into `/workspace`.

## Functions

### ***Function Build Template***

```yaml
apiVersion: build.knative.dev/v1alpha1
kind: BuildTemplate
metadata:
  name: function-template
  namespace: dispatch
spec:
  parameters:
  - name: DESTINATION
    description: The destination to push the image (image name)
  - name: IMAGE
    description: The image which this function image is built from
  - name: HANDLER
    description: The fully-qualified function implementation name
  - name: DOCKERFILE
    description: Path to the Dockerfile to build
    default: Dockerfile

  steps:
  - name: build-and-push
    image: gcr.io/kaniko-project/executor:v0.3.0
    args:
    - --dockerfile=${DOCKERFILE}
    - --destination=${DESTINATION}
    #- --insecure-skip-tls-verify (not yet implemented)
    - --build-arg=IMAGE=${IMAGE}
    - --build-arg=HANDLER=${HANDLER}
```

Similar to the image build template. Requires `DESTINATION`, `IMAGE`, and `HANDLER` arguments.

### ***Function Build Example***

Similar to the image build example, the Dispatch `function-manager` will be responsible for creating the knative build object to create function images. This build should be embedded into a Knative configuration as a Knative service (functions will be a service with a configuration and route).

Here is an example for a `nodejs` function build:
```yaml
apiVersion: build.knative.dev/v1alpha1
kind: Build
metadata:
  name: hello-js
  namespace: dispatch
spec:
  #serviceAccountName: docker-reg-auth
  #serviceAccountName: git-auth
  source:
    git:
      url: https://github.com/vmware/dispatch-nodejs-example.git
      revision: master
    custom:
      image: host.docker.internal:5000/575f1f5e-abbb-4056-8879-9bf5694c1f0d:latest # image
      command: ['cp']
      args: ['${FUNCTION_TEMPLATE_DIR}/Dockerfile', '/workspace']
      env:
      - name: 'FUNCTION_TEMPLATE_DIR'
        value: /function-template
  template:
    name: function-template
    arguments:
    - name: DESTINATION
      value: host.docker.internal:5000/func-7d343a62-3acb-466d-9560-56c45255eaed:latest
    - name: IMAGE
      value: host.docker.internal:5000/575f1f5e-abbb-4056-8879-9bf5694c1f0d:latest
    - name: HANDLER
      value: hello.js
```

We utilize the function-template to build and push the function image to the given registry.

Here we have two sources:
* `Source code` - This is supplied using github and copies the source code into `/workspace`. The repository can be private if provided a git service account.
* `Dockerfile` - This utilitzes the image (i.e. `nodejs`) and extracts the function-template Dockerfile and copies it into `/workspace`

Note:
* Similarly, can't use the docker image label `io.dispatchframework.function.template` (we currently use) to get the function-template directory

Docker registry auth and git auth can be easily added by specifying service accounts in the build spec. See [here](https://github.com/knative/docs/blob/master/build/auth.md). One thing to note is the need for special metadata annotations in the secrets (`build.knative.dev/docker-` and `build.knative.dev/git-`).

Similary, the source can be easily extended to support different types. `gcs` already has built in support and we can easily create a custom container to pull a source from `s3` or somewhere else.