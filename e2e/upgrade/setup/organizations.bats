#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Create resources in test-org-a" {

    ##################
    # Setup test-org-a
    ##################
    tmp_dir_a=$(mktemp -d)
    setup_test_org test-org-a ${tmp_dir_a}

    # Unset the CI accounts (if any)
    unset DISPATCH_SERVICE_ACCOUNT
    unset DISPATCH_JWT_PRIVATE_KEY
    unset DISPATCH_ORGANIZATION

    #####################
    # Login to test-org-a
    #####################
    # Set a custom config file to test login
    export DISPATCH_CONFIG=${tmp_dir_a}/config.json
    cp ~/.dispatch/config.json ${DISPATCH_CONFIG}
    run dispatch login --service-account test-org-a-user --jwt-private-key ${tmp_dir_a}/private.key --organization test-org-a
    echo_to_log
    assert_success

    # Create images in test-org-a
    batch_create_images

    # Create function in test-org-a
    run dispatch create function --image=nodejs node-hello-no-schema ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-no-schema --json | jq -r .status" "READY" 15 5

    # Should not be able to view resources in test-org-b
    run dispatch get functions --organization test-org-b
    echo_to_log
    assert_failure

    run dispatch create api api-test-https-only node-hello-no-schema -m POST --https-only -p /https-only --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test-https-only --json | jq -r .status" "READY" 6 5
    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTPS_HOST}/test-org-a/https-only -k -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"HTTPS ONLY\"
        }' | jq -r .myField" "Hello, VMware from HTTPS ONLY" 6 5

    run dispatch logout
    assert_success

    rm -r ${tmp_dir_a}
    unset DISPATCH_CONFIG
}