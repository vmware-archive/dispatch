#!/bin/bash
# Post install script for configuring knative for dispatch

: ${DISPATCH_NAMESPACE:="dispatch"}
: ${SERVING_NAMESPACE:="knative-serving"}
: ${RELEASE_NAME:="knative-serving"}

# Get the clusterIP address of the docker registry
DOCKER_REG_IP=$(kubectl get service --namespace ${DISPATCH_NAMESPACE} ${RELEASE_NAME}-docker-registry -o jsonpath="{.spec.clusterIP}")
# Update the knative serving controller to skip validation for the registry (it's insecure)
kubectl -n $SERVING_NAMESPACE get configmap config-controller -o yaml | sed 's/registriesSkippingTagResolving:\(.*\)/registriesSkippingTagResolving:\1,'${DOCKER_REG_IP}:5000'/g' > edited.yaml
kubectl apply -f edited.yaml
# Bounce the controller to take advantage of the new config
POD_NAME=$(kubectl get pods --namespace $SERVING_NAMESPACE -l "app=controller" -o jsonpath="{.items[0].metadata.name}")
kubectl -n $SERVING_NAMESPACE delete pod $POD_NAME