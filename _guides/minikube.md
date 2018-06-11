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
[vmwarefusion](https://www.vmware.com/products/fusion.html), or [virtualbox](https://www.virtualbox.org) as the VM
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
NAMESPACE     NAME                          READY     STATUS    RESTARTS   AGE
kube-system   kube-addon-manager-minikube   1/1       Running   0          54s
kube-system   kube-dns-545bc4bfd4-mjljc     3/3       Running   0          43s
kube-system   kube-proxy-nmzhd              1/1       Running   0          43s
kube-system   kubernetes-dashboard-5fllx    1/1       Running   2          41s
kube-system   storage-provisioner           1/1       Running   0          42s
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

1. Download [OVA](https://s3-us-west-2.amazonaws.com/dispatch-imgs/Ubuntu_16.04.2_server_amd64_dispatch.ova)

2. Import the OVA.  Instructions vary depending on which virtualization software you are using.

3. Login to console with credentials: vmware/vmware
    - Get IP address ifconfig (look for the 10.x.x.x address)

4. SSH to the deployed VM with same credentials (`ssh vmware@10.x.x.x`):

```bash
$ ssh vmware@10.52.72.124
vmware@10.52.72.124's password:
Welcome to Ubuntu 16.04.2 LTS (GNU/Linux 4.4.0-62-generic x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage

167 packages can be updated.
90 updates are security updates.


Last login: Thu Feb  8 06:07:13 2018
vmware@pek2-office-9th-10-117-171-69:~$
```

5. Run `./install-minikube.sh` (there will be some warnings, which you can safely ignore):

```bash
$ ./install-minikube.sh
[sudo] password for vmware:
Starting local Kubernetes v1.9.0 cluster...
Starting VM...
Getting VM IP address...
Moving files into cluster...
Setting up certs...
Connecting to cluster...
Setting up kubeconfig...
Starting cluster components...
Kubectl is now configured to use the cluster.
...
```

6. Verify installation:

```bash
$ kubectl -n kube-system get pods
NAME                                    READY     STATUS    RESTARTS   AGE
etcd-minikube                           1/1       Running   0          5m
kube-addon-manager-minikube             1/1       Running   0          5m
kube-apiserver-minikube                 1/1       Running   0          5m
kube-controller-manager-minikube        1/1       Running   0          5m
kube-dns-6f4fd4bdf-n58vd                3/3       Running   0          6m
kube-proxy-5s52t                        1/1       Running   0          6m
kube-scheduler-minikube                 1/1       Running   0          5m
kubernetes-dashboard-77d8b98585-49wrf   1/1       Running   1          6m
storage-provisioner                     1/1       Running   0          6m
```

7. Initialize Helm:

```bash
$ helm init
Creating /home/vmware/.helm
Creating /home/vmware/.helm/repository
Creating /home/vmware/.helm/repository/cache
Creating /home/vmware/.helm/repository/local
Creating /home/vmware/.helm/plugins
Creating /home/vmware/.helm/starters
Creating /home/vmware/.helm/cache/archive
Creating /home/vmware/.helm/repository/repositories.yaml
Adding stable repo with URL: https://kubernetes-charts.storage.googleapis.com
Adding local repo with URL: http://127.0.0.1:8879/charts
$HELM_HOME has been configured at /home/vmware/.helm.

Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.
Happy Helming!
```

As a convenience, the Dispatch project has been cloned into `~/code/dispatch` on the VM.
