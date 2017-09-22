#!/bin/bash

: ${WORKDIR:="/root/go/src/gitlab.eng.vmware.com/serverless/serverless"}
: ${CI_IMAGE:="berndtj/photon-golang-ci"}

PACKAGE=${1}
APP=${2}
SWAGGER=${3}

mkdir -p ./pkg/$PACKAGE/gen
if [ -z $CI ]; then
    docker run --rm -v `pwd`:$WORKDIR $CI_IMAGE go-bindata -o $WORKDIR/pkg/$PACKAGE/gen/bindata.go -pkg gen -prefix "$WORKDIR/swagger" $WORKDIR/swagger
    docker run --rm -v `pwd`:$WORKDIR $CI_IMAGE swagger generate server -A $APP -t $WORKDIR/pkg/$PACKAGE/gen -f $WORKDIR/swagger/$SWAGGER --exclude-main
else
    echo "CI is set $CI"
    go-bindata -o ./pkg/$PACKAGE/gen/bindata.go -pkg gen -prefix "./swagger" ./swagger
    swagger generate server -A $APP -t ./pkg/$PACKAGE/gen -f ./swagger/$SWAGGER --exclude-main
fi
