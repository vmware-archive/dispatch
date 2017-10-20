#!/bin/bash

script_path="$(dirname "$0")"
source "$script_path/../tests/helpers.bash"

for i in $(vs get function --json | jq -r .[].name); do
    vs delete function $i
done

for i in $(vs get image --json | jq -r .[].name); do
    vs delete image $i
done

for i in $(vs get base-image --json | jq -r .[].name); do
    vs delete base-image $i
done

set -e
vs get function
retry_timeout "[[ $(vs get function --json | jq '. | length') == 0 ]]" 30
retry_timeout "[[ $(vs get image --json | jq '. | length') == 0 ]]" 30
retry_timeout "[[ $(vs get base-image --json | jq '. | length') == 0 ]]" 30
