#! /bin/bash

: ${DISPATCH_NAMESPACE:="dispatch"}
: ${DISPATCH_SERVER_DOCKER_REPOSITORY:="vmware"}
: ${VALUES_PATH:="values.yaml"}

cat << EOF > $VALUES_PATH
image:
  host: ${DISPATCH_SERVER_DOCKER_REPOSITORY}
  tag: ${TAG}
registry:
  url: http://${DISPATCH_NAMESPACE}-docker-registry:5000/
  repository: ${DISPATCH_NAMESPACE}-docker-registry:5000
storage:
  minio:
    address: ${DISPATCH_NAMESPACE}-minio:9000
    username: ${MINIO_USERNAME}
    password: ${MINIO_PASSWORD}
EOF
