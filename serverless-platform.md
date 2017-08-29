# VMware Serverless Platform

The VMware serverless platform is a set of services which wrap and enhance an existing open source FaaS implementation.
The enhancements include security, control, and flexibility solutions to provide an enterprise ready experience.

## Architecture

![serverless platform](serverless-platform.png "VMware serverless platform")

As can be seen from the  above diagram, the VMware serverless platform is comprised of several services:

* Event Gateway
* FaaS Implementation (Openwhisk)
* [Image Manager](image-manager.md)
* Identity Service
* REST API
* API Gateway
* CLI
* UI

All services are deployed as containers on Kubernetes (PKS).  There are some additional components called out in the
diagram which the serverless platform depends on, but may be managed separately:

* Config DB (Etcd)
    * Possibly use the etcd which kubernetes uses.
* Docker Registry (Harbor)
    * Harbor also may be part of the kubernetes deployment, or integrating with Dockerhub should be easy as they are
      API compatible.
* Identity Provider (vIDM)
    * One of the goals is to integrate with existing customer directories and services.
* Secret Store
    * Likely use kubernetes secrets initially.

## Milestones

TBD

## Team

Karol Stepniewski <kstepniewski@vmware.com> - Event Gateway/REST API/CLI

Berndt Jung <bjung@vmware.com> - Image Manager

Xueyang Hu <xueyangh@vmware.com> - Identity Service

Nick Tenczar <ntenczar@vmware.com> - Secret Injection

Ivan Mikushin <imikushin@vmware.com> - Function Schema and Validation