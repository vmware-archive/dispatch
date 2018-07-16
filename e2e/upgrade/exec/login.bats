#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Login to test service account" {

    # Unset the CI accounts (if any)
    svc_acct=${DISPATCH_SERVICE_ACCOUNT}
    pri_key=${DISPATCH_JWT_PRIVATE_KEY}
    unset DISPATCH_SERVICE_ACCOUNT
    unset DISPATCH_JWT_PRIVATE_KEY

    # Set a custom config file to test login
    tmp_dir=$(mktemp -d)
    export DISPATCH_CONFIG=${tmp_dir}/config.json
    cp ~/.dispatch/config.json ${DISPATCH_CONFIG}

    # Login to test service account with invalid private key
    head -c 4096 < /dev/urandom > ${tmp_dir}/test-invalid.key
    run dispatch login --service-account ${DISPATCH_TEST_SERVICE_ACCOUNT} --jwt-private-key ${tmp_dir}/test-invalid.key
    echo_to_log
    assert_failure

    # Login to test service account with correct private key
    run dispatch login --service-account ${DISPATCH_TEST_SERVICE_ACCOUNT} --jwt-private-key ${DISPATCH_TEST_JWT_PRIVATE_KEY}
    echo_to_log
    assert_success

    run dispatch get functions
    echo_to_log
    assert_success

    run dispatch logout
    assert_success

    unset DISPATCH_CONFIG

    # Cleanup
    rm -r ${tmp_dir}

    # Reset the CI accounts (if any) and custom config file
    export DISPATCH_SERVICE_ACCOUNT=${svc_acct}
    export DISPATCH_JWT_PRIVATE_KEY=${pri_key}

    run dispatch iam delete serviceaccount ${DISPATCH_TEST_SERVICE_ACCOUNT}
    echo_to_log
    assert_success

    run dispatch iam delete policy ${DISPATCH_TEST_SERVICE_ACCOUNT}-policy
    echo_to_log
    assert_success
}