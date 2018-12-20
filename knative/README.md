# Dispatch the Hard Way (Dispatch without Dispatch)

Dispatch is largely an API proxy to Knative.  This document describes getting Dispatch functionality without using
Dispatch.  Instead use `kubectl` with yaml as well as a popular knative CLI `knctl` for a little convenience.

## Prerequisites

### Install Knative

Install Knative components to the Kubernetes distribution of your choice.

### Deploy and configure Docker repository

It isn't really required to use an interal registry, but this is about keeping the functionality as close to Dispatch
proper as possible.  You *can* skip this section, but the build templates also assume an insecure registry, so some
modifications would be required.

1. Setup some environment vars for use later
   ```
   $ export DISPATCH_NAMESPACE=default
   $ export RELEASE_NAME=knative-images
   ```

2. Deploy docker repository as an internal repository
   ```
   $ helm install stable/docker-registry --name ${RELEASE_NAME}
   ```
3. Mark the repository as insecure within Knative
   ```
   $ ../scripts/configure-knative.sh
   ```
4. Save the Cluster IP address of the repository
   ```
   $ export DOCKER_REG_IP=$(kubectl get service --namespace ${DISPATCH_NAMESPACE} ${RELEASE_NAME}-docker-registry -o jsonpath="{.spec.clusterIP}")
   $ echo $DOCKER_REG_IP
   10.7.248.157
   ```

### Install build templates

Knative build templates build both images and functions.

```
$ kubectl apply -f build-template.yaml
buildtemplate.build.knative.dev "image-template" configured
buildtemplate.build.knative.dev "function-template" configured
```

## Create a "Dispatch" Image

There is no analogue to a base-image here, just a docker URL.  However, we can use knative build directly to create
images the same way Dispatch does.  You will notice that the following command uses `kubectl` as opposed to `knctl` even
though `knctl` support build functionality, it's not flexible enough for this use-case.

1. Set the image name
   ```bash
   export IMAGE_NAME=python-image
   ```
2. Encode the image depenencies (if any)
   ```bash
   $ IMAGE_MANIFEST=$(base64 -i requirements.txt)
   ```
3. Create the build definition
   ```bash
   $ cat python-image-build.yaml.tmpl | sed "s/IMAGE_NAME/$IMAGE_NAME/g; s/DOCKER_REG_IP/$DOCKER_REG_IP/g; s/IMAGE_MANIFEST/$IMAGE_MANIFEST/g" > python-image.yaml
   ```
4. Apply the image
   ```
   $ kubectl apply -f python-image.yaml
   build.build.knative.dev "python-image" created
   ```
5. Validate build completion.  This can be done by inspecting the pod status via `kubectl`:
   ```
   $ kubectl get pods | grep $IMAGE_NAME
   python-image-f76nf                                0/1       Completed   0          1m
   ```
   Or more conveniently with `knctl`:
   ```
   $ knctl build list
   Builds in namespace 'default'

   Name                    Succeeded  Age
   python-image            true       2m

   1 builds

   Succeeded
   ```
   If something went wrong use `knctl` to get the build logs:
   ```
   $ knctl build show --build python-image
   Build 'python-image'

   Name          python-image
   Timeout       10m0s
   Started at    2018-12-20T10:07:33-08:00
   Completed at  2018-12-20T10:07:58-08:00
   Succeeded     true
   Age           4m

   Conditions

   Type       Status  Age  Reason  Message
   Succeeded  True    3m   -       -

   Watching build logs...
   ... (lots of debug output)
   ```

## Build a "Dispatch" Function

Building a function is a little easier than building an image because it's a use-case that is pretty well supported.
`knctl` makes this a LOT easier as it takes care of uploading the source (you must give it a directory not just a file).

1. Set the function name
   ```bash
   FUNCTION_NAME=python-function
   ```

2. Build the function with `knctl`
   ```bash
   $ knctl build create --build $FUNCTION_NAME \
        --template function-template \
        --template-arg SOURCE_IMAGE=$DOCKER_REG_IP:5000/$IMAGE_NAME \
        --template-arg HANDLER=hello.handle \
        -i $DOCKER_REG_IP:5000/$FUNCTION_NAME \
        -d ../examples/python3
   ```

## Deploy a "Dispatch" Function

```
$ knctl deploy -s $FUNCTION_NAME --image $DOCKER_REG_IP:5000/$FUNCTION_NAME
```

> NOTE: This command hangs waiting for logs, just `CTRL-C` out.

Validate the service was deployed successfully.
```
$ knctl service list
Services in namespace 'default'

Name             Domain                               Annotations  Conditions  Age
python-function  python-function.default.example.com  -            3 OK / 3    1m

1 services

Succeeded
```

## Execute a "Dispatch" Function

Here again `knctl` is a bit help.

```
$ knctl curl --service python-function
Running: curl '-sS' '-H' 'Host: python-function.default.example.com' 'http://35.243.243.55:80'

{"context":{"logs":{"stdout":["Serving on http://0.0.0.0:9000","messages to stdout show up in logs"],"stderr":null}},"payload":{"myField":"Hello, Noone from Nowhere"}}

Succeeded
```