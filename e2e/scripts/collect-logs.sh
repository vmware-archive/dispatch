#!/bin/bash

echo -e "\n\n"
echo "Job failed. Download logs from $CI_PROJECT_URL/-/jobs/$CI_JOB_ID/artifacts/download"
echo -e "\n\n"

kubectl logs --tail 100 -n kube-system $(kubectl get deployments -n kube-system -o=name | grep ingress-controller) > ingress_controller.out 2>&1
kubectl logs deploy/gateway -n $FAAS_NAMESPACE > openfaas_gateway.out 2>&1
kubectl logs deploy/faas-netesd -n $FAAS_NAMESPACE > openfaas_faas-netesd.out 2>&1
kubectl logs deploy/vmware-serverless-function-manager -n $VS_NAMESPACE | sed 's/\\n/\n/g' > serverless-function-manager.out 2>&1
kubectl logs deploy/vmware-serverless-image-manager -n $VS_NAMESPACE | sed 's/\\n/\n/g' > serverless-image-manager.out 2>&1
kubectl logs deploy/vmware-serverless-identity-manager -n $VS_NAMESPACE | sed 's/\\n/\n/g' > serverless-identity-manager.out 2>&1
kubectl logs deploy/vmware-serverless-secret-store -n $VS_NAMESPACE | sed 's/\\n/\n/g' > serverless-secret-store.out 2>&1
kubectl logs deploy/vmware-serverless-event-manager -n $VS_NAMESPACE | sed 's/\\n/\n/g' > serverless-event-manager.out 2>&1

kubectl logs deploy/vmware-serverless-api-manager -n $VS_NAMESPACE | sed 's/\\n/\n/g' > serverless-api-manager.out 2>&1
kubectl logs deploy/${APIGATEWAY_NAME}-kong -n $APIGATEWAY_NAMESPACE | sed 's/\\n/\n/g' > ${APIGATEWAY_NAME}-kong.out 2>&1
kubectl logs deploy/${APIGATEWAY_NAME}-postgresql -n $APIGATEWAY_NAMESPACE | sed 's/\\n/\n/g' > ${APIGATEWAY_NAME}-postgresql.out 2>&1
KONG_PODNAME=$(kubectl get pods -n=${APIGATEWAY_NAMESPACE} -o=custom-columns=NAME:.metadata.name --no-headers | grep -i ${APIGATEWAY_NAME}-kong )
kubectl exec ${KONG_PODNAME} -n=${APIGATEWAY_NAMESPACE} cat /usr/local/kong/logs/access.log > kong-access.log 2>&1
kubectl exec ${KONG_PODNAME} -n=${APIGATEWAY_NAMESPACE} cat /usr/local/kong/logs/admin_access.log > kong-admin-access.log 2>&1
kubectl exec ${KONG_PODNAME} -n=${APIGATEWAY_NAMESPACE} cat /usr/local/kong/logs/error.log > kong-error.log 2>&1