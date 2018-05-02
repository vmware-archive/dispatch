#!/bin/bash

### COMMON VARIABLES ###
: ${FAAS:="openfaas"}
: ${DOCKER_REGISTRY:="dispatchframework"}
: ${BASE_IMAGE_PYTHON3:="python3-base:0.0.3"}
: ${BASE_IMAGE_NODEJS6:="nodejs-base:0.0.3"}
: ${BASE_IMAGE_POWERSHELL:="powershell-base:0.0.4"}
: ${BASE_IMAGE_JAVA8:="java8-base:0.0.2"}