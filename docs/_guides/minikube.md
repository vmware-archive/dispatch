---
layout: default
---
# Installing Kubernetes via Minikube for Dispatch

We recommend using minikube to install and manage local kubernetes clusters for Dispatch development, or for simply
getting up and running quickly. If you have VMware Fusion or Workstation, we recommend [minikube via a Linux
VM](#minikube_via_linux_vm) as the easiest way to get started.  This is also the only supported way to run Dispatch on
Windows.

## Minikube on Mac

Installing minikube is on mac easy.  The following installs the latest minikube on a mac.  See the official [minikube
repository](https://github.com/kubernetes/minikube) for more details.  This guide assumes you have
[homebrew](https://brew.sh) installed.

1. Install minikube via brew:

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

2. Install the VM driver of your choice.  We recommend hyperkit (unless you use a VPN):

```bash
$ curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit && chmod +x docker-machine-driver-hyperkit && sudo mv docker-machine-driver-hyperkit /usr/local/bin/ && sudo chown root:wheel /usr/local/bin/docker-machine-driver-hyperkit && sudo chmod u+s /usr/local/bin/docker-machine-driver-hyperkit
```

> **Note:** If you require a VPN for connectivity, you should choose either a
[vmwarefusion](https://www.vmware.com/products/fusion.html),[vmwareworkstation](https://www.vmware.com/products/workstation-pro.html) or [virtualbox](https://www.virtualbox.org) as the VM
driver, as hyperkit will bypass the VPN.  This means that Dispatch (or any other application) deployed on your cluster
will not have access to the private network you are connecting to with the VPN.

3. Create a kubernetes cluster:

```bash
$ minikube start --vm-driver=hyperkit --disk-size=50g --memory=6144
Starting local Kubernetes v1.10.0 cluster...
Starting VM...
Getting VM IP address...
Moving files into cluster...
Setting up certs...
Connecting to cluster...
Setting up kubeconfig...
Starting cluster components...
Kubectl is now configured to use the cluster.
Loading cached images from config file.
```

4. Verify installation:

```bash
$ kubectl get pods --all-namespaces
NAMESPACE     NAME                                    READY     STATUS    RESTARTS   AGE
kube-system   etcd-minikube                           1/1       Running   0          25s
kube-system   kube-addon-manager-minikube             1/1       Running   0          17s
kube-system   kube-apiserver-minikube                 1/1       Running   0          22s
kube-system   kube-controller-manager-minikube        1/1       Running   0          20s
kube-system   kube-dns-86f4d74b45-cklkj               3/3       Running   0          1m
kube-system   kube-proxy-f6mqs                        1/1       Running   0          1m
kube-system   kube-scheduler-minikube                 1/1       Running   0          28s
kube-system   kubernetes-dashboard-5498ccf677-bj8l2   1/1       Running   0          1m
kube-system   storage-provisioner                     1/1       Running   0          59s
```

5. Install and initialize Helm:

```bash
$ brew install kubernetes-helm
$ helm init
$HELM_HOME has been configured at /Users/bjung/.helm.

Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.
Happy Helming!
```

## Minikube via Linux VM

In order to provide a means of getting a Dispatch environement ready that "just works", we've prepared a VM which
includes just about everything needed to deploy kubernetes and install Dispatch.  By following these steps, you
should have an environemnt ready for Dispatch quickly and repeatably without installing any dependencies on your
local machine.  This does assume that you have VMware Fusion, Workstation or VirtualBox installed.

1. Download [OVA](https://s3-us-west-2.amazonaws.com/dispatch-imgs/VMware-ubuntu-kubernetes-dispatch.ova)

2. Import the OVA.  Instructions vary depending on which virtualization software you are using.

3. Login to console with credentials: kube/kube
    - Get IP address ifconfig (look for the 10.x.x.x address)

4. SSH to the deployed VM with same credentials (`ssh vmware@10.x.x.x`):

```bash
$ ssh kube@10.64.236.81
kube@10.64.236.81's password:
Welcome to Ubuntu 16.04.5 LTS (GNU/Linux 4.4.0-131-generic x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage
New release '18.04.1 LTS' available.
Run 'do-release-upgrade' to upgrade to it.

Last login: Wed Sep  5 00:43:47 2018
kube@VMware-ubuntu-kubernetes:~$
```

5. Run `sudo setup_dispatch` (there will be some warnings, which you can safely ignore):

```bash
$ sudo setup_dispatch
[sudo] password for kube:
>>>> Running kubeadm
[init] using Kubernetes version: v1.11.2
[preflight] running pre-flight checks
I0901 00:08:44.675493    1381 kernel_validator.go:81] Validating kernel version
I0901 00:08:44.675776    1381 kernel_validator.go:96] Validating kernel config
[preflight/images] Pulling images required for setting up a Kubernetes cluster
[preflight/images] This might take a minute or two, depending on the speed of your internet connection
...
```

6. Verify installation:

```bash
$ kubectl -n kube-system get pods
NAME                                                     READY     STATUS    RESTARTS   AGE
coredns-78fcdf6894-mz8xq                                 1/1       Running   0          7m
coredns-78fcdf6894-vtsc4                                 1/1       Running   0          7m
etcd-vmware-ubuntu-kubernetes                            1/1       Running   0          6m
ingress-nginx-ingress-controller-9fbc9b487-5dc9k         1/1       Running   0          7m
ingress-nginx-ingress-default-backend-677b99f864-694f8   1/1       Running   0          7m
kube-apiserver-vmware-ubuntu-kubernetes                  1/1       Running   0          6m
kube-controller-manager-vmware-ubuntu-kubernetes         1/1       Running   0          6m
kube-flannel-ds-w8rwz                                    1/1       Running   0          7m
kube-proxy-ljgpd                                         1/1       Running   0          7m
kube-scheduler-vmware-ubuntu-kubernetes                  1/1       Running   0          6m
tiller-deploy-56c4cf647b-8pv4b                           1/1       Running   0          7m
```