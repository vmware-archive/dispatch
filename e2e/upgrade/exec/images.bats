#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Image creation" {
    run dispatch create image new-nodejs nodejs-base
    assert_success
    run_with_retry "dispatch get image new-nodejs --json | jq -r .status" "READY" 8 5
}

@test "Update images" {
    run dispatch update --work-dir ${E2E_TEST_ROOT} -f images_update.yaml
    assert_success

    run_with_retry "dispatch get image python3 --json | jq -r .status" "READY" 6 30
    run_with_retry "dispatch get base-image nodejs-base --json | jq -r .status" "READY" 6 30

    run_with_retry "dispatch get base-image nodejs-base --json | jq -r .language" "python3" 1 0
    assert_success

    run_with_retry "dispatch get image python3 --json | jq -r .tags[0].key" "update" 1 0
    assert_success

    run bash -c "dispatch update --work-dir ${E2E_TEST_ROOT} -f image_not_found_update.yaml"
    assert_failure
}