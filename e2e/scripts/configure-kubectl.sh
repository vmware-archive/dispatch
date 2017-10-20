#!/bin/bash

kubectl config set-cluster ci --insecure-skip-tls-verify=true --server=https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT
kubectl config set-credentials cluster-admin --token=$(</var/run/secrets/kubernetes.io/serviceaccount/token)
kubectl config set-context ci --cluster=ci --namespace=$CI_NAMESPACE --user=cluster-admin
kubectl config use-context ci