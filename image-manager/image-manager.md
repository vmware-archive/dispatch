# Serverless Image Manager

A service integrated with VMware serverless which enables extensible and custom base images enabling a wider range
of function based workloads, smaller function payloads, and better tracking of function dependencies.

## Current Issues

In order to run a function on a FaaS implementation, a function needs a runtime.  Generally, these runtimes are dynamic
languages like python, javascript/node.js, ruby, etc.  The runtime also includes a set of pre-installed packages
associated with the given language.  These pre-packaged runtimes are often docker containers, though there may be
FaaS implementations which use a different packaing primitive.  With regards to VMware serverless, we are only
interested in FaaS implementations which are container based.

Most serverless services only offer a select number of runtimes with a specific list of pre-installed tools and
packages.  For instance, on AWS Lambda, the python3 runtime which includes the boto (aws) toolkit, so that interacting
with the AWS services is easy.  If you need additional packages, then you will need to zip up all the dependencies
and submit the entire package as your function.  Because, most problems require 3rd party libraries, most functions
end up being zipped up packages.  This works fine, but there are limitations:

* The maximum package size is fixed, though serverless service dependent.
* Extentions must not depend on system libraries (python postgres client for example).
* The larger the "function" becomes, the harder it is to debug and instrument.
* Often related functions will require a similar set of "base" dependecies, and there is no way for the infrastructure
  to take care of packaging up these dependencies in a common, consistent fashion.
* Developers don't know what packages/libraries are pre-installed, leading to a lot of trial and error.

## Proposed Solution

An image manager service which integrates with a FaaS implementation could address a lot of the current issues with
a fixed set of runtime images, without giving up the control and security that fixed runtimes enable.  In fact,
implemented correctly, security and visibility could be increased with an integrated image manager.  Allowing
administrators and developers to easily manage and customize the function runtime could offer the following benefits:

* Actual functions become a single code artifact, with all the dependences integrated into the runtime image.
* Function dependencies may be managed and tracked.
    * Ensure compatible licenses for all dependencies.
    * Know which base images/functions are susceptible to particular security vulnerabilities.
    * Require libraries which enable seemless integration into customer infrastructure monitoring (logs, events,
      metrics).
* Provide a vetted/whitelisted selection of approved dependencies which developers may create base images from.

## Requirements

* A FaaS which is docker/container based and supports running custom images.
    * Support for private registries (private dockerhub repositories and/or Harbor).
* Image service is a REST based service enabling easy client integrations.
    * Tools (UI/CLI) which support managing available and installed packages.
    * Support for Authentication and Authorization
        * Only admin type roles may create/modify package and image inventory.
* All images must be PhotonOS based

## User Stories

* As a developer, I want the flexibility to use what I feel is the right tool for the job.
* As a cloud administrator, I want to track the libraries and dependencies deployed in my infrastructure so that I'm
  aware of the security risks at all times.
* As a developer, I want to focus on just the function definition, not packaging or security

## Design

![image manager](image-manager.png "Image Manager Design")

### Root Images

