#!/bin/bash

### COMMON VARIABLES ###
: ${FAAS:="openfaas"}
: ${DOCKER_REGISTRY:="dispatchframework"}
: ${BASE_IMAGE_PYTHON3:="python3-base:0.0.6"}
: ${BASE_IMAGE_NODEJS6:="nodejs-base:0.0.6"}
: ${BASE_IMAGE_POWERSHELL:="powershell-base:0.0.7"}
: ${BASE_IMAGE_JAVA:="java-base:0.0.6"}