#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Batch load images" {
    batch_create_images
}

@test "Create node function no schema" {

    run dispatch create function --image=nodejs node-hello-no-schema ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-no-schema -o json | jq -r .status" "READY" 8 2
}

@test "Create a function with duplicate name" {
    run dispatch create function --image=nodejs node-hello-dup ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    echo_to_log
    assert_success

    run dispatch create function --image=nodejs node-hello-dup ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    assert_failure
}

@test "Execute node function no schema" {
    run_with_retry "dispatch exec node-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait -o json | jq -r .output.myField" "Hello, Jon from Winterfell" 10 5
}

@test "Delete node function no schema" {
    run dispatch delete function node-hello-no-schema
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs node-hello-no-schema -o json | jq '. | length'" 0 5 2
}

@test "Create python function no schema" {
    run dispatch create function --image=python3 python-hello-no-schema ${DISPATCH_ROOT}/examples/python3 --handler=hello.handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function python-hello-no-schema -o json | jq -r .status" "READY" 10 2
}

@test "Execute python function no schema" {
    run_with_retry "dispatch exec python-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait -o json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 2
}

@test "Create powershell function no schema" {
    run dispatch create function --image=powershell powershell-hello-no-schema ${DISPATCH_ROOT}/examples/powershell --handler=hello.ps1::handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function powershell-hello-no-schema -o json | jq -r .status" "READY" 20 2
}

@test "Execute powershell function no schema" {
    run_with_retry "dispatch exec powershell-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait -o json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 2
}

@test "Create powershell function with runtime deps" {
    skip "Skipped until #602 is resolved"
    run dispatch create image powershell-with-slack powershell-base --runtime-deps ${DISPATCH_ROOT}/examples/powershell/requirements.psd1
    assert_success
    run_with_retry "dispatch get image powershell-with-slack -o json | jq -r .status" "READY" 10 2

    run dispatch create function --image=powershell-with-slack powershell-slack ${DISPATCH_ROOT}/examples/powershell --handler=test-slack.ps1::handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function powershell-slack -o json | jq -r .status" "READY" 20 2
}

@test "Execute powershell with runtime deps" {
    skip "Skipped until #602 is resolved"
    run_with_retry "dispatch exec powershell-slack --wait -o json | jq -r .output.result" "true" 5 2
}

@test "Create java function no schema" {
    run dispatch create function --image=java java-hello-no-schema ${DISPATCH_ROOT}/examples/java/hello-with-deps --handler=io.dispatchframework.examples.Hello
    echo_to_log
    assert_success

    run_with_retry "dispatch get function java-hello-no-schema -o json | jq -r .status" "READY" 20 2
}

@test "Execute java function no schema" {
    run_with_retry "dispatch exec java-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait -o json | jq -r .output" "Hello, Jon from Winterfell" 5 2
}

@test "Create java function with runtime deps" {
    run dispatch create image java-with-deps java-base --runtime-deps ${DISPATCH_ROOT}/examples/java/hello-with-deps/pom.xml
    assert_success
    run_with_retry "dispatch get image java-with-deps -o json | jq -r .status" "READY" 20 2

    run dispatch create function --image=java-with-deps java-hello-with-deps ${DISPATCH_ROOT}/examples/java/hello-with-deps --handler=io.dispatchframework.examples.HelloWithDeps
    echo_to_log
    assert_success

    run_with_retry "dispatch get function java-hello-with-deps -o json | jq -r .status" "READY" 20 2
}

@test "Execute java with runtime deps" {
    run_with_retry "dispatch exec java-hello-with-deps --wait -o json | jq -r .output" "Hello, Someone from timezone UTC" 5 2
}

@test "Create java function with classes" {
    run dispatch create function --image=java java-hello-with-classes ${DISPATCH_ROOT}/examples/java/hello-with-deps --handler=io.dispatchframework.examples.HelloWithClasses
    echo_to_log
    assert_success

    run_with_retry "dispatch get function java-hello-with-classes -o json | jq -r .status" "READY" 20 2
}

