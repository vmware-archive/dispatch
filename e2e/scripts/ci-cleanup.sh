#!/bin/bash

helm delete --purge vmware-serverless 
helm delete --purge openfaas 
helm delete --purge ${APIGATEWAY_NAME}
kubectl delete namespace $VS_NAMESPACE
kubectl delete namespace $FAAS_NAMESPACE
kubectl delete namespace $APIGATEWAY_NAMESPACE