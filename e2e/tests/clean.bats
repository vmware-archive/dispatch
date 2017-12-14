#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Clean all" {
    cleanup
}