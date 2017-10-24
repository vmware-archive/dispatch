#!/bin/bash

### COMMON VARIABLES ###

: ${DOCKER_REGISTRY:="serverless-docker-local.artifactory.eng.vmware.com"}
: ${BASE_IMAGE_FUNC_DEPS:="openfaas-nodejs-base:0.0.1-dev1"}




### COMMON FUNCTIONS ###

# retry_timeout takes 2 args: command [timeout (secs)]
retry_timeout () {
  count=0
  while [[ ! `eval $1` ]]; do
    sleep 1
    count=$((count+1))
    if [[ "$count" -gt $2 ]]; then
      return 1
    fi
  done
}
 
# values_equal takes 2 values, both must be non-null and equal
values_equal () {
  if [[ "X$1" != "X" ]] || [[ "X$2" != "X" ]] && [[ $1 == $2 ]]; then
    return 0
  else
    return 1
  fi
}
 
# min_value_met takes 2 values, both must be non-null and 2 must be equal or greater than 1
min_value_met () {
  if [[ "X$1" != "X" ]] || [[ "X$2" != "X" ]] && [[ $2 -ge $1 ]]; then
    return 0
  else
    return 1
  fi
}
