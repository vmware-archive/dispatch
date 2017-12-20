#! /bin/bash

mkdir -p s3/dispatch-charts

aws s3 sync s3://dispatch-charts s3/dispatch-charts

charts="dispatch nginx-ingress kong openfaas"

for i in $charts; do
    helm package -u -d s3/dispatch-charts charts/$i
done

helm repo index --merge s3/dispatch-charts/index.yaml --url https://s3-us-west-2.amazonaws.com/dispatch-charts s3/dispatch-charts

aws s3 sync s3/dispatch-charts s3://dispatch-charts --acl public-read
