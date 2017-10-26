#! /bin/bash
set -xe

: ${MINIO_URL:="http://10.126.236.44:9000"}

cat << EOF > charts/serverless/Chart.yaml
apiVersion: v1
description: A Helm chart for Kubernetes
name: serverless
version: $CI_COMMIT_TAG
EOF

cat charts/serverless/Chart.yaml

cat << EOF > charts/serverless/values.yaml
global:
  pullPolicy: IfNotPresent
  organization: serverless
  tag: $CI_COMMIT_TAG
  debug: false
  data:
    persist: true
image-manager:
  enabled: true
function-manager:
  enabled: true
identity-manager:
  enabled: true
secret-store:
  enabled: true
echo-server:
  enabled: true
oauth2-proxy:
  enabled: true
EOF

cat charts/serverless/values.yaml

curl -OL $MINIO_URL/charts/index.yaml
helm init -c
helm package charts/serverless -d charts/serverless
helm repo index charts/serverless/ --merge index.yaml --url $MINIO_URL/charts
mc config host add serverless $MINIO_URL $MINIO_ACCESS_KEY $MINIO_SECRET_KEY S3v4
mc cp charts/serverless/index.yaml serverless/charts/
mc cp charts/serverless/*.tgz serverless/charts/
mc mb -p serverless/vs-$CI_COMMIT_TAG-darwin serverless/vs-$CI_COMMIT_TAG-linux
mc mc policy download serverless/vs-$CI_COMMIT_TAG-darwin
mc mc policy download serverless/vs-$CI_COMMIT_TAG-linux
mc cp bin/vs-darwin serverless/vs-$CI_COMMIT_TAG-darwin/vs
mc cp bin/vs-linux serverless/vs-$CI_COMMIT_TAG-linux/vs
