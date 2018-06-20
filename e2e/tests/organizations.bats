#!/usr/bin/env bats

set -o pipefail

load helpers
load variables



@test "Verify multi-tenancy by creating resources in two different orgs" {

    #######
    # Setup
    #######
    tmp_dir_a=$(mktemp -d)
    tmp_dir_b=$(mktemp -d)
    setup_test_org test-org-a ${tmp_dir_a}
    setup_test_org test-org-b ${tmp_dir_b}

    # Unset the CI accounts (if any)
    svc_acct=${DISPATCH_SERVICE_ACCOUNT}
    pri_key=${DISPATCH_JWT_PRIVATE_KEY}
    org=${DISPATCH_ORGANIZATION}
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

    run dispatch logout
    assert_success

    unset DISPATCH_CONFIG

    #####################
    # Login to test-org-b
    #####################
    # Set a custom config file to test login
    export DISPATCH_CONFIG=${tmp_dir_b}/config.json
    cp ~/.dispatch/config.json ${DISPATCH_CONFIG}
    run dispatch login --service-account test-org-b-user --jwt-private-key ${tmp_dir_b}/private.key --organization test-org-b
    echo_to_log
    assert_success

    # Ensure no images exist in test-org-b
    run bash -c "dispatch get base-image --json | jq '. | length'"
    assert_equal 0 $output

    # Ensure no functions exist in test-org-b
    run bash -c "dispatch get functions --json | jq '. | length'"
    assert_equal 0 $output

    # Create images in test-org-b
    batch_create_images

    # Create function with same name in test-org-b
    run dispatch create function --image=nodejs node-hello-no-schema ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-no-schema --json | jq -r .status" "READY" 15 5

    run dispatch logout
    assert_success

    unset DISPATCH_CONFIG

    #########
    # Cleanup
    #########
    # Reset the CI accounts (if any)
    export DISPATCH_SERVICE_ACCOUNT=${svc_acct}
    export DISPATCH_JWT_PRIVATE_KEY=${pri_key}
    export DISPATCH_ORGANIZATION=${org}

    # Cleanup resources
    rm -r ${tmp_dir_a} ${tmp_dir_b}
    DISPATCH_ORGANIZATION=test-org-a cleanup
    DISPATCH_ORGANIZATION=test-org-b cleanup
    delete_test_org test-org-a
    delete_test_org test-org-b

}