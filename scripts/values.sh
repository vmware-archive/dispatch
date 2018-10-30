#! /bin/bash

: ${VALUES_PATH:="values.yaml"}

cat << EOF > $VALUES_PATH
image:
  host: ${DISPATCH_SERVER_DOCKER_REPOSITORY}
  tag: ${TAG}
storage:
  minio:
    username: ${MINIO_USERNAME}
    password: ${MINIO_PASSWORD}
EOF
