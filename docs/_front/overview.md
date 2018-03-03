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

## A flexible FaaS

Dispatch itself is not a function scheduler and executor.  Since there are several open source FaaS implementations, the
Dispatch team decided two things early on.  First, we would not create yet another open source FaaS.  Second, the space
is too young and immature to predict with any certainty that one FaaS would be the dominant player over others.

Instead, Dispatch abstracts the FaaS implementation via a driver interface allowing integration with one or more
existing or future FaaS implementations. The initial Dispatch release includes drivers for OpenFaaS and Riff.  Future
drivers could include support for additional FaaS implementations, including public cloud offerings such as AWS Lambda.
The flexibility to integrate multiple FaaS implementations is more than just future proofing, it also opens the door to
interesting scenarios where the same function can be executed on one or more different environments based on criteria
such as locality, compute resources (GPU) or burst scaling onto the public cloud.

## A secure FaaS

Giving developers direct access to VMs and containers can be problematic.  IT can quickly lose control over the
production environment, with no knowledge of what software is actually running in their data center.  On the other hand,
if developers have no direct access to VMs and containers this causes bottlenecks in IT and developer frustration.

Function based deployments can alleviate much of this concern.  Dispatch provides an environment where everything up to
the actual function code can be managed and inspected.  Pivotal Container Service (PKS) from VMware provides a secure
and up-to-date Kubernetes service. Dispatch manages function runtimes and artifacts (containers) built on the latest
Photon OS.

## A multi-tenant FaaS

Dispatch is designed to work in a multi-user and multi-organization environment.  Dispatch will integrate with existing
Oauth2 compatible identity providers such as github or Active Directory.  Administrators will be able to create roles
and permissions to ensure tight access control around the full Dispatch API.

## A proper API gateway

Whether building a simple web-hook or a richer web-service in order to trigger the functions that make up a deployment,
an API gateway is required.  The API gateway provides routing and security.  Dispatch integrates the open source Kong
API gateway to provide a production quality solution for applications built on Dispatch.  Simply define a route and bind
it to a function to create a secure HTTPS endpoint.

## Plays well with others

Integration with external services and events is critical for any serverless solution.  Dispatch includes an external
services interface that allows extending Dispatch to work with just about any other service.  These services could be
databases to provide state to applications, or event sources which provide triggers to functions.  Included in the
preview release is a vCenter driver which ingests vCenter events which functions can now subscribe to.  The interface is
flexible and extensible.

## Join the Dispatch team

You can find Dispatch on Github. Many features are still a work in progress, but we encourage the curious to start
building and imagining with us.
