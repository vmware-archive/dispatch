#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Base image creation" {

    run dispatch get base-image
    assert_success

    # Create base image "base-nodejs6"
    run dispatch create base-image base-nodejs6 $DOCKER_REGISTRY/$BASE_IMAGE_NODEJS6 --language nodejs6
    assert_success

    # Ensure starting status is "INITIALIZED". Wait 20 seconds for status "READY"
    run_with_retry "dispatch get base-image base-nodejs6 --json | jq -r .status" "INITIALIZED" 1 0
    run_with_retry "dispatch get base-image base-nodejs6 --json | jq -r .status" "READY" 4 5

    # Create base image "base-python3"
    run dispatch create base-image base-python3 $DOCKER_REGISTRY/$BASE_IMAGE_PYTHON3 --language python3
    assert_success

    # Ensure starting status is "INITIALIZED". Wait 20 seconds for status "READY"
    run_with_retry "dispatch get base-image base-python3 --json | jq -r .status" "INITIALIZED" 1 0
    run_with_retry "dispatch get base-image base-python3 --json | jq -r .status" "READY" 4 5

    # Create base image "base-powershell"
    run dispatch create base-image base-powershell $DOCKER_REGISTRY/$BASE_IMAGE_POWERSHELL --language powershell
    assert_success

    run_with_retry "dispatch get base-image base-powershell --json | jq -r .status" "INITIALIZED" 1 0
    run_with_retry "dispatch get base-image base-powershell --json | jq -r .status" "READY" 10 5

    # Create third image with non-existing image. Check that get operation returns three images. Wait for "ERROR" status for missing image.
    run dispatch create base-image missing-image missing/image:latest --language nodejs6
    assert_success
    run bash -c "dispatch get base-image --json | jq '. | length'"
    assert_equal 4 $output
    run_with_retry "dispatch get base-image missing-image --json | jq -r .status" "ERROR" 4 5
}

@test "Image creation" {
    run dispatch create image nodejs6 base-nodejs6
    assert_success
    run dispatch create image python3 base-python3
    assert_success
    run dispatch create image powershell base-powershell
    run_with_retry "dispatch get image nodejs6 --json | jq -r .status" "READY" 8 5
    run_with_retry "dispatch get image python3 --json | jq -r .status" "READY" 8 5
    run_with_retry "dispatch get image powershell --json | jq -r .status" "READY" 8 5
}

@test "Delete images" {
    run bash -c "dispatch get images --json | jq -r .[].name"
    assert_success
    for i in $output; do
        run dispatch delete image $i
    done
    run_with_retry "dispatch get images --json | jq '. | length'" 0 4 5
}

@test "Delete base images" {
    run bash -c "dispatch get base-images --json | jq -r .[].name"
    assert_success
    for i in $output; do
        run dispatch delete base-image $i
    done
    run_with_retry "dispatch get base-images --json | jq '. | length'" 0 4 5
}

@test "Batch load images" {
    batch_create_images images.yaml
}

@test "Update base images" {
    run dispatch update base-image nodejs6-base --language python3
    assert_success
    run_with_retry "dispatch get base-image nodejs6-base --json | jq -r .language" "python3" 1 0
    assert_success

    run dispatch update base-image not-found --language nodejs
    assert_failure
}

@test "Batch delete images" {
    batch_delete_images images.yaml
}

@test "Cleanup" {
    cleanup
}