#!/usr/bin/env bats

set -o pipefail

load helpers

@test "Test base image creation" {

    # Create base image "photon-nodejs6"
    run vs create base-image photon-nodejs6 $DOCKER_REGISTRY/$BASE_IMAGE_FUNC_DEPS --language nodejs6 --public 
    [ "0" -eq $status ]
    
    # Ensure starting status is "INITIALIZING". Wait 20 seconds for status "READY"
    run vs get base-image photon-nodejs6 --json | jq -r .status
    values_equal 0 $status
    values_equal "INITIALIZING" $output
    retry_timeout "[[ $(vs get base-image photon-nodejs6 --json | jq -r .status) == 'READY' ]]" 20
    
    # Create second image. check that get operation returns two images
    run vs create base-image photon-nodejs6-to-delete $DOCKER_REGISTRY/$BASE_IMAGE_FUNC_DEPS --language nodejs6 --public
    values_equal 0 $status
    run vs get base-image --json | jq '. | length'
    values_equal 2 $output

    # Create third image with non-existing image. Check that get operation returns three images. Wait for "ERROR" status for missing image.
    run vs create base-image missing-image missing/image:latest --language nodejs6 --public
    values_equal 0 $status
    run vs get base-image --json | jq '. | length'
    values_equal 3 $output
    retry_timeout "[[ $(vs get base-image missing-image --json | jq -r .status) == 'ERROR' ]]" 20

    # Delete the base image to delete. wait for get operation to return 2 images.
    run vs delete base-image photon-nodejs6-to-delete
    values_equal 0 $status
    retry_timeout "[[ $(vs get base-image --json | jq '. | length') == 'ERROR' ]]" 20

}

@test "Test image creation" {
    run vs create image base-node photon-nodejs6
    values_equal 0 $status

    run vs get image base-node --json | jq -r .status
    values_equal "READY" $output 
}