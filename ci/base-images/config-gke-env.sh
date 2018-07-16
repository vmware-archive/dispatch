#!/bin/bash

# config gke environment: account, project and zone
set -e +x -u

: ${GKE_ZONE:="us-west1-c"}

echo "Setting up GKE key"
echo ${GKE_KEY} | base64 --decode --ignore-garbage > ${HOME}/gcloud-service-key.json
echo "Activating GKE account"
gcloud auth activate-service-account --key-file ${HOME}/gcloud-service-key.json >/dev/null 2>&1
echo "Setting GKE project"
gcloud config set project ${GKE_PROJECT_ID} >/dev/null 2>&1
echo "Setting default GKE compute zone"
gcloud config set compute/zone ${GKE_ZONE} >/dev/null 2>&1

set -x
