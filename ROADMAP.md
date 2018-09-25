# Dispatch Roadmap

There are currently 3 efforts going on within Dispatch.

* Dispatch - the original project, installable on Kubernetes and flexible with regards to FaaS backends.

* Dispatch-Solo - a spinoff of Dispatch which removes all dependencies (including Kubernetes) except for Docker.  This
  is intended as a way to explore Dispatch and FaaS with as little fuss as possible.

* Dispatch-Knative - a replatforming of the original project onto the Knative building blocks.  By fully leveraging the
  Knative components, Dispatch can be thinner and focused on enterprise features and usability.

There will be some consolidation and this is reflected in this roadmap.

## Dispatch (v0.1)

Dispatch v0.1 is deprecated.  Releases will be cut on-demand and based on bug fixes.  While the feature set of
Dispatch v1.0 should be comparable with v0.1, there will be breaking changes to the API and CLI.  There is no future
for Dispatch on Kubernetes without Knative.  It remains largely for demonstration purposes at least until Dispatch-Solo
is feature complete.

- [ ] Setup daily e2e test to ensure current release continues to function as expected

## Dispatch-Solo

Dispatch introduced "local mode" with v0.1.21.  Dispatch could now be run as a single binary, locally, without Kubernetes.
The goals were simple:

1. Easy to develop locally
2. Easy to get started for evaluation
3. Easy to distribute

Dispatch-Solo is a Dispatch which *only* supports local mode.  It will be a leaner Dispatch which removes the Kubernetes
support and focuses on lowering the barrier to entry to actually use Dispatch.

Because Dispatch and serverless is still in its infancy, understanding use-cases is the highest priority.  It therefore
makes sense to continue to focus on getting Dispatch in as many hands as possible and making it as easy as possible to
get started exploring usage. Dispatch-Solo will also focus on packaging.  In addition to releasing mac/linux binaries,
minimal configuration machine images will also be released, making getting started in a non-Kubernetes environment quick
and predictable.

Kubernetes (with Knative) will continue to be the future for production deployments of Dispatch, however there may
remain a need for Dispatch-Solo.  There may be a time in the future where getting a managed Kubernetes cluster is a
simple as provisioning a VM today, but we are far from that time today.  However, once Dispatch-Knative is feature and
API compatible, it's possible that the Dispatch-Solo images may package up Kubernetes + Knative with Dispatch-Knative.
At this point we would be back to a single code base.  This would be the preferrable outcome.

### Dispatch-Solo (v0.1) ~ 11/2018

- [ ] Package as PhotonOS based VM image
    - [ ] OVA (VMware)
    - [ ] UI/Wizard configurable
- [ ] Full functional compatibility with Dispatch v0.1 (excluding service catalog)
    - [ ] Enable event drivers
    - [ ] Enable authentication/authorization
- [ ] Support HTTPS
    - [ ] Generate self-signed certificates
    - [ ] Configure/Import certificates
- [ ] Read-only UI showing status of all Dispatch resources
- [ ] Setup CI/e2e tests

### Dispatch-Solo (v0.2) ~ 1/2019

- [ ] Package as PhotonOS based VM image
    - [ ] AMI (AWS)
- [ ] Migrate to Knative-compatible base images
- [ ] Remove Kubernetes dependencies from code and binary
- [ ] Update API to reflect a Knative-compatible Dispatch
- [ ] Update CI/e2e tests
- [ ] Initial function/driver (bundle) catalog support

### Dispatch-Solo (v0.3) ~ 3/2019

- [ ] Upgrade support between Dispatch-Solo releases
- [ ] API compatible with Dispatch-Knative (v0.2)
- [ ] Common CLI
- [ ] (possible) Common Istio based API Gateway

## Dispatch-Knative

Dispatch-Knative is currently a branch of Dispatch which is a replatforming based on the Knative project.  The Knative
project includes a set of primitives or building blocks for creating serverless applications.  The 3 (currently) main
building blocks included in Knative are:

* Building - Source to image
* Serving - Scheduling and executing containers
* Eventing - Extensible eventing system

Dispatch-Knative leverages all of these building blocks to provide the same feature-set of Dispatch while significantly
reducing the footprint and complexity of Dispatch code.  The following dependencies are replaced with functionality
included within the Kubernetes and the Knative building blocks:

* Postgres - state is captured in Knative and custom Kubernetes CRDs
* Kong - API gateway functionality is provided by Istio
* FaaS (OpenFaaS/riff/Kubeless) - Knative serving handles function scheduling and execution
    - Each FaaS also has a set of dependencies which will no longer be required
* Zookeeper - Kubernetes CRDs/controllers manage concurrent access to resources

Knative eventing will still require a persistent queue, therefore Kafka will likely remain a dependency.

The medium-term goal is functional compatibility with Dispatch v0.1, and API compatiblity between Dispatch-Knative and
Dispatch-Solo.  Knative is a very young and evolving project.  It is expected that changes in Knative will affect
Dispatch as it develops.

### Dispatch-Knative (v0.1) ~ 11/2018

- [ ] Basic Dispatch v0.1 functionality
    - [ ] Image building
        - [ ] All base images supported
    - [ ] Function building and execution
        - [ ] Secret injection
    - [ ] Istio-based API gateway functionality (without certificates)
    - [ ] Secret management
- [ ] Single helm chart installation
- [ ] Setup CI/e2e tests

### Dispatch-Knative (v0.2) ~ 3/2019

- [ ] Full Dispatch functionality (based on Dispatch-Solo)
    - [ ] Eventing support
    - [ ] Authentication/Authorization
- [ ] API compatible with Dispatch-Solo (v0.3)

## Dispatch (v1.0)

Dispatch 1.0 represents several milestones:

* Dispatch v0.1 is EOL
* Single API between Dispatch-Solo and Dispatch-Knative
* Common client/UI/CLI across Dispatch servers (Knative or Solo)

By converging on a single API, any clients (libraries/CLI/UI) can all be made
common.  Additionally, CI tests should run against both environments with
minimal differences.  This makes maintaining two servers considerably easier.

Another goal is to maximize the common code.  The servers should be thin, and
plugin-like, where the interface is now the Dispatch API as opposed to lower
level plugins specific to each "service".  Dispatch is more opinionated now,
there are 2 supported implementations (Solo and Knative) not the matrix of
different FaaS' and Queues.

It's also possible that it makes sense to replace the "local mode" Dispatch
packaged within Dispatch-Solo with Dispatch-Knative at this point.  This will
largely be based on the ability to easily package Kubernetes + Knative into
the Dispatch-Solo image.  However, it's too early to tell when this will be
a viable option.

Further details and dates are TBD and will be dictated by learnings from both
Dispatch-Solo and Dispatch-Knative in addition to the evolution of the Knative
project.
