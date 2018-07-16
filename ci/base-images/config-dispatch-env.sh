#!/bin/bash

cp dispatch-release/dispatch-linux /usr/local/bin/dispatch
chmod +x /usr/local/bin/dispatch

export INSTALL_DISPATCH=0
export CI=true
export TERM=linux
export DISPATCH_ORGANIZATION=ci-org
export DISPATCH_SERVICE_ACCOUNT="ci-org/ci-user"
export DISPATCH_JWT_PRIVATE_KEY=$(pwd)/ci-keys/ci-user.key

mkdir -p ~/.dispatch

export LOADBALANCER_IP=$(kubectl get svc/ingress-nginx-ingress-controller -n kube-system -o json | jq -r '.status.loadBalancer.ingress[0].ip')
export API_GATEWAY_IP=$(kubectl get svc/api-gateway-kongproxy -n kong -o json | jq -r '.status.loadBalancer.ingress[0].ip')
cp dispatch/ci/base-images/configs/dispatch-config.json ~/.dispatch/config.json
sed -i "s/LOADBALANCER_IP/$LOADBALANCER_IP/g" ~/.dispatch/config.json
sed -i "s/CURRENT_CONTEXT/$(echo $LOADBALANCER_IP | tr '.' '-')/g" ~/.dispatch/config.json

export API_GATEWAY_HTTPS_HOST="https://${API_GATEWAY_IP}:443"
export API_GATEWAY_HTTP_HOST="http://${API_GATEWAY_IP}:80"
