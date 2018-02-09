#!/bin/bash

set -e +x -u

if [[ -n ${GKE_PROJECT_ID} ]]; then
    source dispatch/ci/e2e/config-gke-env.sh
    cluster_name=$(cat cluster/name)
    gcloud container clusters get-credentials ${cluster_name}

else
    export NODE_IP=$(cat cluster/metadata | jq -r '.nodeIP')
    export K8S_URL=$(cat cluster/metadata | jq -r '.k8sURL')
    export KUBE_USERNAME=$(cat cluster/metadata | jq -r '.k8sUsername')
    export KUBE_PASSWORD=$(cat cluster/metadata | jq -r '.k8sPassword')

    echo "${NODE_IP} api.dispatch.local dispatch.local" >> /etc/hosts

    kubectl config set-cluster ci --insecure-skip-tls-verify=true --server=${K8S_URL}
    kubectl config set-credentials cluster-admin --username=${KUBE_USERNAME} --password=${KUBE_PASSWORD}
    kubectl config set-context ci --cluster=ci --namespace=default --user=cluster-admin
    kubectl config use-context ci
fi

helm init -c
helm repo remove local

set -x