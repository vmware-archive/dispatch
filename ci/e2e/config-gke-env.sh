#!/bin/bash

set -e +x -u

: ${GKE_ZONE:="us-west1-c"}

echo ${GKE_KEY} | base64 --decode --ignore-garbage > ${HOME}/gcloud-service-key.json
gcloud auth activate-service-account --key-file ${HOME}/gcloud-service-key.json >/dev/null 2>&1
gcloud config set project $GKE_PROJECT_ID >/dev/null 2>&1
gcloud config set compute/zone ${GKE_ZONE} >/dev/null 2>&1

set -x