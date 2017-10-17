#! /bin/bash

: ${VALUES_PATH:="values.yaml"}

cat << EOF > $VALUES_PATH
global:
    tag: ${1}
EOF