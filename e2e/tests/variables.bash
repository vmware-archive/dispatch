#!/bin/bash

### COMMON VARIABLES ###
: ${FAAS:="openfaas"}
: ${DOCKER_REGISTRY:="vmware"}
: ${BASE_IMAGE_PYTHON3:="dispatch-python3-base:0.0.2-dev1"}
: ${BASE_IMAGE_NODEJS6:="dispatch-nodejs6-base:0.0.2-dev1"}
: ${BASE_IMAGE_POWERSHELL:="dispatch-powershell-base:0.0.3"}