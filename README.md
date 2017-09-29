# VMware Serverless Platform


## Deploying to Kubernetes

### Building Images

Images are pushed to the `serverless-docker-local.artifactory.eng.vmware.com` repository.  This is a private repository
and requires authentication:

```
docker login serverless-docker-local.artifactory.eng.vmware.com
```

Once logged in, you can now build the images:

```
make images
```

The result should be 3 images:

```
docker images | grep serverless-docker-local.artifactory.eng.vmware.com | grep dev-1 | head -n 3
serverless-docker-local.artifactory.eng.vmware.com/identity-manager                 dev-1506638027      9a855cfa0f34        3 hours ago         155MB
serverless-docker-local.artifactory.eng.vmware.com/image-manager                    dev-1506638027      2249660ef710        3 hours ago         220MB
serverless-docker-local.artifactory.eng.vmware.com/function-manager                 dev-1506638027      b3ab78c500c7        29 hours ago        158MB
```

### Installing the Chart

We use Helm to package and deploy the serverless platform.  Before installing the chart the following prerequisites
must be met:

1. Kubernetes 1.7+ installed and kubectl configured
    - [minikube](https://kubernetes.io/docs/getting-started-guides/minikube/)

2. Configure kubernetes for [private docker registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)
    - `kubectl create secret docker-registry regsecret --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>`

3. Get and install [helm](https://github.com/kubernetes/helm)
    - `brew install kubernetes-helm`
    - `helm init`

At this point, you can install the chart:

```
$ helm install charts/serverless --name demo --set image-manager.image.tag=dev-1506638027 --set function-manager.image.tag=dev-1506638027 --set identity-manager.image.tag=dev-1506638027
NAME:   demo
LAST DEPLOYED: Thu Sep 28 16:12:14 2017
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1/ConfigMap
NAME              DATA  AGE
function-manager  2     1s
image-manager     2     1s

==> v1/Service
NAME                   CLUSTER-IP  EXTERNAL-IP  PORT(S)       AGE
demo-function-manager  10.0.0.48   <nodes>      80:32671/TCP  1s
demo-image-manager     10.0.0.155  <nodes>      80:31318/TCP  1s

==> v1beta1/Deployment
NAME                   DESIRED  CURRENT  UP-TO-DATE  AVAILABLE  AGE
demo-image-manager     1        1        1           0          1s
demo-function-manager  1        1        1           0          1s
```

> NOTE: We are explicitly setting the tag for each image.  The chart defaults to `latest`, which is likely not what you
> want.

You can monitor the deployment with kubectl:

```
$ kubectl get deployment
NAME                          DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
demo-function-manager         1         1         1            1           3m
demo-image-manager            1         1         1            1           3m
$ kubectl get service
NAME                       CLUSTER-IP   EXTERNAL-IP   PORT(S)          AGE
demo-function-manager      10.0.0.48    <nodes>       80:32671/TCP     2m
demo-image-manager         10.0.0.155   <nodes>       80:31318/TCP     2m
```

In order to use the `vs` CLI, set `$HOME/.vs.yaml` to point to the new services:

```
$ minikube service demo-function-manager --url
http://192.168.64.4:32671
$ minikube service demo-image-manager --url
http://192.168.64.4:31318
$ cat << EOF > ~/.vs.yaml
host: 192.168.64.4
port: 8000
organization: vmware
functionManagerPort: 32671
imageManagerPort: 31318
EOF
```

> NOTE: This separate endpoints is temporary.  Once an ingress controller is configured, there will only be one endpoint
> to configure.

At this point the `vs` CLI should work:

```
$ ./bin/vs-darwin get image
Using config file: /Users/bjung/.vs.yaml
  NAME | URL | BASEIMAGE | STATUS | CREATED DATE
--------------------------------------------
```