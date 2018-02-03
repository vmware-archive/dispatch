#!/bin/bash

### COMMON VARIABLES ###
: ${DOCKER_REGISTRY:="vmware"}
: ${BASE_IMAGE_PYTHON3:="dispatch-openfaas-python-base:0.0.5-dev1"}
: ${BASE_IMAGE_NODEJS6:="dispatch-openfaas-nodejs6-base:0.0.3-dev1"}

DISPATCH_CONFIG_ROOT=~/.dispatch
if [ "${CI}" = true ] ; then
    DISPATCH_CONFIG_ROOT=${DISPATCH_ROOT}
fi
API_GATEWAY_HTTPS_PORT=$(cat ${DISPATCH_CONFIG_ROOT}/config.json | jq  '.["api-https-port"]')
API_GATEWAY_HTTP_PORT=$(cat ${DISPATCH_CONFIG_ROOT}/config.json | jq  '.["api-http-port"]')

: ${API_GATEWAY_HTTPS_HOST:="https://api.dev.dispatch.vmware.com:${API_GATEWAY_HTTPS_PORT}"}
: ${API_GATEWAY_HTTP_HOST:="http://api.dev.dispatch.vmware.com:${API_GATEWAY_HTTP_PORT}"}