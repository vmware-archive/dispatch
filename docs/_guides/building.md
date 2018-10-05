---
layout: default
---
# Building and Deploying Dispatch

Now that all of the setup is done, the only thing left is to build and deploy. Ideally, as long as the cluster is up,
you are done with setup.

## Building Images

If you're using the newly built images in minikube locally, you can set your docker client to use the minikube docker daemon:

```bash
eval $(minikube docker-env)
```

By default, the images are prefixed with `vmware` (e.g. `vmware/dispatch-event-driver`). **If you need to use a different registry**, set the
`DOCKER_REGISTRY` environment variable:

```bash
$ echo $DOCKER_REGISTRY
dockerhub-user
```

**If you need to push the new images**, your docker client must be logged into the registry:

```bash
docker login $DOCKER_REGISTRY
```

> **NOTE:** if using docker hub, do not include the registry as an argument (i.e. `docker login`)

To build the images:

```bash
make images
```

To build **and push** the images:

```bash
PUSH_IMAGES=1 make images
```

The result should be 7 images:

```bash
$ docker images | grep $(cat values.yaml | grep tag | cut -d ':' -f 2)
berndtj/dispatch-event-driver                                                                dev-1511831075      285ebca2a1a2        24 minutes ago      106MB
berndtj/dispatch-event-manager                                                               dev-1511831075      537108e4fe2b        24 minutes ago      61.2MB
berndtj/dispatch-api-manager                                                                 dev-1511831075      c735cc5539fe        24 minutes ago      58.8MB
berndtj/dispatch-secrets                                                                dev-1511831075      71c284321f6b        25 minutes ago      124MB
berndtj/dispatch-function-manager                                                            dev-1511831075      ef55b922b661        25 minutes ago      65.7MB
berndtj/dispatch-identity-manager                                                            dev-1511831075      cb045ec07d14        25 minutes ago      60.1MB
berndtj/dispatch-image-manager                                                               dev-1511831075      4efd00b0318f        25 minutes ago      61.4MB
```

## Installing the Dispatch Chart

First create an docker authorization token for openfaas:
```bash
export OPENFAAS_AUTH=$(echo '{"username":"bjung","password":"********","email":"bjung@vmware.com"}' | base64)
```

The first time the chart is installed, chart dependencies must be fetched:

```bash
cd ./charts/dispatch
helm dep up
cd -
```

A values.yaml was created as an artifact of `make images`.
```bash
$ helm install ./charts/dispatch --name=dev-dispatch --namespace dispatch \
  -f values.yaml \
  --set function-manager.faas.openfaas.registryAuth=$OPENFAAS_AUTH \
  --set function-manager.faas.openfaas.imageRegistry=$DOCKER_REGISTRY \
  --set global.host=$DISPATCH_HOST \
  --set global.image.host=$DOCKER_REGISTRY \
  --set oauth2-proxy.app.clientID=$DISPATCH_OAUTH_CLIENT_ID  \
  --set oauth2-proxy.app.clientSecret=$DISPATCH_OAUTH_CLIENT_SECRET \
  --set oauth2-proxy.app.cookieSecret=$DISPATCH_OAUTH_COOKIE_SECRET \
  --debug
```

You can monitor the deployment with kubectl:
```bash
$ kubectl -n dispatch get deployment
NAME                              DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
dev-dispatch-api-manager        1         1         1            1           1h
dev-dispatch-echo-server        1         1         1            1           1h
dev-dispatch-event-manager      1         1         1            1           1h
dev-dispatch-function-manager   1         1         1            1           1h
dev-dispatch-identity-manager   1         1         1            1           1h
dev-dispatch-image-manager      1         1         1            1           1h
dev-dispatch-oauth2-proxy       1         1         1            1           1h
dev-dispatch-rabbitmq           1         1         1            1           1h
dev-dispatch-secrets       1         1         1            1           1h
$ kubectl -n dispatch get service
NAME                              CLUSTER-IP   EXTERNAL-IP   PORT(S)                                 AGE
dev-dispatch-api-manager        10.0.0.179   <none>        80/TCP                                  1h
dev-dispatch-echo-server        10.0.0.58    <none>        80/TCP                                  1h
dev-dispatch-event-manager      10.0.0.13    <none>        80/TCP                                  1h
dev-dispatch-function-manager   10.0.0.145   <none>        80/TCP                                  1h
dev-dispatch-identity-manager   10.0.0.245   <none>        80/TCP                                  1h
dev-dispatch-image-manager      10.0.0.188   <none>        80/TCP                                  1h
dev-dispatch-oauth2-proxy       10.0.0.51    <none>        80/TCP,443/TCP                          1h
dev-dispatch-rabbitmq           10.0.0.180   <none>        4369/TCP,5672/TCP,25672/TCP,15672/TCP   1h
dev-dispatch-secrets       10.0.0.249   <none>        80/TCP                                  1h
```

## Test your Deployment

### Test with ``curl``

Check that the ingress controller is routing to the identity manager
```bash
$ curl -k https://dispatch.local/oauth2/start
<a href="https://github.com/login/oauth/authorize?approval_prompt=force&amp;client_id=8e3894cb0199f8342f0a&amp;redirect_uri=https%3A%2F%2Fdispatch.local%2Foauth2%2Fcallback&amp;response_type=code&amp;scope=openid%2C+user%3Aemail&amp;state=13939a602fb6704121766b0f5ba588cb%3A%2F">Found</a>.
```

### Test with the ``dispatch`` CLI

If you haven't already, build the `dispatch` CLI:

```bash
make cli-darwin
ln -s `pwd`/bin/dispatch-darwin /usr/local/bin/dispatch
```

In order to use the `dispatch` CLI, set `$HOME/.dispatch.yaml` to point to the new services:

```bash
$ cat << EOF > ~/.dispatch.yaml
host: $DISPATCH_HOST
port: 443
organization: vmware
cookie: ""
EOF
```

At this point the `dispatch` CLI should work:

```bash
$ dispatch login
You have successfully logged in, cookie saved to /Users/bjung/.dispatch.yaml
$ dispatch get image
Using config file: /Users/bjung/.dispatch.yaml
  NAME | URL | BASEIMAGE | STATUS | CREATED DATE
```
