#!/bin/bash

# This scripts provides a quick way to run a kong server on docker, good for local test

# Note this script is only used for local unit test for kong

# You should have docker runtime on your localhost

set -e
set -x

docker build -t xueyangh-kong -f images/kong/Dockerfile .

PREFIX=unit-test-kong
DB=${PREFIX}-database
KONG=${PREFIX}-server
PORT=8002

docker rm -f ${DB}  > /dev/null 2>&1 || true
docker rm -f ${KONG} > /dev/null 2>&1 || true

docker run -d --name $DB \
              -p 5432:5432 \
              -e "POSTGRES_USER=kong" \
              -e "POSTGRES_DB=kong" \
              postgres:9.4

# wait for the db get ready
sleep 5

docker run --name ${KONG} \
    --link ${DB}:${DB} \
    -p ${PORT}:8001 \
    -p 8444:8444 \
    -p 8000:8000 \
    -p 8443:8443 \
    -e "KONG_NGINX_DAEMON='off'" \
    -e "KONG_DATABASE=postgres" \
    -e "KONG_PG_HOST=${DB}" \
    xueyangh-kong:latest kong start --conf /etc/kong/kong.conf --nginx-conf /nginx.conf --run-migrations --vv

echo "kong has installed at your local docker host, with ADMIN port ${PORT} and PROXY port 8000"
