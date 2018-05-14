---
layout: default
---
# Standard Installation

## Prerequisites

Dispatch depends on Kubernetes as its deployment infrastructure.  Any
"standard" kubernetes should be supported.  Development is largely done on
minikube as it allows for easy local deployment.  This guide will focus on
two deployment models, minikube and vanilla hosted kubernetes 1.7+.

## Kubernetes

A Kubernetes 1.7+ cluster with PersistentVolume(PV) provisioner support is required for installing the Dispatch project.

### Minikube

Installing minikube is easy.  The following installs the latest minikube on a
mac.  See the official [minikube
repository](https://github.com/kubernetes/minikube) for more details:

```bash
$ brew cask install minikube
==> Satisfying dependencies
All Formula dependencies satisfied.
==> Downloading https://storage.googleapis.com/minikube/releases/v0.24.1/minikube-darwin-amd64
######################################################################## 100.0%
==> Verifying checksum for Cask minikube
==> Installing Cask minikube
==> Linking Binary 'minikube-darwin-amd64' to '/usr/local/bin/minikube'.
🍺  minikube was successfully installed!
```

Install the VM driver of your choice.  We recommend hyperkit:

```bash
$ curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit && chmod +x docker-machine-driver-hyperkit && sudo mv docker-machine-driver-hyperkit /usr/local/bin/ && sudo chown root:wheel /usr/local/bin/docker-machine-driver-hyperkit && sudo chmod u+s /usr/local/bin/docker-machine-driver-hyperkit
```

Create a kubernetes cluster:

```bash
$ minikube start --vm-driver=hyperkit --bootstrapper=kubeadm --disk-size=50g --memory=4096 --kubernetes-version=v1.8.1
Starting local Kubernetes v1.8.0 cluster...
Starting VM...
Downloading Minikube ISO
 140.01 MB / 140.01 MB [============================================] 100.00% 0s
Getting VM IP address...
Moving files into cluster...
Downloading kubeadm v1.8.0
Downloading kubelet v1.8.0
Finished Downloading kubelet v1.8.0
Finished Downloading kubeadm v1.8.0
Setting up certs...
Connecting to cluster...
Setting up kubeconfig...
Starting cluster components...
Kubectl is now configured to use the cluster.
Loading cached images from config file.
```

Verify installation:

```bash
$ kubectl get pods --all-namespaces
NAMESPACE     NAME                          READY     STATUS    RESTARTS   AGE
kube-system   kube-addon-manager-minikube   1/1       Running   0          54s
kube-system   kube-dns-545bc4bfd4-mjljc     3/3       Running   0          43s
kube-system   kube-proxy-nmzhd              1/1       Running   0          43s
kube-system   kubernetes-dashboard-5fllx    1/1       Running   2          41s
kube-system   storage-provisioner           1/1       Running   0          42s
```

### Hosted Kubernetes

There are a variety of methods for installing kubernetes in a hosted environment (including a private datacenter).  This
is beyond the scope of this guide.  However, the dispatch charts depend on RBAC being enabled on the cluster. Check the
documentation of your Kubernetes deployer on how to enable RBAC authorization.

## Helm

[Helm](https://helm.sh) is the package manager for Kubernetes. Dispatch as well as many dependencies are installed and
managed via helm charts. Before anything can be installed, helm must be setup.

### RBAC and Helm

Recent versions of Kubernetes have introduced roles and service accounts. Depending on how your Kubernetes cluster is
configured, one or more of the following may be required:

#### Add the cluster-admin clusterrole (required for Kubernetes on OpenStack - VIOK):

```bash
$ kubectl create clusterrole cluster-admin --verb=get,list,watch,create,delete,update,patch --resource=deployments,services,secrets
$ kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```

> **Note:** some kubernetes deployments come with ``cluster-admin`` already, if so, you could skip the ``kubectl create
> clusterrole`` command

#### Add the tiller service account (required for clusters created via Kops - AWS):

```bash
kubectl -n kube-system create serviceaccount tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
```

> **Note:** Although required for clusters created with Kops, creating a specific service account for tiller should be
> OK, and even encouraged, for all deployments.

### Download Helm and install tiller:

```bash
$ brew install kubernetes-helm
# If not using the tiller service account, simply run `helm init`
$ helm init --service-account tiller
$HELM_HOME has been configured at /Users/bjung/.helm.

Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.
Happy Helming!
```

## Set a Few Environment Variables

Define the following environment variables.  The actual values can be whatever
you like:

```bash
export DISPATCH_HOST=dispatch.local
export DISPATCH_NAMESPACE=dispatch
```

## Configure Image Registry

Dispatch pulls and pushes images as part of the image manager component.  In order to do so, image registry credentials
must be [configured](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/):

```bash
$ kubectl create secret docker-registry regsecret --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```

Docker hub example:

```bash
$ kubectl create secret docker-registry regsecret --docker-server='https://index.docker.io/v1/' --docker-username=dockerhub-user --docker-password='...' --docker-email=dockerhub-user@gmail.com
```

### Image Pull Secrets

Depending on the image registry, secrets may be required to pull images.  **Only OpenFaaS supports image
pull secrets**.  To configure image pull secrets simply configure installation:

```yaml
apiGateway:
  ...
dispatch:
  ...
  imagePullSecret: pull-secret
  imageRegistry:
    name: some-repo.example.com
    username: username
    password: password
```

The installer will take care of creating a secret and associating it to OpenFaaS.

## Import self-signed TLS certificates into Kubernetes secret

To be able to securely connect to Dispatch, we need to set up a TLS certificate.

```bash
$ ./scripts/make-ssl-crt.sh $DISPATCH_NAMESPACE $DISPATCH_HOST
```

With the above command, we create a self-signed certificates named ``dispatch-tls``, and ``api-dispatch-tls``, and then
import it into kubernetes secret store.

Note this is only required ONCE per a kubernetes cluster.


## Install Ingress Controller

A nginx ingress controller is required for Dispatch, please install it with (this will be installed into the kube-system
namespace).

For minikube:
```bash
$ helm install ./charts/nginx-ingress --namespace=kube-system --name=ingress --set=controller.service.type=NodePort --wait
```

For hosted kubernetes:
```bash
$ helm install ./charts/nginx-ingress --namespace=kube-system --name=ingress --wait
```

Get IP address to ingress controller by:

For minikube:
```bash
$ minikube ip
```

For hosted kubernetes:
```bash
$ kubectl describe service ingress-nginx-ingress-controller --namespace=kube-system
```
You should find the public IP from the ``LoadBalancer`` section,

Edit `etc/hosts` and add/edit a record for this Dispatch deployment:

```bash
$ cat /etc/hosts
##
# Host Database
#
# localhost is used to configure the loopback interface
# when the system is booting.  Do not change this entry.
##
127.0.0.1	localhost
255.255.255.255	broadcasthost
::1             localhost
192.168.64.7	dispatch.local
```

## Install FaaS (OpenFaaS)

The framework is architected to support multiple FaaS implementations. Presently
[OpenFaaS](https://github.com/openfaas/faas) is the preferred FaaS:

```bash
$ helm install --namespace=openfaas --name=openfaas --set=exposeServices=false ./charts/openfaas --wait
```

## Install Api-gateway (Kong)

For minikube:
```bash
helm install --namespace=kong --name=api-gateway --set services.proxyService.type=NodePort ./charts/kong --wait
```

For hosted kubernetes:
```bash
helm install --namespace=kong --name=api-gateway ./charts/kong --wait
```

