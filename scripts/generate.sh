#!/bin/bash
set -e -o pipefail

: ${WORKDIR:="/root/go/src/gitlab.eng.vmware.com/serverless/serverless"}
: ${CI_IMAGE:="$(head -n 1 ./.gitlab-ci.yml | awk '{print $2}')"}

echo Using image ${CI_IMAGE}

PACKAGE=${1}
APP=${2}
SWAGGER=${3}

mkdir -p ./pkg/$PACKAGE/gen
if [ -z $CI ]; then
    docker run --rm -v `pwd`:$WORKDIR $CI_IMAGE swagger generate server -A $APP -t $WORKDIR/pkg/$PACKAGE/gen -f $WORKDIR/swagger/$SWAGGER --exclude-main
    docker run --rm -v `pwd`:$WORKDIR $CI_IMAGE swagger generate client -A $APP -t $WORKDIR/pkg/$PACKAGE/gen -f $WORKDIR/swagger/$SWAGGER
else
    echo "CI is set $CI"
    swagger generate server -A $APP -t ./pkg/$PACKAGE/gen -f ./swagger/$SWAGGER --exclude-main
    swagger generate client -A $APP -t ./pkg/$PACKAGE/gen -f ./swagger/$SWAGGER
fi
