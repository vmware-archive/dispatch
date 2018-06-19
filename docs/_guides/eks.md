---
layout: default
---

# Setting up Dispatch on EKS (via eksctl)

Amazon Container Service (EKS) provides fully managed hosted kubernetes clusters.  This is great environment for running
Dispatch, especially if you'd like Dispatch exposed to the internet.

This guide walks through an installation of Dispatch on EKS with certificates and authorization enabled.  It's a good
idea to enable authorization as otherwise the Dispatch API is unprotected and on the internet.  Addtionally, enabling
certificates via letsencrypt enables use-cases which require signed certificates (like slack integrations).

The certificate support is pretty experiemental at this point and requires AWS Route53.  Support for other DNS services
is coming.

## Prerequisites

### Update kubectl if necessary

EKS requires a `kubectl` version >= 1.10.0.  Check your version and update if necessary:

```
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"10", GitVersion:"v1.10.1", GitCommit:"d4ab47518836c750f9949b9e0d387f20fb92260b", GitTreeState:"clean", BuildDate:"2018-04-13T22:27:55Z", GoVersion:"go1.9.5", Compiler:"gc", Platform:"darwin/amd64"}

```

### Install eksctl

In order to make the installation and management of EKS simpler, this guide suggests using the
[eksctl tool from WeaveWorks](https://github.com/weaveworks/eksctl):

```
$ curl --silent --location "https://github.com/weaveworks/eksctl/releases/download/latest_release/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
$ sudo mv /tmp/eksctl /usr/local/bin
```

The location of the `eksctl` binary should be somewhere in your path.

### Install heptio-authenticator-aws

EKS relies on heptio authenticator for cluster credentials.  Download and install:

```
$ curl -o heptio-authenticator-aws https://amazon-eks.s3-us-west-2.amazonaws.com/1.10.3/2018-06-05/bin/darwin/amd64/heptio-authenticator-aws
$ chmod +x heptio-authenticator-aws
$ sudo mv heptio-authenticator-aws /usr/local/bin/.
```

## Create EKS cluster

Make sure that you AWS credentials are configured, then simply create the cluster with `eksctl`:

```
export CLUSTER_NAME=dispatch-demo
$ eksctl create cluster --cluster-name $CLUSTER_NAME
2018-06-18T09:23:10-07:00 [ℹ]  importing SSH public key "/Users/bjung/.ssh/id_rsa.pub" as "EKS-dispatch-demo"
2018-06-18T09:23:11-07:00 [ℹ]  creating EKS cluster "dispatch-demo" in "us-west-2" region
2018-06-18T09:23:11-07:00 [ℹ]  creating VPC stack "EKS-dispatch-demo-VPC"
2018-06-18T09:23:11-07:00 [ℹ]  creating ServiceRole stack "EKS-dispatch-demo-ServiceRole"
2018-06-18T09:23:51-07:00 [✔]  created ServiceRole stack "EKS-dispatch-demo-ServiceRole"
2018-06-18T09:24:51-07:00 [✔]  created VPC stack "EKS-dispatch-demo-VPC"
2018-06-18T09:24:51-07:00 [ℹ]  creating control plane "dispatch-demo"
2018-06-18T09:35:14-07:00 [✔]  created control plane "dispatch-demo"
2018-06-18T09:35:14-07:00 [ℹ]  creating DefaultNodeGroup stack "EKS-dispatch-demo-DefaultNodeGroup"
2018-06-18T09:39:55-07:00 [✔]  created DefaultNodeGroup stack "EKS-dispatch-demo-DefaultNodeGroup"
2018-06-18T09:39:55-07:00 [✔]  all EKS cluster "dispatch-demo" resources has been created
2018-06-18T09:39:55-07:00 [ℹ]  wrote "/Users/bjung/.kube/eksctl/clusters/dispatch-demo"
2018-06-18T09:39:59-07:00 [ℹ]  the cluster has 0 nodes
2018-06-18T09:39:59-07:00 [ℹ]  waiting for at least 2 nodes to become ready
2018-06-18T09:40:34-07:00 [ℹ]  the cluster has 2 nodes
2018-06-18T09:40:34-07:00 [ℹ]  node "ip-192-168-100-121.us-west-2.compute.internal" is ready
2018-06-18T09:40:34-07:00 [ℹ]  node "ip-192-168-244-7.us-west-2.compute.internal" is ready
2018-06-18T09:40:34-07:00 [ℹ]  all command should work, try ' --kubeconfig /Users/bjung/.kube/eksctl/clusters/dispatch-demo get nodes'
2018-06-18T09:40:34-07:00 [ℹ]  EKS cluster "dispatch-demo" in "us-west-2" region is ready
```

## Insecure registries

Unfortunately, the default node AMI for EKS nodes does not include the configuration to support insecure registries.
This is necessary to support the internal docker registry.  There are two options (easy and hard):

1. The easy way - simply use an external registry (see the configuration below)

2. The hard way - add the required `/etc/docker/daemon.json` file to the EKS nodes:

```json
{
    "insecure-registries": ["10.0.0.0/8"]
}
```

This can be done a number of ways, all of which are outside the scope of this guide.

## Fetch cluster credentials

```
$ eksctl utils write-kubeconfig --cluster-name $CLUSTER_NAME
2018-06-18T09:53:20-07:00 [ℹ]  wrote kubeconfig file "/Users/bjung/.kube/eksctl/clusters/dispatch-demo"
```

Notice that the config is writen to a non-standard location.  You can specify the location if you like with the
`--kubeconfig` option with the above command or just set the `KUBECONFIG` env var as we are doing below:

```
$ export KUBECONFIG=/Users/bjung/.kube/eksctl/clusters/dispatch-demo
```

You should now be all set to use kubectl and install Dispatch.

## Setup Helm and Tiller

By default tiller will not have necessary permissions for installing Dispatch.

```
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```

```
helm init --wait
```

## Create Secret for DNS Provider

Dispatch uses letsencrypt and relies on the DNS challenge type for issuing certificates.  This requirement
will hopefully go away in the near future, but is necessary for now.  Also, additional DNS services may be
supported.

### Route53

If you are enabling certificates (and you are using AWS Route53), create the following secret:

```
kubectl create secret generic route53 --namespace kube-system --from-literal secret-access-key=$AWS_SECRET_ACCESS_KEY
```

### CloudDNS

In order to use CloudDNS, you must first obtain a service account key.  The assocated service account must have
DNS admin rights.

```
kubectl create secret generic clouddns --namespace kube-system --from-literal service-account.json=$GCP_SERVICE_ACCOUNT_KEY
```

## Install Dispatch

Create a installation config as follows, subsituting values where appropriate:

```yaml
ingress:
  serviceType: LoadBalancer
apiGateway:
  serviceType: LoadBalancer
  # A hostname for the API gateway, set this in Route53 when prompted
  host: api-host.example.com
letsEncrypt:
  # An email address where certificate expiration warnings are sent
  email: user@example.com
  # Use letsencrypt prod
  staging: false
  # Default DNS provider is clouddns
  dns:
    provider: route53
    # The AWS_ACCESS_KEY_ID associated with the route53 secret set above
    route53:
      accessKeyID: ***********
    # provider: clouddns
    # clouddns:
    #   project: example-com
dispatch:
  # A hostname for the dispatch API, set this in Route53 when prompted
  host: host.example.com
  port: 443
  tls:
    ca: letsEncrypt
  faas: openfaas
  eventTransport: kafka
  imageRegistry:
    name: dockerhub-org
    username: dockerhub-user
    password: **********
  # Bootstrap user should be the email address of the github user assocated
  # with the github app (below).  See the [authorization docs]
  # (https://vmware.github.io/dispatch/documentation/guides/setup-authentication)
  bootstrapUser: user@example.com
  oauth2Proxy:
    provider: github
    clientID: **********
    clientSecret: ***********
```

Run the dispatch install command:

```
$ dispatch install -f eks-install.yaml --single-namespace=dispatch
Installing cert-manager helm chart
Successfully installed cert-manager chart - cert-manager
Installing certificate helm chart
Successfully installed certificate chart - dispatch-certificate
Installing nginx-ingress helm chart
Successfully installed nginx-ingress chart - ingress
Installing postgresql helm chart
Successfully installed postgresql chart - postgres
Installing kong helm chart
Successfully installed kong chart - api-gateway
Installing openfaas helm chart
Successfully installed openfaas chart - openfaas
Installing kafka helm chart
Successfully installed kafka chart - transport
Installing docker-registry helm chart
Successfully installed docker-registry chart - docker-registry
Installing dispatch helm chart
Successfully installed dispatch chart - dispatch
##############################
Add the following DNS records:
	***-1356830742.us-west-2.elb.amazonaws.com		api-eks-01.example.com
	***-562108187.us-west-2.elb.amazonaws.com		eks-01.example.com
##############################
Config file written to: /Users/****/.dispatch/config.json
```

## Add the DNS Records

If the install is successful, you should be presented with instructions for updating 2 DNS records.  Add the records
using your DNS provider.

## Authenticate and Setup User Policies

**Ensure that the configured [GitHub OAuth App](https://github.com/settings/developers)
points to the dispatch hostname (i.e. `host.example.com`)**

At this point dispatch is installed, but is lacking users and polices.  We need to "bootstrap"

The following instructions are abbreviated from the
[setup authorization guide](https://vmware.github.io/dispatch/documentation/guides/setup-authentication).
See the guide for more detail and troubleshooting.

1. Create a rolebinding for the anonymous user (this should go away soon):

```
$ kubectl create rolebinding anonymous-dispatch --clusterrole=admin --user=system:anonymous --namespace=dispatch
rolebinding.rbac.authorization.k8s.io "anonymous-dispatch" created
```

2. Use the email address associated with your GitHub account as the boostrap user (you can leave out the bootstrap
user to just use an auto-generated service account):

```
$ dispatch manage bootstrap --kubeconfig $KUBECONFIG --bootstrap-user github-user@example.com
enabling bootstrap mode
waiting for bootstrap status............success
Creating Organization: default
Creating Policy: default-policy
disabling bootstrap mode (takes up to 30s to take effect)
waiting for bootstrap status..................success
```

3. Login with the bootstrap user to your IDP (github):

```
$ dispatch login
You have successfully logged in, cookie saved to /Users/****/.dispatch/config.json
```

## Use Dispatch

```
$ dispatch create seed-images
Created BaseImage: nodejs-base
Created BaseImage: python3-base
Created BaseImage: powershell-base
Created BaseImage: java-base
Created Image: nodejs
Created Image: python3
Created Image: powershell
Created Image: java
```
