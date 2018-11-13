#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

: ${API_GATEWAY_HTTPS_HOST:="https://api.dev.dispatch.vmware.com:${API_GATEWAY_HTTPS_PORT}"}
: ${API_GATEWAY_HTTP_HOST:="http://api.dev.dispatch.vmware.com:${API_GATEWAY_HTTP_PORT}"}

@test "Create Images for test" {

    run dispatch create base-image base-nodejs $DOCKER_REGISTRY/$BASE_IMAGE_NODEJS6 --language nodejs
    assert_success
    run_with_retry "dispatch get base-image base-nodejs -o json | jq -r .status" "READY"

    run dispatch create image nodejs base-nodejs
    assert_success
    run_with_retry "dispatch get image nodejs -o json | jq -r .status" "READY"
}

@test "Create Functions for test" {
    run dispatch create function --image=nodejs func-nodejs ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function func-nodejs -o json | jq -r .status" "READY"

    run dispatch create function --image=nodejs node-echo-back ${DISPATCH_ROOT}/examples/nodejs --handler=./debug.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-echo-back -o json | jq -r .status" "READY"
}

@test "Test APIs with HTTP(S)" {
    run dispatch create api api-test-http func-nodejs -m POST -p /http --auth public
    echo_to_log
    assert_success

    run_with_retry "dispatch get api api-test-http -o json | jq -r .status" "READY"

    echo "${API_GATEWAY_HTTPS_HOST}"

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTP_HOST}/${DISPATCH_ORGANIZATION}/http -H \"Content-Type: application/json\" -d '{
            \"name\": \"VMware\",
            \"place\": \"HTTP\"
        }' | jq -r .myField" "Hello, VMware from HTTP"

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/http -k -H \"Content-Type: application/json\" -d '{
            \"name\": \"VMware\",
            \"place\": \"HTTPS\"
        }' | jq -r .myField" "Hello, VMware from HTTPS"
}

@test "Test APIs with HTTPS ONLY" {
    run dispatch create api api-test-https-only func-nodejs -m POST --https-only -p /https-only --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test-https-only -o json | jq -r .status" "READY"

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTP_HOST}/${DISPATCH_ORGANIZATION}/https-only -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"HTTPS ONLY\"
        }' | jq -r .message" "Please use HTTPS protocol"

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/https-only -k -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"HTTPS ONLY\"
        }' | jq -r .myField" "Hello, VMware from HTTPS ONLY"
}

@test "Test APIs with Kong Plugins" {

    run dispatch create api api-test func-nodejs -m GET -m DELETE -m POST -m PUT -p /hello --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test -o json | jq -r .status" "READY"

    run dispatch create api api-echo node-echo-back -m GET -m DELETE -m POST -m PUT -p /echo --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-echo -o json | jq -r .status" "READY"

    # "x-dispatch-blocking: true" is default header
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' | jq -r .myField" "Hello, VMware from Palo Alto"

    # with "x-dispatch-blocking: false", it will not return an result
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k -H \"Content-Type: application/json\" -H 'x-dispatch-blocking: false' -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }'" ""

    # with "x-dispatch-org: invalid", setting this header should have no effect as it's overwritten by the plugin.
    # if the plugin fails to overwrite this HEADER, it will allow end-users to switch orgs and this test should fail
    # with {"message":"function not found"}.
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k -H \"Content-Type: application/json\" -H 'x-dispatch-org: invalid' -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' | jq -r .myField" "Hello, VMware from Palo Alto"

    # PUT with no content-type and no payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k | jq -r .myField" "Hello, Noone from Nowhere"

    # PUT with json content-type and non-json payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k \
        -H \"Content-Type: application/json\" -d \"not a json payload\" | jq -r .message" "request body is not json"

    # PUT with x-www-form-urlencoded content-type and x-www-form-urlencoded payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k \
        -H \"Content-Type: application/x-www-form-urlencoded\" -d \"name=VMware&place=Palo Alto\" | jq -r .myField" "Hello, VMware from Palo Alto"

    # PUT with non-supported content-type and payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k \
        -H \"Content-Type: unsupported-content-type\" -d \"some payload\" | jq -r .message" "request body type is not supported: unsupported-content-type"

    # GET with parameters
    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello?name=vmware\&place=PaloAlto -k | jq -r .myField" "Hello, vmware from PaloAlto"

    # GET without parameters
    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k | jq -r .myField" "Hello, Noone from Nowhere"

    # Test HTTP Context
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/echo -k | jq -r .context.httpContext.method" "PUT"
}

@test "Test APIs with CORS" {
    run dispatch create api api-test-cors func-nodejs -m POST -m PUT -p /cors --auth public --cors
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test-cors -o json | jq -r .status" "READY"

    # contains "Access-Control-Allow-Origin: *"
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/cors -k -v -H \"Content-Type: application/json\" -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' 2>&1 | grep -c \"Access-Control-Allow-Origin: *\"" 1
}

@test "Test API Updates" {
    run dispatch create api api-test-update func-nodejs -m GET -p /hello --auth public
    assert_success
    run_with_retry "dispatch get api api-test-update -o json | jq -r .status" "READY"

    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTP_HOST}/${DISPATCH_ORGANIZATION}/hello -k | jq -r .myField" "Hello, Noone from Nowhere"

    # update path and https
    run dispatch update --work-dir ${BATS_TEST_DIRNAME} -f api_update.yaml
    assert_success
    run_with_retry "dispatch get api api-test-update -o json | jq -r .status" "READY"

    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/goodbye -k | jq -r .myField" "Hello, Noone from Nowhere"
}

@test "Cleanup" {
    cleanup
}
