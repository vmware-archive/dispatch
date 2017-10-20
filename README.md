# VMware Serverless Platform


## Deploying to Kubernetes

### Minikube

Our charts assume RBAC is enabled, therefore create a cluster via the following command:

```
minikube start --vm-driver=xhyve --extra-config=apiserver.Authorization.Mode=RBAC
```

Next add the necessary cluster-role bindings:

```
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```

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

### Import self-signed TLS certificates into Kubernetes secret

To be able to securely connect to the serverless platfrom, we need to set up a TLS certificate.

```
$ ./scripts/make-ssl-crt.sh
```

With the above command, we create a self-signed certificates named ``serverless-vmware-tls``, and then import it into kubernetes secret store.

Note this is only required ONCE per a kubernetes cluster.

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

4. Install nginx ingress controller chart

A nginx ingress controller is required for the serverless platfrom, please install it with
```
$ helm upgrade --install demo-ingress-ctrl ./charts/nginx-ingress --namespace=kube-system
```

5. Install serverless chart

First create an docker authorization token for openfaas:

```
OPENFAAS_AUTH=$(echo '{"username":"bjung","password":"********","email":"bjung@vmware.com"}' | base64)
```

A values.yaml is created as an artifact of `make images`.
```
$ helm upgrade --install demo ./charts/serverless -f values.yaml --set function-manager.faas.openfaas.registryAuth=$OPENFAAS_AUTH
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


### Test your Deployment

Add an entry in your ``/etc/hosts``

```
<k8s_ip> serverless.vmware.com
```
Please replace ``<kubernetes_ip>`` with your kubernetes cluster ip, if you are using minikube, using command ``minikube ip`` to get it.

#### Test with ``curl``

Get a list of image in your serverless host with
```
$ curl http://serverless.vmware.com/v1/image
[]
```

#### Test with ``vs`` CLI

In order to use the `vs` CLI, set `$HOME/.vs.yaml` to point to the new services:

```
$ cat << EOF > ~/.vs.yaml
host: serverless.vmware.com
port: 443
organization: vmware
cookie: ""
EOF
```

At this point the `vs` CLI should work:

```
$ ./bin/vs-darwin get image
Using config file: /Users/bjung/.vs.yaml
  NAME | URL | BASEIMAGE | STATUS | CREATED DATE
--------------------------------------------
```