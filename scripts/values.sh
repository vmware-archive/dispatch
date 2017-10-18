#! /bin/bash

: ${VALUES_PATH:="values.yaml"}

cat << EOF > $VALUES_PATH
global:
    tag: dev-${1}
EOF