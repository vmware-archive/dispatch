---
layout: default
---
# Service Manager

The purpose of the service manager is to provide a bridge between external services and functions. Two new concepts
will be introduced into Dispatch: catalogs which advertise available services, and services which represent a service
instance and a binding. The binding is transmitted to functions at execution time.

## Problem statement

Functions are small stateless code artifacts.  Alone, they are not very useful. They need services like databases,
messangers, queues, etc. to address a wide variety of use-cases.

### User stories (examples)

1. As a **developer**, I need to my function to connect to database.
2. As an **administrator**, I want to manage the services and credentials that developer and their functions have
   access to.
3. As a **developer/administrator** I want a unified and consistent way of exposing services to developers and
   functions.
4. As an **administrator** I need an easy way to add arbitrary services, whether they are common and included with
   Dispatch, or custom internal services.
5. As a **developer**, I want a catalog of services that I can use with minimal effort/configuration.

## Proposed Solution

A Service Manager component, responsible for catalog and service resources.  The service manager will provide a
Dispatch API layer on top of the [Kubernetes service catalog](https://github.com/kubernetes-incubator/service-catalog),
meaning that services must be Open Service Broker compatible.

While we could use the Kubernetes service catalog directly, there is no notion of tenancy so all services would be
globally available.  For Dispatch, we want the catalog to be organizationally scoped, meaning an administrator could
manage the catalog on a per-organization (perhaps per-application) basis.  Additionally, brokers should be dynamically
provisionable through the Dispatch API, without needing access to Kubernetes directly.  Therefore a light-weight
Dispatch service-manager API is required.

### Brokers

Brokers are [Open Service Broker API](https://www.openservicebrokerapi.org) (OSBAPI) compatible service brokers.  The
contain a set of service classes which are responsibile for provisioning and binding to the services the service
classes represent.  The brokers are identifiable via a URL and implement the OSBAPI.

Dispatch administrators manage the available brokers.  The brokers will be organization scoped.  Upon registering a
broker a list of service classes will be returned.

The following CRUD operations will be supported:

* CREATE - register a broker via URL
    - optionally specify service classes to register
* GET - return a broker object which lists all available service classes
* LIST - return all brokers
* DELETE - delete a broker and associated service classes

### Service Classes

Service classes know how to provision and bind to particular services.  Once a broker is registered, a list of service
classes is made available.  Administrators may manage the list of services which are made available to developers.

The following CRUD operations will be supported:

* CREATE - register a service class from a specific broker and make it available for the organization
* GET - return a service class object including available plans (and possibly schema)
* LIST - return all service classes
* DELETE - remove a service class making it unavailable to the organization

### Service Instance

A service instance represents a provisioning or instance of a service.  For instance the instantiation of a postgres
service class would result in a provisioned database and a binding created (credentials stored as secrets).  The service
is identifieable with a unique name (e.g. `my-first-postgres`).  Functions can then be registered with services by name
(same was as secrets) and have the binding data injected into the function context (again, just like secrets).  In fact,
in this scenario the credentials are stored as secrets and the service instance maintains the secret name, which is how
the credentials are looked up.

The following CRUD operations will be supported:

* CREATE - create a service instance from a service class
    - schema will be service class specific
* GET - return the service which includes the binding information (e.g. Dispatch secret keys)
* LIST - return a list of all service instances
* DELETE - delete the service and do any service cleanup

## Design

As stated earlier, the service-manager is a light API layer on top of the kubernetes service catalog.  Additionally,
we are adding per-organization management of available services.

### Broker object schema

* `name` - `string` - broker name
* `URL` - `string` - URL to OSBAPI compatible broker
* `parameters` - `map` - key/value paramteters passed as args to the broker
* `serviceClasses` - `map` - key/serviceClass mapping (not sure if map or list)

> NOTE: Broker management will not be available immediately.

### Service Class object schema

* `name` - `string` - service class name
* `brokerID` - `string` - ID of broker
* `plans` - `map` - key/servicePlan mapping
    - service plan objects may contain schema with regards to service creation and binding

### Service Instance object schema

* `name` - `string` - service instance name
* `serviceClassID` - `string` - ID of service class
* `planID` - `string` - ID of service plan
* `parameters` - `map` - parameters used for service provisioning and binding
* `binding` - `map` - binding data (secret names) used by functions to use to this service instance

## Milestones

### Phase 1

* Brokers
    - no Dispatch broker objects/APIs
    - registered externally with kubernetes/service-catalog
* Service Classes
    - pulled directly from service-catalog (auto-registered)
    - read-only API (only GET/LIST operations)
* Service Instance
    - CREATE/GET/LIST/DELETE all supported
    - pass `--service <serviceName>` to function to inject binding data
    - Binding data unpacked and repacked into Dispatch secret (kind of ugly)

