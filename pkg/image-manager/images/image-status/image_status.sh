#! /bin/sh

docker pull $@ &>/dev/null
echo "{\"result\": $?}"