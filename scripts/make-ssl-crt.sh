#!/bin/bash
set -e -x

NAMESPACE=${1:-default}
DOMAIN=${2:-dispatch.vmware.com}

openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=${DOMAIN}/O=${DOMAIN}"
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout api-tls.key -out api-tls.crt -subj "/CN=api.${DOMAIN}/O=api.${DOMAIN}"

kubectl delete secret tls dispatch-tls -n=$NAMESPACE 2>/dev/null || :
kubectl delete secret tls api-dispatch-tls -n=kong 2>/dev/null || :

kubectl create namespace $NAMESPACE 2>/dev/null || :
kubectl create secret tls dispatch-tls -n=$NAMESPACE --key tls.key --cert tls.crt
# HACK, should probably not create separate namespace for kong
kubectl create namespace kong 2>/dev/null || :
kubectl create secret tls api-dispatch-tls -n=kong --key api-tls.key --cert api-tls.crt
