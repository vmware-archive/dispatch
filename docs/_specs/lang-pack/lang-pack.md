---
layout: default
---
# Decouple language support from FaaS drivers
Adding language support could be trivial if FaaS-specific function adapter only needed to use one interface to call function implementations. All language-specific runtime code could be packaged in a single docker image. 


# Problem

Currently, it’s hard to add support for new languages to Dispatch: FaaS/language specific code (function wrapper) for M drivers and N languages, has to be implemented M×N times. 

1. To add a new language, it should be added to all FaaSs currently supported in a FaaS-specific way. 
2. To add a support for a new FaaS to Dispatch, one not only needs to implement the FaaS driver, but also add the languages in a FaaS-specific way. As a consequence, language support differs between FaaS-drivers.
3. It’s impossible to add a language without rebuilding Dispatch.


# Proposed solution

Base-image should include everything required to support a language, including:

- The language runtime and its standard library — usually installed as a system package
- Function adapter — a simple HTTP server implementing Function Runtime API (see below)
- Dockerfile template to build the function image

At function image building phase, function-manager would extract the template files from the base-image and render them using the function source code and metadata. The resulting Dockerfile are used to build the function image. 

For riff and openfaas we don't need FaaS specific files in the final function image: 
- riff uses a sidecar container to communicate with the function
- openfaas function container just needs to expose HTTP.

For other/future FaaS drivers, we will similarly inject FaaS specific files at function image build time. Ideally, these files will be packaged in FaaS driver images which is TBD in a future spec.


## Function Runtime HTTP API

It’s a low level function invocation API typically implemented by a simple HTTP server written in the language we’re building the base image for. This API is called by the FaaS driver (through the FaaS engine) at function invocation time: 

- OpenFaaS would call it through its faas-netes proxy
- Riff would call it through the function-sidecar

Endpoint: /

**Request**
Method: POST
Headers:

- Content-Type: application/json
- Accept: application/json

Body:

- context
- payload

**Response**
Headers:

- Content-Type: application/json

Body:

- context
  - error — error object/value, null if no error occured
  - logs — array of log strings from the function run
- payload


## Required changes

**Base Images**

1. Add the Function Runtime HTTP API server.
2. Add function image Dockerfile template. 

**Function-manager**

1. Remove built-in FaaS- and language-specific function image templates.
2. Extract function image templates from the base-image.

