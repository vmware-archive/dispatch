#!/bin/bash
# Post install script for configuring knative for dispatch

: ${DISPATCH_NS:="dispatch-server"}
: ${SERVING_NS:="knative-serving"}

# Get the clusterIP address of the docker registry
DOCKER_REG_IP=$(kubectl get service --namespace ${DISPATCH_NS} ${DISPATCH_NS}-docker-registry -o jsonpath="{.spec.clusterIP}")
# Update the knative serving controller to skip validation for the registry (it's insecure)
kubectl -n $SERVING_NS get configmap config-controller -o yaml | sed 's/registriesSkippingTagResolving:\(.*\)/registriesSkippingTagResolving:\1,'${DOCKER_REG_IP}:5000'/g' > edited.yaml
kubectl apply -f edited.yaml
# Bounce the controller to take advantage of the new config
POD_NAME=$(kubectl get pods --namespace $SERVING_NS -l "app=controller" -o jsonpath="{.items[0].metadata.name}")
kubectl -n $SERVING_NS delete pod $POD_NAME