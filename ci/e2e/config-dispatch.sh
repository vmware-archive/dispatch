#!/bin/bash

set -e +x -u

export NODE_IP=$(cat cluster/metadata | jq -r '.nodeIP')
if [ -e properties/keyval.properties ]; then
    export IMAGE_TAG=$(cat properties/keyval.properties | sed 's/tag=//')
fi
export DOCKER_REGISTRY_HOST=$(cat cluster/metadata | jq -r '.registryHost')
export DOCKER_REGISTRY_USER=$(cat cluster/metadata | jq -r '.registryUser')
export DOCKER_REGISTRY_PASS=$(cat cluster/metadata | jq -r '.registryPass')
export DOCKER_REGISTRY_EMAIL=$(cat cluster/metadata | jq -r '.registryEmail')

echo "${NODE_IP} dev.dispatch.vmware.com api.dev.dispatch.vmware.com" >> /etc/hosts

cat << EOF > config.yaml
namespace: dispatch
hostname: dev.dispatch.vmware.com
certificateDirectory: /tmp
chart:
  image:
    host: ${DOCKER_REGISTRY_HOST}
    tag: ${IMAGE_TAG:-}
repository:
  host: ${DOCKER_REGISTRY_HOST}
  username: ${DOCKER_REGISTRY_USER}
  email: ${DOCKER_REGISTRY_EMAIL}
  password: ${DOCKER_REGISTRY_PASS}
oauth2Proxy:
  clientID: 8b92faa61dcc111a5bbb
  clientSecret: b4e1c35bdf8f84d13547a2d34da73bc2661f91de
  cookieSecret: NPSE1cO+WarW8q3mrVq70Q==
EOF

cp dispatch-cli/dispatch /usr/local/bin/dispatch

set -x