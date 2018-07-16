#!/bin/bash

# config k8s cluster environment: get cluster credentials
set -e +x -u

export cluster_name=$(cat k8s-cluster/keyval.properties | grep "cluster_name" | cut -d'=' -f2)
gcloud container clusters get-credentials ${cluster_name}

helm init -c
helm repo remove local

set -x
