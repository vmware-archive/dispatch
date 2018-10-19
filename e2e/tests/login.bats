#!/usr/bin/env bats

set -o pipefail

load helpers
load variables



@test "Login with service account" {
    skip "skipping login tests"

    tmp_dir=$(mktemp -d)
    create_test_svc_account login-with-svc-acc ${tmp_dir}

    # Unset the CI accounts (if any)
    svc_acct=${DISPATCH_SERVICE_ACCOUNT}
    pri_key=${DISPATCH_JWT_PRIVATE_KEY}
    unset DISPATCH_SERVICE_ACCOUNT
    unset DISPATCH_JWT_PRIVATE_KEY

    # Set a custom config file to test login
    export DISPATCH_CONFIG=${tmp_dir}/config.json
    cp ~/.dispatch/config.json ${DISPATCH_CONFIG}
    run dispatch login --service-account login-with-svc-acc-user --jwt-private-key ${tmp_dir}/private.key
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

    delete_test_svc_account login-with-svc-acc
}

@test "Login with invalid service account" {
    skip "skipping login tests"

    tmp_dir=$(mktemp -d)
    create_test_svc_account login-invalid-svc-acc ${tmp_dir}

    # Unset the CI accounts (if any)
    svc_acct=${DISPATCH_SERVICE_ACCOUNT}
    pri_key=${DISPATCH_JWT_PRIVATE_KEY}
    unset DISPATCH_SERVICE_ACCOUNT
    unset DISPATCH_JWT_PRIVATE_KEY
    # Set a custom config file to test login
    export DISPATCH_CONFIG=${tmp_dir}/config.json
    cp ~/.dispatch/config.json ${DISPATCH_CONFIG}

    run dispatch login --service-account login-invalid-svc-acc-user-invalid --jwt-private-key ${tmp_dir}/private.key
    echo_to_log
    assert_failure

    unset DISPATCH_CONFIG

    export DISPATCH_SERVICE_ACCOUNT=${svc_acct}
    export DISPATCH_JWT_PRIVATE_KEY=${pri_key}

    delete_test_svc_account login-invalid-svc-acc
}

@test "Login with service account with invalid private key" {
    skip "skipping login tests"

    tmp_dir=$(mktemp -d)
    create_test_svc_account login-invalid-pvt-key ${tmp_dir}

    # Unset the CI accounts (if any)
    svc_acct=${DISPATCH_SERVICE_ACCOUNT}
    pri_key=${DISPATCH_JWT_PRIVATE_KEY}
    unset DISPATCH_SERVICE_ACCOUNT
    unset DISPATCH_JWT_PRIVATE_KEY

    # Set a custom config file to test login
    export DISPATCH_CONFIG=${tmp_dir}/config.json
    cp ~/.dispatch/config.json ${DISPATCH_CONFIG}
    run dispatch login --service-account login-invalid-pvt-key-user --jwt-private-key ${tmp_dir}/test-invalid.key
    echo_to_log
    assert_failure

    unset DISPATCH_CONFIG

    export DISPATCH_SERVICE_ACCOUNT=${svc_acct}
    export DISPATCH_JWT_PRIVATE_KEY=${pri_key}

    delete_test_svc_account login-invalid-pvt-key
}
