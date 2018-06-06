#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

: ${API_GATEWAY_HTTPS_HOST:="https://api.dev.dispatch.vmware.com:${API_GATEWAY_HTTPS_PORT}"}
: ${API_GATEWAY_HTTP_HOST:="http://api.dev.dispatch.vmware.com:${API_GATEWAY_HTTP_PORT}"}

@test "Create Images for test" {

    run dispatch create base-image base-nodejs $DOCKER_REGISTRY/$BASE_IMAGE_NODEJS6 --language nodejs
    assert_success
    run_with_retry "dispatch get base-image base-nodejs --json | jq -r .status" "READY" 4 5

    run dispatch create image nodejs base-nodejs
    assert_success
    run_with_retry "dispatch get image nodejs --json | jq -r .status" "READY" 8 5
}

@test "Create Functions for test" {
    run dispatch create function --image=nodejs func-nodejs ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function func-nodejs --json | jq -r .status" "READY" 10 5

    run dispatch create function --image=nodejs node-echo-back ${DISPATCH_ROOT}/examples/nodejs --handler=./debug.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-echo-back --json | jq -r .status" "READY" 10 5
}

@test "Test APIs with HTTP(S)" {
    run dispatch create api api-test-http func-nodejs -m POST -p /http --auth public
    echo_to_log
    assert_success

    run_with_retry "dispatch get api api-test-http --json | jq -r .status" "READY" 10 5

    echo "${API_GATEWAY_HTTPS_HOST}"

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTP_HOST}/http -H \"Content-Type: application/json\" -d '{
            \"name\": \"VMware\",
            \"place\": \"HTTP\"
        }' | jq -r .myField" "Hello, VMware from HTTP" 6 5

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTPS_HOST}/http -k -H \"Content-Type: application/json\" -d '{
            \"name\": \"VMware\",
            \"place\": \"HTTPS\"
        }' | jq -r .myField" "Hello, VMware from HTTPS" 6 5
}

@test "Test APIs with HTTPS ONLY" {
    run dispatch create api api-test-https-only func-nodejs -m POST --https-only -p /https-only --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test-https-only --json | jq -r .status" "READY" 6 5

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTP_HOST}/https-only -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"HTTPS ONLY\"
        }' | jq -r .message" "Please use HTTPS protocol" 6 5

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTPS_HOST}/https-only -k -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"HTTPS ONLY\"
        }' | jq -r .myField" "Hello, VMware from HTTPS ONLY" 6 5
}

@test "Test APIs with Kong Plugins" {

    run dispatch create api api-test func-nodejs -m GET -m DELETE -m POST -m PUT -p /hello --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test --json | jq -r .status" "READY" 6 5

    run dispatch create api api-echo node-echo-back -m GET -m DELETE -m POST -m PUT -p /echo --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-echo --json | jq -r .status" "READY" 6 5

    # "blocking: true" is default header
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' | jq -r .myField" "Hello, VMware from Palo Alto" 6 5

    # with "x-serverless-blocking: false", it will not return an result
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k -H \"Content-Type: application/json\" -H 'x-dispatch-blocking: false' -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }'" "" 6 5

    # with "x-dispatch-org: invalid", setting this header should have no effect as it's overwritten by the plugin.
    # if the plugin fails to overwrite this HEADER, it will allow end-users to switch orgs and this test should fail
    # with {"message":"function not found"}.
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k -H \"Content-Type: application/json\" -H 'x-dispatch-org: invalid' -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' | jq -r .myField" "Hello, VMware from Palo Alto" 6 5

    # PUT with no content-type and no payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5

    # PUT with json content-type and non-json payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k \
        -H \"Content-Type: application/json\" -d \"not a json payload\" | jq -r .message" "request body is not json" 6 5

    # PUT with x-www-form-urlencoded content-type and x-www-form-urlencoded payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k \
        -H \"Content-Type: application/x-www-form-urlencoded\" -d \"name=VMware&place=Palo Alto\" | jq -r .myField" "Hello, VMware from Palo Alto" 6 5

    # PUT with non-supported content-type and payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/hello -k \
        -H \"Content-Type: unsupported-content-type\" -d \"some payload\" | jq -r .message" "request body type is not supported: unsupported-content-type" 6 5

    # GET with parameters
    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/hello?name=vmware\&place=PaloAlto -k | jq -r .myField" "Hello, vmware from PaloAlto" 6 5

    # GET without parameters
    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/hello -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5

    # Test HTTP Context
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/echo -k | jq -r .context.httpContext.method" "PUT" 6 5
}

@test "Test APIs with CORS" {
    run dispatch create api api-test-cors func-nodejs -m POST -m PUT -p /cors --auth public --cors
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test-cors --json | jq -r .status" "READY" 10 5

    # contains "Access-Control-Allow-Origin: *"
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/cors -k -v -H \"Content-Type: application/json\" -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' 2>&1 | grep -c \"Access-Control-Allow-Origin: *\"" 1 10 5
}

@test "Test API Updates" {
    run dispatch create api api-test-update func-nodejs -m GET -p /hello --auth public
    assert_success
    run_with_retry "dispatch get api api-test-update --json | jq -r .status" "READY" 6 5

    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTP_HOST}/hello -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5

    # update path and https
    run dispatch update --work-dir ${BATS_TEST_DIRNAME} -f api_update.yaml
    assert_success
    run_with_retry "dispatch get api api-test-update --json | jq -r .status" "READY" 6 20

    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/goodbye -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5
}

@test "Cleanup" {
    cleanup
}
