---
layout: default
---

# Setting up Dispatch on GKE

Google Container Engine (GKE) provides fully managed hosted kubernetes clusters.  This is great environment for running
Dispatch, especially if you'd like Dispatch exposed to the internet.

This guide walks through an installation of Dispatch on GKE with certificates and authorization enabled.  It's a good
idea to enable authorization as otherwise the Dispatch API is unprotected and on the internet.  Addtionally, enabling
certificates via letsencrypt enables use-cases which require signed certificates (like slack integrations).

The certificate support is pretty experiemental at this point and requires AWS Route53.  Support for other DNS services
is coming.

## Create GKE cluster

```
export CLUSTER_NAME=dispatch-demo
gcloud container clusters create -m n1-standard-2 --cluster-version 1.9.6-gke.1 ${CLUSTER_NAME}
```

## Fetch cluster credentials

```
gcloud container clusters get-credentials ${CLUSTER_NAME}
```

Kubectl should now be configured to the new cluster:

```
$ kubectl config current-context
gke_dispatch-193801_us-west1-c_dispatch-demo
```

## Setup Helm and Tiller

```
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```

The following commands it to work around a known [helm](https://github.com/kubernetes/helm/issues/3409)
[issues](https://github.com/kubernetes/helm/issues/3379).  If it fails, retry:

```
helm init --wait
```

## Create Secret for DNS Provider

Dispatch uses letsencrypt and relies on the DNS challenge type for issuing certificates.  This requirement
will hopefully go away in the near future, but is necessary for now.  Also, additional DNS services may be
supported.

### CloudDNS

In order to use CloudDNS, you must first obtain a service account key.  The assocated service account must have
DNS admin rights.

```
kubectl create secret generic clouddns --namespace kube-system --from-literal service-account.json=$GCP_SERVICE_ACCOUNT_KEY
```

### Route53

If you are enabling certificates (and you are using AWS Route53), create the following secret:

```
kubectl create secret generic route53 --namespace kube-system --from-literal secret-access-key=$AWS_SECRET_ACCESS_KEY
```

## Install Dispatch without DNS Names

For local development, you may want to install dispatch without worrying about secrets and credentials. The following installation config does so.
```yaml
apiGateway:
  host: 10.0.0.1
  serviceType: LoadBalancer
dispatch:
  host: 10.0.0.1
  port: 443
  debug: true
  skipAuth: true
  eventTransport: kafka
  faas: riff
```

The host values here will automatically get configured during the dispatch install. To see the correct values, look at the configuration file produced at the end of the installation process.

## Install Dispatch With Secrets

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
    provider: clouddns
    clouddns:
      project: example-com
  #   provider: route53
  #   # The AWS_ACCESS_KEY_ID associated with the route53 secret set above
  #   route53:
  #     accessKeyID: ***********
dispatch:
  # A hostname for the dispatch API, set this in Route53 when prompted
  host: host.example.com
  port: 443
  tls:
    ca: letsEncrypt
  faas: riff
  eventTransport: kafka
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
$ dispatch install -f gke-install.yaml --single-namespace=dispatch
Installing cert-manager helm chart
Successfully installed cert-manager chart - cert-manager
Installing charts/certificate helm chart
Successfully installed charts/certificate chart - dispatch-certificate
Installing nginx-ingress helm chart
Successfully installed nginx-ingress chart - ingress
Installing postgresql helm chart
Successfully installed postgresql chart - postgres
Installing charts/kong helm chart
Successfully installed charts/kong chart - api-gateway
Installing kafka helm chart
Successfully installed kafka chart - transport
Installing riff helm chart
Successfully installed riff chart - riff
Installing docker-registry helm chart
Successfully installed docker-registry chart - docker-registry
Installing charts/dispatch helm chart
Successfully installed charts/dispatch chart - dispatch
##############################
Add the following DNS records:
	35.255.3.124		api-host.example.com
	104.2.233.161		host.example.com
##############################
Config file written to: /Users/bjung/.dispatch/config.json
```

## Add the DNS Records

If the install is successful, you should be presented with instructions for updating 2 DNS records.  Add the records
using your DNS provider.

## Authenticate and Setup User Policies

**Ensure that the configured [GitHub OAuth App](https://github.com/settings/developers)
points to the dispatch hostname (i.e. `host.example.com`)**

At this point dispatch is installed, but is in "bootstrap mode".

The following instructions are abbreviated from the
[setup authorization guide](https://vmware.github.io/dispatch/documentation/guides/setup-authentication).
See the guide for more detail and troubleshooting.

1. Log in to dispatch. If everything is setup correctly you should be redirected to github for authentication:
```
dispatch login
```

2. Use the email address associated with your GitHub account to create an iam policy:
```
dispatch iam create policy default-admin-policy --subject github-user@example.com --action "*" --resource "*"
```

3. Validate the policy:
```
$ dispatch iam get policy default-admin-policy --wide
          NAME         |         CREATED DATE         |             RULES
----------------------------------------------------------------------------------
  default-admin-policy | Sat Jan  1 10:17:16 PST 0000 | {
                       |                              |   "actions": [
                       |                              |     "*"
                       |                              |   ],
                       |                              |   "resources": [
                       |                              |     "*"
                       |                              |   ],
                       |                              |   "subjects": [
                       |                              |     "github-user@example.com"
                       |                              |   ]
                       |                              | }
```

4. Disable bootstrap mode (you will now have full access to the API - according to the configured policy):

```
dispatch manage --disable-bootstrap-mode
```

> **Important**: You will need to wait up to 30 seconds for change to be applied
