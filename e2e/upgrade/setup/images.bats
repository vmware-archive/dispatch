#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Batch load images" {
    batch_create_images
}