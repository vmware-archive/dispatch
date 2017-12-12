#! /bin/bash

: ${VALUES_PATH:="values.yaml"}

cat << EOF > $VALUES_PATH
global:
    image:
        tag: dev-${1}
EOF
