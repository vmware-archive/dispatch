#!/bin/bash

### COMMON VARIABLES ###
: ${DOCKER_REGISTRY:="vmware"}
: ${BASE_IMAGE_PYTHON3:="dispatch-openfaas-python-base:0.0.5-dev1"}
: ${BASE_IMAGE_NODEJS6:="dispatch-openfaas-nodejs6-base:0.0.4-dev1"}