Although not really a component, root images must also be considered as part of the image manager.  The root images are
FaaS dependent as they include the mechanisms for interacting directly with the FaaS, and they define the function
signatures to some degree.  The root images could just be the runtime images which are packaged as part of the FaaS
implementation (i.e. [openwhisk](https://hub.docker.com/u/openwhisk/)).  However, as stated previously as one of the
requirements, all images should be PhotonOS based.  Therefore, at a minimum, the FaaS specific images will need to be
repackaged.  A set of base images could be available publicly on dockerhub.  The administrators of the image manager
may also specify specific base images.  **All images must derive from a base image directly or indirectly.**

Base images can also allow us to insert hooks into other features such function schema validation, logging and metrics.

### Packages

The goal of the image manager is to add both flexibility and control into the software installed on the runtime images.
Packages have been split into two categories, system packages and runtime packages.

#### System Packages (PhotonOS)

System packages are Linux (PhotonOS) packages.  While most function will only require dependencies associated with the
runtime (i.e. python packages), some runtime packages depend on system packages.  There are also use-cases with regard
to logging and telemetry.  By allowing developers to select "blessed" system packages to build their runtime images,
we open up the serverless use-cases and ease of development.

Initially, the list of available system packages will be based on the [PhotonOS package repository](https://dl.bintray.com/vmware/photon_dev_x86_64/packages/).  Administrators may choose which packages may be available
for developers to use.  This give developers flexibility to customize the image, and administrators the control to
know what packages are installed on what image.

#### Runtime Packages

Runtime packages are language/runtime dependent.  They are the node, python, golang, etc. packages which functions
directly depend on.  Generally when a FaaS function requires a runtime dependency, the function author must resort
to zipping up the dependency code along with the function code.  While this works, it's not ideal:

* Harder to inspect the actual function code, as the function artifact is now a binary containing multiple files.
    * Harder to use a UI to edit function code.
* The size of the artifact is drastically larger (depends on the size of the dependencies).
* All functions which have similar dependencies, must go throught the same process of bundling the dependencies.

If the runtime packages can be included in the runtime image, functions can be simply functions: small, easy to manage
bites of code.

### Image Definition

Docker files are a really great way to easily define container images.  However we require a more structured format
for parsing packages and dependencies.  In order to truely manage the images from base operating system to system
libraries to runtime, application and infrastructure dependent packages, a well defined schema is required.  A
simplistic example could be:

```
python3-postgres.yaml

name: python3-postgres
version: 1.1
parent: photon-python3
language: python3
metadata:
    role: x-labs-admins
    extendable: true
    author: bjung@vmware.com
photon-packages:
    - name: libpq-dev
      version: 9.6.4-1
    - name: python2-dev
      version: 3.5.3-3
runtime-packages:
    - name: psycopg2
      version: 2.7.3.1
```

From this image definition, it would be easy to validate all of the photon and runtime packages against an approved list
of dependencies.  Additionally, creating a Dockerfile and resulting image is relatively trivial.  This is of course a
very simple image.  A full image schema will likely require more features and flexibility:

* managing packages not present in default repositories
* managing packages installed via other means (i.e. `curl -OL ... | sh`)
* running commands required for package setup/initialization (i.e. `chmod +x ...`)

The image definition itself should be defined in jsonschema/swagger.  This provides easy documentation and API
validation.

### REST Server

There are two integration points with external services, an HTTP based REST server and the image registry itself.  The
REST server exposes all of the management utility of the image manager.  The actual REST server should be very
lightweight.  At a minimum, the server should expose the following functionality:

* Full swagger schema definition
    * Generate documentation
    * Validate input/output
* Support authentication and authorization
* CRUD (Create/Read/Update/Delete) "root" images
    * Pure docker images represent the root of the image heirarchy
* CRUD system and runtime packages
* CRUD runtime images

#### Authorization

The server should enable administrators to associate roles to particular actions.  Obvious permissions may be:

* Administrator (Dev/Cloud ops)
    * full access
* Priviledged Developer (Dev leads)
    * read-only access to root images
    * read-only access to system packages
    * read-only access to runtime packages
    * read/write access to runtime images
* Function Developer
    * read-only access

#### Persistence

All objects configuration/state (root images, system and runtime packages, and runtime images) need to be persisted in a
key/value store. A key/value store offers a very simple interface and can be backed my many different database
implementations.  The image manager should take advantage of existing database infrastructure and not require additional
stateful services to be managed.  For instance, a service manager deployed on kubernetes may take use the same etcd
database as kubernetes though in a different namespace.  This reduces the operational overhead of the entire system.

### Runtime Plugins

The purpose of the runtime plugins is to generate runtime/language specific Dockerfiles.  It's likely that the plugins
may consist of little more than template files.  The [image definition schema](#image-definition) will not be
runtime/language specific.

### Image Builder

The image builder is another small component whose primary responsibility is to take the generated Dockerfiles and build
and deploy the created images.  Additionally, as the builder is the image manager's interface to docker, all image
delete and update operations should go through the image builder.

### Image Repository

The managed container images are stored and accessed in a docker image repository.  The image manager could support
both public repositories (Dockerhub) as well as local repositories (Harbor).  Much like the database, we would like
to leverage (or at least support) existing installations.  It may be necessary to include a Harbor installation with
the deployment of the image service, though that support may be a lower priority.

Because the FaaS interacts directly with the repository, the FaaS itself may need patching (Openwhisk) to support
private repositories (i.e. image pull secrets).

## Milestones

TBD

## Open Issues

* Deleting/updating dependencies may cause dependent images to become invalid.
    * Updates could trigger batch jobs to update dependent images (Image
      Builder).
    * Deleting dependencies could result in a couple options:
        * Mark any dependent images as "deprecated" or the like.
        * Delete all dependent images.
* What if the backing database does not support transactions?
    * Would affect any operations which writes to multiple objects.
* Does this model change in a multi-tenant environment?
    * Should there be "public" images?
* How do we track which images are currently in use by functions?
    * This is probably a task/requirement for the event gateway, as the image
      manager has no knowledge of functions.