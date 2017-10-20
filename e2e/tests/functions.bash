#!/usr/bin/env bats

set -o pipefail

load helpers

@test "Test function creation" {

    run vs create function base-node hello-no-schema examples/nodejs6/hello.js
    values_equal 0 $status

    run vs get function hello-no-schema --json | jq -r .name
    values_equal "hello-no-schema" $output
}