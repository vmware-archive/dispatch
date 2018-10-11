#! /bin/bash

: ${DOCKER_URL:="https://index.docker.io/v1/"}
: ${DOCKER_REPOSITORY:="vmware"}
: ${VALUES_PATH:="values.yaml"}

cat << EOF > $VALUES_PATH
image:
  host: ${DOCKER_REPOSITORY}
  tag: ${TAG}
storage:
  minio:
    username: ${MINIO_USERNAME}
    password: ${MINIO_PASSWORD}
minio:
  accessKey: ${MINIO_USERNAME}
  secretKey: ${MINIO_PASSWORD}
registry:
  insecure: false
  # Use https://index.docker.io/v1/ for dockerhub
  url: ${DOCKER_URL}
  repository: ${DOCKER_REPOSITORY}
  username: ${DOCKER_USERNAME}
  password: ${DOCKER_PASSWORD}
debug: ${DISPATCH_DEBUG}
EOF
