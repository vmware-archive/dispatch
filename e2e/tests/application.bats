#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Application creation" {

    run_with_retry "dispatch get application --json | jq -r length" 0 1 1
    assert_success

    # create applications "foo" and "bar"
    run dispatch create application foo-app
    assert_success
    run dispatch create app bar-app
    assert_success

    run_with_retry "dispatch get app foo-app --json | jq -r .status" "READY" 4 5
    run_with_retry "dispatch get app bar-app --json | jq -r .status" "READY" 4 5
}

@test "Base image creation" {

    run dispatch get base-image
    assert_success

    # Create base image "base-nodejs"
    run dispatch create base-image base-nodejs $DOCKER_REGISTRY/$BASE_IMAGE_NODEJS6 --language nodejs
    assert_success

    run_with_retry "dispatch get base-image base-nodejs --json | jq -r .status" "READY" 4 5
}

@test "Image creation" {

    run_with_retry "dispatch get image --json | jq -r length" 0 1 1
    assert_success

    run dispatch create image foo-app-image base-nodejs --application foo-app
    assert_success
    run dispatch create image bar-app-image base-nodejs --application bar-app
    assert_success

    # list image of foo app, only one returns
    run_with_retry "dispatch get image --application foo-app --json | jq -r length" 1 1 1
    run_with_retry "dispatch get image --json | jq -r length" 2 1 1

    # list image of a non-exist application => no result
    run_with_retry "dispatch get image --application boo --json | jq length" 0 1 1

    # get image by name and application
    run_with_retry "dispatch get image foo-app-image --application foo-app --json | jq -r .status" "READY" 4 5
    run_with_retry "dispatch get image bar-app-image --application bar-app --json | jq -r .status" "READY" 4 5

    # get image of a wrong applicatio name => no result
    run_with_retry "dispatch get image foo-app-image --application bar-app --json | jq length" 0 1 1
}


@test "Secret creation" {

    run dispatch create secret open-sesame -a foo-app ${DISPATCH_ROOT}/examples/nodejs/secret.json
    echo_to_log
    assert_success

    run dispatch create secret open-sesame-bar -a bar-app ${DISPATCH_ROOT}/examples/nodejs/secret.json
    echo_to_log
    assert_success

    # list secret of foo app, only one returns
    run_with_retry "dispatch get secret --application foo-app --json | jq length" 1 1 1
}

@test "Function creation" {

    run dispatch create function --image=foo-app-image foo-app-func -a foo-app ${DISPATCH_ROOT}/examples/nodejs --handler=./i-have-a-secret.js --secret open-sesame
    echo_to_log
    assert_success

    run_with_retry "dispatch get function -a foo-app --json | jq -r .[].status" "READY" 20 5
}

@test "Function execution" {
    run_with_retry "dispatch exec foo-app-func -a foo-app --wait --json | jq -r .output.message" "The password is OpenSesame" 5 5
}

@test "API creation" {

    run dispatch create api foo-app-api foo-app-func --application foo-app -m POST -p /foo --auth public
    echo_to_log
    assert_success

    # list apis of app foo
    run_with_retry "dispatch get api -a foo-app --json | jq length" 1 1 1

    # list with an wrong application name, => no result
    run_with_retry "dispatch get api -a bar-app --json | jq length" 0 1 1

    # get with application flag
    run_with_retry "dispatch get api foo-app-api -a foo-app --json | jq -r .status" "READY" 6 5

    # get with wrong application name => no result
    run_with_retry "dispatch get api foo-app-api -a bar-app --json |  jq length" 0 1 1
}

@test "Event" {

    run dispatch create subscription -a foo-app --name app-one-sub --event-type foo-app-one.event foo-app-func
    echo_to_log
    assert_success

    run dispatch create subscription -a foo-app --name app-two-sub --event-type foo-app-two.event foo-app-func
    echo_to_log
    assert_success

    # list subscriptions
    run_with_retry "dispatch get subscription -a foo-app --json | jq length" 2 1 1

    # get subscription by name with an application flag
    run_with_retry "dispatch get subscription app-one-sub -a foo-app --json | jq -r .status" "READY" 4 5

    # get with wrong application name, no result
    run_with_retry "dispatch get subscription app-one-sub -a bar-app --json | jq length" 0 1 1
}

@test "Cleanup" {
    cleanup
}