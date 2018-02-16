---
layout: default
---
# Troubleshooting Dispatch


If you're running Dispatch on macOS, there are some possible issues with `dispatch install --file config.yaml` you may hit:

##### Issue:

```
Configuration error: Key: 'installConfig.DispatchConfig.ImageRegistry.Name' Error:Field validation for 'Name' failed on the 'required' tag
```

###### Solution:

Check your Dispatch CLI version. Since 1.6+ `imageRegistry` is not required and an embedded registry will be deployed to the cluster.


##### Issue:

```
Error installing nginx-ingress chart: Error: could not find tiller
```

###### Solution:

```
helm init --upgrade

kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}
```


##### Issue:

```
Error installing nginx-ingress chart: Release "ingress" does not exist. Installing it now.
Error: release ingress failed: clusterroles.rbac.authorization.k8s.io "ingress-nginx-ingress" is forbidden: attempt to grant extra privileges
```

or 

```
Error installing nginx-ingress chart: WARNING: Namespace doesn't match with previous. Release will be deployed to default
Error: UPGRADE FAILED: no ServiceAccount with the name "ingress-nginx-ingress" found
```

###### Note regarding Helm:

A good option is to check your helm version and upgrade if it's not up-to-date with the latest.
Make sure to delete the failed nginx-ingress chart before you try to fix the issue.

```
helm delete --purge ingress

helm upgrade --service-account helm
```

###### Note regarding Minikube:

Minikube is possible to fail sometimes on macOS across sleeps and reboots.

Make sure you've started it properly with `--bootstrapper=kubeadm`.
Otherwise the service-accounts/roles will not be created and tiller will fail on the nginx-ingress chart.

###### Solution:

```
minikube stop
minikube delete && minikube start --vm-driver=hyperkit --bootstrapper=kubeadm --disk-size=50g --memory=6144 --kubernetes-version=v1.8.1
```


##### Issue:

```
Error installing nginx-ingress chart: Error: UPGRADE FAILED: Get https://10.96.0.1:443/api/v1/namespaces/kube-system/configmaps?labelSelector=NAME%!D(MISSING)ingress%!C(MISSING)OWNER%!D(MISSING)TILLER%!C(MISSING)STATUS%!D(MISSING)DEPLOYED: dial tcp 10.96.0.1:443: i/o timeout
```

Most probably you have forgotten to execute `minikube delete` before start.

###### Solution: 

```
minikube stop
minikube delete && minikube start --vm-driver=hyperkit --bootstrapper=kubeadm --disk-size=50g --memory=6144 --kubernetes-version=v1.8.1
```

##### Issue:

```
received unexpected error: Post https://<dispatch-host>:32015/v1/image/base: x509: cannot validate certificate for <dispatch-host> because it doesn't contain any IP SANs

```

###### Solution:

Check whether you have set `"insecure": true` in the `~/.dispatch/config.json` file.