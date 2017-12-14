#!/bin/bash

cat << EOF > config.yaml
namespace: serverless
hostname: serverless.vmware.com
certificateDirectory: ci_output
chart:
  image:
    host: $DOCKER_REGISTRY
    tag: $IMAGE_TAG
repository:
  host: $DOCKER_REGISTRY
  username: $DOCKER_USERNAME
  email: serverless@vmware.com
  password: $DOCKER_PASSWORD
serviceType: LoadBalancer
apiGateway:
  serviceType: ClusterIP
EOF
