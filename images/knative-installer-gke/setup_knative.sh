#!/bin/bash

set -ex

PROJECT=${1}
export K8S_CLUSTER_OVERRIDE="${2}"
KEY_FILE=${3}

REVISION=${4}
: ${REVISION:=origin/master}

cd ${GOPATH}/src/github.com/knative/serving
git fetch origin
git checkout ${REVISION}

gcloud auth activate-service-account --key-file ${KEY_FILE}
gcloud auth configure-docker -q

SA_USER=$(gcloud config list account --format "value(core.account)")
export K8S_USER_OVERRIDE="${SA_USER}"
export KO_DOCKER_REPO="gcr.io/${PROJECT}"
export DOCKER_REPO_OVERRIDE="${KO_DOCKER_REPO}"

kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user="${K8S_USER_OVERRIDE}"

kubectl apply -f ./third_party/istio-1.0.2/istio.yaml
kubectl apply -f ./third_party/config/build/release.yaml

ko apply -f config/
