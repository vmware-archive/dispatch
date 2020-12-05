#! /bin/bash

: ${DISPATCH_SERVER_DOCKER_REPOSITORY:="vmware"}
: ${VALUES_PATH:="values.yaml"}

cat << EOF > $VALUES_PATH
image:
  host: ${DISPATCH_SERVER_DOCKER_REPOSITORY}
  tag: ${TAG}
registry:
  url: http://dispatch-docker-registry:5000/
  repository: dispatch-docker-registry:5000
storage:
  minio:
    address: dispatch-minio:9000
    username: ${MINIO_USERNAME}
    password: ${MINIO_PASSWORD}
EOF
