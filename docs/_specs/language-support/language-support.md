---
layout: default
---
# Programming Language Support in Dispatch
To be able to create functions written in a particular programming language, one must provide a *base image* for that language. 

# Problem

Previously to adding this feature, it was hard to add new languages to Dispatch: 

1. To add a new language, it should have been added to all FaaSs currently supported in a FaaS-specific way. 
2. To add a support for a new FaaS to Dispatch, one not only needed to implement the FaaS driver, but also add the languages in a FaaS-specific way. As a consequence, language support differed between FaaS-drivers.
3. It was impossible to add a language without rebuilding Dispatch.


# Images
## Base Image

A *base image* is (as the name implies) used as the base image for Dispatch functions (and images). As such, it usually contains the language runtime and its standard library, as well as any necessary system packages. 

It is a requirement for the base image to have these metadata labels (using Dockerfile `LABEL` instruction):

- **`io.dispatchframework.imageTemplate`** = `/image-template` (the default value)
- **`io.dispatchframework.functionTemplate`** =`/function-template` (the default value)

These labels point to directories within the _base image_ and _image_ correspondingly. 


## Image Template

This directory contains the Dockerfile and any supporting files (but typically, just the Dockerfile) to build *images* (a.k.a. “deps-images”) for functions written in the language provided by the base image. 

The Dockerfile for the *image* accepts build args (using `ARG` instruction):

- **`BASE_IMAGE`** — the *base image* name, to be used as the base of the image:
    ARG BASE_IMAGE
    FROM ${BASE_IMAGE}
- **`SYSTEM_PACKAGES_FILE`**, defaults to `system-packages.txt` — a text file within the build container, into which the required system package names and versions are written by Dispatch image builder. Image builder expects these system packages to be installed as part of the image build.
  Each package is written to a separate line, name and version are space separated:
    gzip 1.8-1.ph2
    curl 1.19.1-4.ph2
- **`PACKAGES_FILE`**, defaults to `packages.txt` — a dependency manifest file, put into the image build container AS IS by Dispatch image builder. Image builder expects the dependency manifest to be used by the language dependent package manager to install the libraries listed in it. 


## Function Template

This directory contains the Dockerfile and any supporting files (but typically, just the Dockerfile) to build *function images*. A function image is the image for function implementation containers. 

The Dockerfile for the function image accepts build args (using `ARG` instruction):

- **`IMAGE`** — the *image* the be used as the base for the function image:
    ARG IMAGE
    FROM ${IMAGE}
- **`FUNCTION_SRC`**, defaults to `function.txt` — the file with the function source code, put into the function image build container AS IS by Dispatch function builder. 

Function builder expects the Dockerfile to have all the necessary instructions to build and install the function, so that when a container is run from the function image, the *Function Runtime API* (see below) is exposed on the port specified by `PORT` env var, which is 8080 by default. 


# Function Runtime API

It’s a request/reply RPC-style function invocation API implemented with a simple HTTP server written in the language provided by the *base image*. All function invocations go through this API. How the invocations reach this API is FaaS specific and outside of the function runtime responsibilities. 

When a container is normally run from a *function* image (without specifying commands), the main (and usually, the only) process in that container listens on the specified `PORT` and serves this API via HTTP. 

It is the API server process’s responsibility to adequately react to OS signals (such as `SIGTERM`) and perform graceful shutdown, i.e. to stop accepting new requests and finish processing already accepted invocation(s) before exiting. 


## Endpoint: **`/healthz`**

To be used by the infrastructure to check that the container is up and running. 

**Request**
Method: GET

**Response**
Status: 200 OK
Headers:

- `Content-Type: application/json`

Body:
`{}`


## Endpoint: **`/*`** (any URI)

To be used as the function invocation endpoint.

**Request**
Method: POST
Headers:

- Content-Type: application/json
- Accept: application/json

Body:

- **`context`**:
  - **`secrets`** — key/value map of secrets provided to the invocation
  - *(TBD — other contextual data provided to the invocation by Dispatch)*
- **`payload`** — input object/value, can be `null`

**Response**
Status: 200 OK
Headers:

- `Content-Type: application/json`

Body:

- **`context`**:
  - **`error`** — error object/value, `null` if no error occured
  - **`logs`**:
    - **`stdout`** — array of log strings printed by the function to stdout
    - **`stderr`** — array of log strings printed by the function to stderr
- **`payload`** — output object/value, can be `null`

Payloads must be JSON encodable. 


# Function Creation Flow
## Register a *Base Image*

A **user** registers an existing docker image as a *base image* in Dispatch:

    $ dispatch create base-image js dispatchframework/nodejs-base:0.0.2

Here, `dispatchframework/nodejs-base:0.0.2` is registered as the *base image* named `js`.


## Create an *Image*

A **user** creates an *image* to be used for their functions, including any needed system packages and library dependencies:

    $ dispatch create image js-deps js --runtime-deps ./package.json

Here, the *image* named `js-deps` is created from the *base image* `js` adding dependencies from the manifest file `./package.json`. 

In order to create the *image*, **Dispatch image manager** does the following:

1. creates (doesn’t run) a temporary container from the *base image*
2. copies the directory specified by the metadata label `io.dispatchframework.imageTemplate` (in our case, `/image-template`) from the container into a temporary directory
3. copies the manifest file (`package.json`) into the same directory
4. builds the docker image from the temporary directory, using the *base image* docker image as the value of `BASE_IMAGE` build argument
5. registers the docker image as a Dispatch *image*.


## Create a *Function*

A **user** creates a *function* from a source file using the specified *image*: 

    $ dispatch create function js-deps hello1 ./hello.js

Here, the *function* named `hello1` is created from the source file `./hello.js` using the image `js-deps` containing library packages that can be used by the *function*. 

In order to create the function, **Dispatch function manager** does the following:

1. creates (doesn’t run) a temporary container from the specified *image*
2. copies the directory specified by the metadata label `io.dispatchframework.functionTemplate` (in our case, `/function-template`) from the container into a temporary directory
3. copies the function source file into the same directory
4. builds the docker image from the temporary directory, using the *image* docker image as the value of `IMAGE` build argument
5. registers the docker image as the *function* image.


# Summary

*Base-image* contains everything required to support a language:

- The language runtime and its standard library 
- A simple HTTP server implementing Function Runtime API
- *Image* Dockerfile template to install system and runtime dependency packages
- *Function* image Dockerfile template


# Examples

There is a selection of base-images implementing this spec in [dispatchframework](https://github.com/dispatchframework) organization on GitHub:

- https://github.com/dispatchframework/nodejs-base-image
- https://github.com/dispatchframework/python3-base-image
- https://github.com/dispatchframework/powershell-base-image
- https://github.com/dispatchframework/java-base-image
- https://github.com/dispatchframework/clojure-base-image

Any one of these images can be used as an example of how to add a new language support to Dispatch.
