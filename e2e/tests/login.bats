#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Login with service account" {

    svc_acct=${DISPATCH_SERVICE_ACCOUNT}
    pri_key=${DISPATCH_JWT_PRIVATE_KEY}
    unset DISPATCH_SERVICE_ACCOUNT
    unset DISPATCH_JWT_PRIVATE_KEY

    run dispatch login --service-account ${svc_acct} --jwt-private-key ${pri_key}
    echo_to_log

    run dispatch get functions
    assert_success

    export DISPATCH_SERVICE_ACCOUNT=${svc_acct}
    export DISPATCH_JWT_PRIVATE_KEY=${pri_key}

    # using enviroment var in CI, delete login info in dispatch config
    run dispatch logout
}

@test "Login with invalid service account" {

    svc_acct=${DISPATCH_SERVICE_ACCOUNT}
    pri_key=${DISPATCH_JWT_PRIVATE_KEY}
    unset DISPATCH_SERVICE_ACCOUNT
    unset DISPATCH_JWT_PRIVATE_KEY

    run dispatch login --service-account ${svc_acct}_invalid --jwt-private-key ${pri_key}
    echo_to_log

    run dispatch get functions
    assert_failure

    export DISPATCH_SERVICE_ACCOUNT=${svc_acct}
    export DISPATCH_JWT_PRIVATE_KEY=${pri_key}

    # using enviroment var, delete login info in dispatch config
    run dispatch logout
}

@test "Login with service account with invalid private key" {

    svc_acct=${DISPATCH_SERVICE_ACCOUNT}
    pri_key=${DISPATCH_JWT_PRIVATE_KEY}
    unset DISPATCH_SERVICE_ACCOUNT
    unset DISPATCH_JWT_PRIVATE_KEY

    run dispatch login --service-account ${svc_acct}_invalid --jwt-private-key ${pri_key}_invalid
    echo_to_log

    run dispatch get functions
    assert_failure

    export DISPATCH_SERVICE_ACCOUNT=${svc_acct}
    export DISPATCH_JWT_PRIVATE_KEY=${pri_key}

    # using enviroment var in CI, delete login info in dispatch config
    run dispatch logout
}
