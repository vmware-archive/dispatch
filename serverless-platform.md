# VMware Serverless Platform

The VMware serverless platform is a set of services which wrap and enhance an existing open source FaaS implementation.
The enhancements include security, control, and flexibility solutions to provide an enterprise ready experience.

## Architecture

![serverless platform](serverless-platform.png "VMware serverless platform")

As can be seen from the  above diagram, the VMware serverless platform is comprised of several services:

* [REST API](rest-api/rest-api.md)
* [Event gateway](event-gateway/event-gateway.md)
* [Function gateway](function-gateway/function-gateway.md)
    * [Function schema validation](function-gateway/function-schema-validation.md)
* FaaS Implementation (Openwhisk)
* [Image Manager](image-manager/image-manager.md)
* [Identity Service](identity-management/identiy-management.md)
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

### Persistence

All services utilize a single central data store for persisence.  Access to the datastore will be managed via an
interface which should provide adequate abstraction that the actual database is not exposed.  This ensures that the
services/application are not too tightly coupled to any particular database, and provides a means for testing servies
without external dependencies.  For instance, a simple map based mememory store could be used for unittesting, so long
as the interface is maintained.  Additionally, this is a distributed system and both lock and watch semantics are
important.  Locks provide a way of synchronizing work across multiple components, and watches can reduce the dependency
on polling for changes.

It may also make sense for the interface to manage and enforce schemas for the objects being persisted.  At a minimum
an envelope with common fields as well as a key scheme which segments the data by organization.

In the absense of transactions, a revision field will be used to ensure consistency.  When updating a record, the
previous revision is sent with the update.  If the revision on the server/database differs from the previous revison
the record has been updated by another party.  Therefore the update should fail.  This requires some atomicity support
by the database.

## Terminology

See the following [full glossary of terms](entities.md)

## Milestones

TBD

## Team

Karol Stepniewski <kstepniewski@vmware.com> - Event Gateway/REST API/CLI

Berndt Jung <bjung@vmware.com> - Image Manager

Xueyang Hu <xueyangh@vmware.com> - Identity Service

Nick Tenczar <ntenczar@vmware.com> - Secret Injection

Ivan Mikushin <imikushin@vmware.com> - Function Schema and Validation
