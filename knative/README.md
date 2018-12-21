# Dispatch the Hard Way (Dispatch without Dispatch)

Dispatch is largely an API proxy to Knative.  This document describes getting Dispatch functionality without using
Dispatch.  Instead use `kubectl` with yaml as well as a popular knative CLI `knctl` for a little convenience.

## Prerequisites

### Download and install `knctl`

Either build or download the latest release from here:
[https://github.com/cppforlife/knctl](https://github.com/cppforlife/knctl)

### Install Knative

Install Knative components to the Kubernetes distribution of your choice.  Hint, you can use `knctl` to install Knative.

#### Install Knative Eventing

Depending on how you're installing Knative, you will likely have to also install the Knative eventing components and eventing sources.
```
$ kubectl apply --filename https://github.com/knative/eventing/releases/download/v0.2.1/release.yaml
$ kubectl apply --filename https://github.com/knative/eventing-sources/releases/download/v0.2.1/release.yaml
```

Lastly, install Kafka and the Kafka channel provisioner.
```
kubectl create namespace kafka
kubectl apply -n kafka -f https://raw.githubusercontent.com/knative/eventing/master/config/provisioners/kafka/broker/kafka-broker.yaml
kubectl apply --filename https://github.com/knative/eventing/releases/download/v0.2.1/kafka.yaml
```

### Deploy and configure Docker repository

It isn't really required to use an interal registry, but this is about keeping the functionality as close to Dispatch
proper as possible.  You *can* skip this section, but the build templates also assume an insecure registry, so some
modifications would be required.

1. Setup some environment vars for use later.
   ```
   $ export DISPATCH_NAMESPACE=default
   $ export RELEASE_NAME=knative-images
   ```

2. Deploy docker repository as an internal repository.
   ```
   $ helm install stable/docker-registry --name ${RELEASE_NAME}
   ```
3. Mark the repository as insecure within Knative.
   ```
   $ ../scripts/configure-knative.sh
   ```
4. Save the Cluster IP address of the repository.
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

## "Dispatch" Images

### Create a "Dispatch" image

There is no analogue to a base-image here, just a docker URL.  However, we can use knative build directly to create
images the same way Dispatch does.  You will notice that the following command uses `kubectl` as opposed to `knctl` even
though `knctl` support build functionality, it's not flexible enough for this use-case.

1. Set the image name.
   ```bash
   export IMAGE_NAME=python-image
   ```
2. Encode the image depenencies (if any).
   ```bash
   $ IMAGE_MANIFEST=$(base64 -i requirements.txt)
   ```
3. Create the build definition.
   ```bash
   $ cat python-image-build.yaml.tmpl | sed "s/IMAGE_NAME/$IMAGE_NAME/g; s/DOCKER_REG_IP/$DOCKER_REG_IP/g; s/IMAGE_MANIFEST/$IMAGE_MANIFEST/g" > python-image.yaml
   ```
4. Apply the image.
   ```
   $ kubectl apply -f python-image.yaml
   build.build.knative.dev "python-image" created
   ```
5. Validate build completion.  This can be done by inspecting the pod status via `kubectl`.
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

## "Dispatch" Functions

### Build a "Dispatch" function

Building a function is a little easier than building an image because it's a use-case that is pretty well supported.
`knctl` makes this a LOT easier as it takes care of uploading the source (you must give it a directory not just a file).

1. Set the function name.
   ```bash
   export FUNCTION_NAME=python-function
   ```

2. Build the function with `knctl`.
   ```bash
   $ knctl build create --build $FUNCTION_NAME \
        --template function-template \
        --template-arg SOURCE_IMAGE=$DOCKER_REG_IP:5000/$IMAGE_NAME \
        --template-arg HANDLER=hello.handle \
        -i $DOCKER_REG_IP:5000/$FUNCTION_NAME \
        -d ../examples/python3
   ```

### Deploy a "Dispatch" function

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

### Execute a "Dispatch" function

Here again `knctl` is a bit help.

```
$ knctl curl --service python-function
Running: curl '-sS' '-H' 'Host: python-function.default.example.com' 'http://35.243.243.55:80'

{"context":{"logs":{"stdout":["Serving on http://0.0.0.0:9000","messages to stdout show up in logs"],"stderr":null}},"payload":{"myField":"Hello, Noone from Nowhere"}}

Succeeded
```

## "Dispatch" Events

`knctl` and other CLI tools don't have rich eventing support yet, so the rest of this guide is pure `kubectl` and yaml.

### Create a "Dispatch" event driver

We can use the combination of a Knative event source, channel and Kubernetes service to create a "Dispatch" event
driver.  In this example we are creating a generic cloudevents event source to which events are pushed (POSTed).

It should be mentioned that we are using the same event driver as Dispatch proper.  Some very [slight modifications](https://github.com/dispatchframework/dispatch-events-cloudevent/commit/f495db3dcf270ebe1c1ef22f78ad4eca7be36ff1) were required (accept the `sink` argument) to make the driver work with Knative directly.

1.  Create a Kafka channel for our event source.
    ```
    $ kubectl apply -f channel.yaml
    channel.eventing.knative.dev "cloudevents-kafka-channel" created
    ```
2.  Create the cloudevents event source (using the Dispatch event driver).  This uses the Knative ContainerSource type.
    ```
    $ kubectl apply -f event-source.yaml
    containersource.sources.eventing.knative.dev "cloudevents-source" created
    ```
3.  Expose the event source.  This step is only required for push type event sources.  It creates a Kubernetes service
    and Istio VirtualService to route external requests to the event driver.  We are reusing the knative shared gateway,
    but you could create a separate gateway for events also.
    ```
    $ kubectl apply -f event-source-service.yaml
    service "cloudevents-source" created
    virtualservice.networking.istio.io "cloudevents-source-route" created
    ```
4.  Create the subscription.  We are going to subscribe the cloudevent source/channel to the python-function created above.  This is not an ideal function as it doesn't expect a cloud event, nor do anything, but the plumbing should work.
    ```
    $ cat subscription.yaml.tmpl | sed "s/FUNCTION_NAME/$FUNCTION_NAME/g" > subscription.yaml
    $ kubectl apply -f subscription.yaml
    subscription.eventing.knative.dev "cloudevents-subscription" created
    ```

### Post a cloudevent

At this point the plumbing is all setup to push event and trigger functions.

1.  Get the IP address of the gateway.
    ```bash
    export INGRESS_IP=$(kubectl get service -n istio-system knative-ingressgateway -o json | jq -r .status.loadBalancer.ingress[].ip)
    ```
2.  Push a sample event.
    ```
    curl -v $INGRESS_IP -H 'Host: cloudevents-source.default.example.com' -H 'Content-Type: application/cloudevents+json' -d @event.json
    * Rebuilt URL to: 35.243.243.55/
    *   Trying 35.243.243.55...
    * TCP_NODELAY set
    * Connected to 35.243.243.55 (35.243.243.55) port 80 (#0)
    > POST / HTTP/1.1
    > Host: cloudevents-source.default.example.com
    > User-Agent: curl/7.54.0
    > Accept: */*
    > Content-Type: application/cloudevents+json
    > Content-Length: 185
    >
    * upload completely sent off: 185 out of 185 bytes
    < HTTP/1.1 200 OK
    < date: Fri, 21 Dec 2018 19:09:47 GMT
    < content-length: 0
    < x-envoy-upstream-service-time: 18
    < server: envoy
    <
    * Connection #0 to host 35.243.243.55 left intact
    ```

3.  Did it work? Because the function doesn't do anything observable, the only real way of checking is the pod status.
    If the function was executed, the function pod would have "woken up".  A much better example would be writing to
    container logs (which can be fetched) or posting to an external service.  That's up to you :).

## "Dispatch" API Gateway

We can reimplement the Dispatch API gateway directly using and Istio gateway and VirtualServices for routes.

### Create an Istio Gateway

We need a couple pieces of infastructure, a Kubernetes service to expose the knative ingress (which already exists) and
an istio gateway used for routing.  The idea behind creating a new gateway and service is that we could confine the
default `knative-ingressgateway` service so as not to expose all Knative services by default extnerally.  Then use our
new service and gateway to expose services more granularly.

```
$ kubectl apply -f api-gateway.yaml
gateway.networking.istio.io "dispatch-api-gateway" created
service "dispatch-api-gateway" created
```

### Create a "Dispatch" API endpoint

To create a specific route we use an Istio VirtualService.

1. Get the IP address of the API gateway Kubernetes service.
```bash
export API_GATEWAY_IP=$(kubectl get service -n istio-system dispatch-api-gateway -o json | jq -r .status.loadBalancer.ingress[].ip)
```
2. Create an `HTTP POST` route with path `/$FUNCTION_NAME` to our function.  This example will also enable CORS (look at the yaml for details).
```
$ cat api-endpoint.yaml.tmpl | sed "s/FUNCTION_NAME/$FUNCTION_NAME/g; s/HTTP_METHOD/POST/g" > api-endpoint.yaml
$ kubectl apply -f api-endpoint.yaml
virtualservice.networking.istio.io "python-function-route" created
```