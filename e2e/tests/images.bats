#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Base image creation" {

    run dispatch get base-image
    assert_success

    # Create base image "base-nodejs"
    run dispatch create base-image base-nodejs $DOCKER_REGISTRY/$BASE_IMAGE_NODEJS6 --language nodejs
    assert_success

    # Ensure starting status is "INITIALIZED". Wait 20 seconds for status "READY"
    run_with_retry "dispatch get base-image base-nodejs -o json | jq -r .status" "INITIALIZED" 1 0
    run_with_retry "dispatch get base-image base-nodejs -o json | jq -r .status" "READY"

    # Create base image "base-python3"
    run dispatch create base-image base-python3 $DOCKER_REGISTRY/$BASE_IMAGE_PYTHON3 --language python3
    assert_success

    # Ensure starting status is "INITIALIZED". Wait 20 seconds for status "READY"
    run_with_retry "dispatch get base-image base-python3 -o json | jq -r .status" "INITIALIZED" 1 0
    run_with_retry "dispatch get base-image base-python3 -o json | jq -r .status" "READY"

    # Create base image "base-powershell"
    run dispatch create base-image base-powershell $DOCKER_REGISTRY/$BASE_IMAGE_POWERSHELL --language powershell
    assert_success

    run_with_retry "dispatch get base-image base-powershell -o json | jq -r .status" "INITIALIZED" 1 0
    run_with_retry "dispatch get base-image base-powershell -o json | jq -r .status" "READY"

    # Create base image "base-java"
    run dispatch create base-image base-java $DOCKER_REGISTRY/$BASE_IMAGE_JAVA --language java
    assert_success

    run_with_retry "dispatch get base-image base-java -o json | jq -r .status" "INITIALIZED" 1 0
    run_with_retry "dispatch get base-image base-java -o json | jq -r .status" "READY"

    # Create fifth image with non-existing image. Check that get operation returns five images. Wait for "ERROR" status for missing image.
    run dispatch create base-image missing-image missing/image:latest --language nodejs
    assert_success
    run bash -c "dispatch get base-image -o json | jq '. | length'"
    assert_equal 5 $output
    run_with_retry "dispatch get base-image missing-image -o json | jq -r .status" "ERROR"
}

@test "Image creation" {
    run dispatch create image nodejs base-nodejs
    assert_success
    run dispatch create image python3 base-python3
    assert_success
    run dispatch create image powershell base-powershell
    assert_success
    run dispatch create image java base-java
    assert_success
    run_with_retry "dispatch get image nodejs -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image python3 -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image powershell -o json | jq -r .status" "READY"
    run_with_retry "dispatch get image java -o json | jq -r .status" "READY"
}

@test "Delete images" {
    run bash -c "dispatch get images -o json | jq -r .[].name"
    assert_success
    for i in $output; do
        run dispatch delete image $i
    done
    run_with_retry "dispatch get images -o json | jq '. | length'" 0
}

@test "Delete base images" {
    run bash -c "dispatch get base-images -o json | jq -r .[].name"
    assert_success
    for i in $output; do
        run dispatch delete base-image $i
    done
    run_with_retry "dispatch get base-images -o json | jq '. | length'" 0
}

@test "Batch load images" {
    batch_create_images
}

@test "Create with URL" {
    skip "skipping seed create"
    run dispatch create --file=https://raw.githubusercontent.com/vmware/dispatch/master/examples/seed.yaml --work-dir=test_create_by_url
    assert_success

    run_with_retry "dispatch get function hello-js -o json | jq -r .status" "READY"
    run_with_retry "dispatch get function hello-py -o json | jq -r .status" "READY"
    run_with_retry "dispatch get function http-py -o json | jq -r .status" "READY"
    run_with_retry "dispatch get function hello-ps1 -o json | jq -r .status" "READY"
}

@test "Update images" {
    run dispatch update --work-dir ${BATS_TEST_DIRNAME} -f images_update.yaml
    assert_success

    run_with_retry "dispatch get image python3 -o json | jq -r .status" "READY"
    run_with_retry "dispatch get base-image nodejs-base -o json | jq -r .status" "READY"

    run_with_retry "dispatch get base-image nodejs-base -o json | jq -r .language" "python3" 1 0
    assert_success

    run_with_retry "dispatch get image python3 -o json | jq -r .tags[0].key" "update" 1 0
    assert_success

    run bash -c "dispatch update --work-dir ${BATS_TEST_DIRNAME} -f image_not_found_update.yaml"
    assert_failure
}

@test "Batch delete images" {
    batch_delete_images ${IMAGES_YAML}
}

@test "Cleanup" {
    cleanup
}