@test "Execute java with classes" {
    run_with_retry "dispatch exec java-hello-with-classes --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait -o json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 2
}

@test "Create java function with spring support" {
    run dispatch create image java-spring java-base --runtime-deps ${DISPATCH_ROOT}/examples/java/spring-pom.xml
    assert_success

    run_with_retry "dispatch get image java-spring -o json | jq -r .status" "READY" 10 2

    run dispatch create function --image=java-spring spring-fn ${DISPATCH_ROOT}/examples/java/hello-with-deps --handler=io.dispatchframework.examples.HelloSpring
    echo_to_log
    assert_success

    run_with_retry "dispatch get function spring-fn -o json | jq -r .status" "READY" 20 2
}

@test "Execute java with spring support" {
    run_with_retry "dispatch exec spring-fn --input='{\"name\":\"Jon\", \"place\":\"Winterfell\"}' --wait -o json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 2
}

@test "Create node function with schema" {
    run dispatch create function --image=nodejs node-hello-with-schema ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js --schema-in ${DISPATCH_ROOT}/examples/nodejs/hello.schema.in.json --schema-out ${DISPATCH_ROOT}/examples/nodejs/hello.schema.out.json
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-with-schema -o json | jq -r .status" "READY" 6 2
}

@test "Execute node function with schema" {
    run_with_retry "dispatch exec node-hello-with-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait -o json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 2
}

@test "Validate schema errors" {
    skip "schema validation only returns null"
}

@test "Execute node function with input schema error" {
    run_with_retry "dispatch exec node-hello-with-schema --wait -o json | jq -r .error.type" "InputError" 5 2
}

@test "Create python function with runtime deps" {
    run dispatch create function --image=python3 http ${DISPATCH_ROOT}/examples/python3 --handler=http_func.handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function http -o json | jq -r .status" "READY" 6 2
}

@test "Execute python function with runtime deps" {
    run_with_retry "dispatch exec http --wait -o json | jq -r .output.status" "200" 5 2
}

@test "Create python function with invalid handler" {
    run dispatch create function --image=python3 invalid ${DISPATCH_ROOT}/examples/python3 --handler=non_existent_file.handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function invalid -o json | jq -r .status" "ERROR" 5 2
    run_with_retry "dispatch get function invalid -o json | jq -r '.reason | length'" "1" 5 2
}

@test "Create python function with logging" {
    src_dir=$(mktemp -d)
    cat << EOF > ${src_dir}/logging_test.py
import sys

def handle(ctx, payload):
    print("this goes to stdout", flush=True)
    print("this goes to stderr", file=sys.stderr, flush=True)
EOF

    run dispatch create function --image=python3 logger ${src_dir} --handler=logging_test.handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function logger -o json | jq -r .status" "READY" 6 2
}

@test "Execute python function with logging" {
    run_with_retry "dispatch exec logger --wait -o json | jq -r \".logs.stderr | .[0]\"" "this goes to stderr" 5 2
    run_with_retry "dispatch exec logger --wait -o json | jq -r \".logs.stdout | .[0]\"" "this goes to stdout" 5 2
}

@test "Execute function with a literal value payload" {
    run_with_retry "dispatch exec http --input='42' --wait -o json | jq -r .output.status" "200" 5 2
}

@test "Update function python" {
    run dispatch update --work-dir ${BATS_TEST_DIRNAME} -f function_update.yaml
    assert_success

    run_with_retry "dispatch get function python-hello-no-schema -o json | jq -r .status" "READY" 6 2

    run_with_retry "dispatch exec python-hello-no-schema --wait -o json | jq -r .output.myField" "Goodbye, Noone from Nowhere" 6 2
}

@test "Delete functions" {
    delete_entities function
}

@test "Cleanup" {
    cleanup
}
