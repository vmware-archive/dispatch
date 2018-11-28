---
# You don't need to edit this file, it's empty on purpose.
# Edit theme's home layout instead if you wanna make some changes
# See: https://jekyllrb.com/docs/themes/#overriding-theme-defaults
layout: default
---

# Dispatch - more than Functions

Building, deploying and administering serverless applications requires more than just a function scheduler and executor.
Dispatch brings features and services together to easily manage production-caliber applications and services which are
built upon functions.

## Features

### A secure FaaS

Giving developers direct access to VMs and containers can be problematic.  IT can quickly lose control over the
production environment, with no knowledge of what software is actually running in their data center.  On the other hand,
if developers have no direct access to VMs and containers this causes bottlenecks in IT and developer frustration.

Function based deployments can alleviate much of this concern.  Dispatch provides an environment where everything up to
the actual function code can be managed and inspected.  Pivotal Container Service (PKS) from VMware provides a secure
and up-to-date Kubernetes service. Dispatch manages function runtimes and artifacts (containers) built on the latest
Photon OS.

### A multi-tenant FaaS

Dispatch is designed to work in a multi-user and multi-organization environment.  Dispatch will integrate with existing
Oauth2 compatible identity providers such as github or Active Directory.  Administrators will be able to create roles
and permissions to ensure tight access control around the full Dispatch API.

### An integrated API gateway

Whether building a simple web-hook or a richer web-service in order to trigger the functions that make up a deployment,
an API gateway is required.  The API gateway provides routing and security.  Dispatch integrates an
API gateway to provide a solution for applications built on Dispatch.  Simply define a route and bind
it to a function to create a secure HTTPS endpoint.

### Plays well with others

Integration with external services and events is critical for any serverless solution.  Dispatch includes an external
services interface that allows extending Dispatch to work with just about any other service.  These services could be
databases to provide state to applications, or event sources which provide triggers to functions.  The interface is flexible and extensible.

## Versions

The Dispatch project consists of two versions or branches:

* [Dispatch-Solo](#dispatch-solo) (solo branch) - Minimal dependencies and convenient packaging for evaluation
* [Dispatch-Knative](#dispatch-knative) (master branch) - Production ready, built upon Kubernetes and Knative

## Dispatch-Solo

Dispatch-Solo is a branch of Dispatch which is intended to offer the full functionality and user experience of Dispatch
while requiring very few dependencies.  Additionally, Dispatch-Solo is packaged as a VM appliance making it as easy as
possible to get a functioning Dispatch environment in seconds.

There are of course limitations:
* Scale - Dispatch-Solo is a single binary and not designed to scale beyond that.
* IAM/Tenancy - Dispatch-Solo is intended to be single-user.
* Services - Dispatch relies on the Kubernetes service catalog to bring services integration.  Dispatch-Solo does not include Kubernetes and therefore the service catalog.
* Persistence - Dispatch-Solo relies on a simple embeded database (BoltDB)

## Dispatch-Knative

Dispatch-Knative is the long-term production version of Dispatch. This version of Dispatch is dependent on Kubernetes
and [Knative](https://knative.dev).  It is currently under heavy development.

## Join the Dispatch team

You can find Dispatch on Github. Many features are still a work in progress, but we encourage the curious to start
building and imagining with us.
