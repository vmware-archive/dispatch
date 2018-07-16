#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Create duplicate API" {
    run dispatch create api api-test-http func-nodejs -m POST -p /http --auth public
    echo_to_log
    assert_failure
}

@test "Test APIs with HTTP(S)" {
    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTP_HOST}/${DISPATCH_ORGANIZATION}/http -H \"Content-Type: application/json\" -d '{
            \"name\": \"VMware\",
            \"place\": \"HTTP\"
        }' | jq -r .myField" "Hello, VMware from HTTP" 6 5

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/http -k -H \"Content-Type: application/json\" -d '{
            \"name\": \"VMware\",
            \"place\": \"HTTPS\"
        }' | jq -r .myField" "Hello, VMware from HTTPS" 6 5
}

@test "Test APIs with HTTPS ONLY" {
    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTP_HOST}/${DISPATCH_ORGANIZATION}/https-only -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"HTTPS ONLY\"
        }' | jq -r .message" "Please use HTTPS protocol" 6 5

    run_with_retry "curl -s -X POST ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/https-only -k -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"HTTPS ONLY\"
        }' | jq -r .myField" "Hello, VMware from HTTPS ONLY" 6 5
}

@test "Test APIs with Kong Plugins" {
    # "x-dispatch-blocking: true" is default header
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k -H \"Content-Type: application/json\" -d '{ \
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' | jq -r .myField" "Hello, VMware from Palo Alto" 6 5

    # with "x-dispatch-blocking: false", it will not return an result
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k -H \"Content-Type: application/json\" -H 'x-dispatch-blocking: false' -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }'" "" 6 5

    # with "x-dispatch-org: invalid", setting this header should have no effect as it's overwritten by the plugin.
    # if the plugin fails to overwrite this HEADER, it will allow end-users to switch orgs and this test should fail
    # with {"message":"function not found"}.
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k -H \"Content-Type: application/json\" -H 'x-dispatch-org: invalid' -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' | jq -r .myField" "Hello, VMware from Palo Alto" 6 5

    # PUT with no content-type and no payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5

    # PUT with json content-type and non-json payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k \
        -H \"Content-Type: application/json\" -d \"not a json payload\" | jq -r .message" "request body is not json" 6 5

    # PUT with x-www-form-urlencoded content-type and x-www-form-urlencoded payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k \
        -H \"Content-Type: application/x-www-form-urlencoded\" -d \"name=VMware&place=Palo Alto\" | jq -r .myField" "Hello, VMware from Palo Alto" 6 5

    # PUT with non-supported content-type and payload
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k \
        -H \"Content-Type: unsupported-content-type\" -d \"some payload\" | jq -r .message" "request body type is not supported: unsupported-content-type" 6 5

    # GET with parameters
    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello?name=vmware\&place=PaloAlto -k | jq -r .myField" "Hello, vmware from PaloAlto" 6 5

    # GET without parameters
    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/hello -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5

    # Test HTTP Context
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/echo -k | jq -r .context.httpContext.method" "PUT" 6 5
}

@test "Test APIs with CORS" {
    # contains "Access-Control-Allow-Origin: *"
    run_with_retry "curl -s -X PUT ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/cors -k -v -H \"Content-Type: application/json\" -d '{
            \"name\": \"VMware\",
            \"place\": \"Palo Alto\"
        }' 2>&1 | grep -c \"Access-Control-Allow-Origin: *\"" 1 10 5
}

@test "Test API Updates" {
    # update path and https
    run dispatch update --work-dir ${E2E_TEST_ROOT} -f api_update.yaml
    assert_success
    run_with_retry "dispatch get api api-test-update --json | jq -r .status" "READY" 6 20

    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTPS_HOST}/${DISPATCH_ORGANIZATION}/goodbye -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5
}