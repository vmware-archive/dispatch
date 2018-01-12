#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

: ${API_GATEWAY_HTTPS_HOST:="https://api.dev.dispatch.vmware.com:${API_GATEWAY_HTTPS_PORT}"}
: ${API_GATEWAY_HTTP_HOST:="http://api.dev.dispatch.vmware.com:${API_GATEWAY_HTTP_PORT}"}

@test "Create Images for test" {

    run dispatch create base-image base-nodejs6 $DOCKER_REGISTRY/$BASE_IMAGE_NODEJS6 --language nodejs6 --public
    assert_success
    run_with_retry "dispatch get base-image base-nodejs6 --json | jq -r .status" "READY" 4 5

    run dispatch create image nodejs6 base-nodejs6
    assert_success
    run_with_retry "dispatch get image nodejs6 --json | jq -r .status" "READY" 4 5
}

@test "Create Functions for test" {
    run dispatch create function nodejs6 func-nodejs6 ${DISPATCH_ROOT}/examples/nodejs6/hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function func-nodejs6 --json | jq -r .status" "READY" 6 5
}

@test "Test APIs with HTTP(S)" {
    run dispatch create api api-test-http func-nodejs6 -m POST -p /http -a public
    echo_to_log
    assert_success

    run_with_retry "dispatch get api api-test-http --json | jq -r .status" "READY" 6 5

    echo "${API_GATEWAY_HTTPS_HOST}"

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTP_HOST}/http -d '{
            \"name\": \"VMware\",
            \"place\": \"HTTP\"
        }' | jq -r .myField" "Hello, VMware from HTTP" 6 5

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTPS_HOST}/http -k -d '{
            \"name\": \"VMware\",
            \"place\": \"HTTPS\"
        }' | jq -r .myField" "Hello, VMware from HTTPS" 6 5
}

@test "Test APIs with HTTPS ONLY" {
    run dispatch create api api-test-https-only func-nodejs6 -m POST --https-only -p /https-only -a public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test-https-only --json | jq -r .status" "READY" 6 5

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTP_HOST}/https-only -d '{ \
            \"name\": \"VMware\",
            \"place\": \"HTTPS ONLY\"
        }' | jq -r .message" "Please use HTTPS protocol" 6 5

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTPS_HOST}/https-only -k -d '{ \
            \"name\": \"VMware\",
            \"place\": \"HTTPS ONLY\"
        }' | jq -r .myField" "Hello, VMware from HTTPS ONLY" 6 5
}

@test "Test APIs with Kong Plugins" {

    run dispatch create api api-test func-nodejs6 -m GET -m DELETE -m POST -m PUT -p /hello -a public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test --json | jq -r .status" "READY" 6 5

    # "blocking: true" is default header
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k -d '{ \
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' | jq -r .myField" "Hello, VMware from Palo Alto" 6 5

    # with "x-serverless-blocking: false", it will not return an result
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k -H 'x-serverless-blocking: false' -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' | jq -r .status" "CREATING" 6 5

    # PUT with no payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5

    # PUT with non-json payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k \
        -d \"not a json payload\" | jq -r .myField" "Hello, Noone from Nowhere" 6 5

    # GET with parameters
    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/hello?name=vmware\&place=PaloAlto -k | jq -r .myField" "Hello, vmware from PaloAlto" 6 5

    # GET without parameters
    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/hello -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5

}

@test "Test APIs with CORS" {
    run dispatch create api api-test-cors func-nodejs6 -m POST -m PUT -p /cors -a public --cors
    echo_to_log
    assert_success

    run_with_retry "dispatch get api api-test-cors --json | jq -r .status" "READY" 6 5

    # contains "Access-Control-Allow-Origin: *"
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/cors -k -v -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' 2>&1 | grep -c \"Access-Control-Allow-Origin: *\"" 1 1 1
}

@test "Cleanup" {
    cleanup
}