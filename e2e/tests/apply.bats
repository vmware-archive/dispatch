#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Apply seed images" {
    run dispatch apply seed
    assert_success

    run_with_retry "dispatch get base-image java-base -o json | jq -r .status" "READY"
    run_with_retry "dispatch get base-image nodejs-base -o json | jq -r .status" "READY"
    run_with_retry "dispatch get base-image powershell-base -o json | jq -r .status" "READY"
    run_with_retry "dispatch get base-image python3-base -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image java -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image nodejs -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image powershell -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image python3 -o json | jq -r .status" "READY"
}

@test "Apply seed images again" {
    run dispatch apply seed
    assert_success

    run_with_retry "dispatch get base-image java-base -o json | jq -r .status" "READY"
    run_with_retry "dispatch get base-image nodejs-base -o json | jq -r .status" "READY"
    run_with_retry "dispatch get base-image powershell-base -o json | jq -r .status" "READY"
    run_with_retry "dispatch get base-image python3-base -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image java -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image nodejs -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image powershell -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image python3 -o json | jq -r .status" "READY"
}

@test "Apply with URL" {
    run dispatch apply --file=https://raw.githubusercontent.com/vmware/dispatch/solo/examples/seed.yaml --work-dir=test_create_by_url
    assert_success

    run_with_retry "dispatch get function hello-js -o json | jq -r .status" "READY"
    run_with_retry "dispatch get function hello-py -o json | jq -r .status" "READY"
    run_with_retry "dispatch get function http-py -o json | jq -r .status" "READY"
    run_with_retry "dispatch get function hello-ps1 -o json | jq -r .status" "READY"

    run_with_retry "dispatch get secret open-sesame -o json | jq -r .secrets.password" "OpenSesame"
    run_with_retry "dispatch get api post-hello -o json | jq -r .status" "READY"
    run_with_retry "dispatch get eventdriver ticker -o json | jq -r .status" "READY"
    run_with_retry "dispatch get subscription ticker-sub -o json | jq -r .status" "READY"
}

@test "Delete with URL" {
    run dispatch delete --file=https://raw.githubusercontent.com/vmware/dispatch/solo/examples/seed.yaml --work-dir=test_create_by_url
    assert_success

    run_with_retry "dispatch get functions -o json | jq '. | length'" 0
    run_with_retry "dispatch get secrets -o json | jq '. | length'" 0
    run_with_retry "dispatch get api -o json | jq '. | length'" 0
    run_with_retry "dispatch get eventdriver -o json | jq '. | length'" 0
    run_with_retry "dispatch get eventdrivertype -o json | jq '. | length'" 0
    run_with_retry "dispatch get subscription -o json | jq '. | length'" 0
}

@test "Delete seed images" {
    run dispatch delete seed
    run_with_retry "dispatch get images -o json | jq '. | length'" 0
    run_with_retry "dispatch get base-images -o json | jq '. | length'" 0
}

@test "Cleanup" {
    cleanup
}
