#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Batch load images" {
    batch_create_images images.yaml
}

@test "Create node function no schema" {

    run dispatch create function nodejs6 node-hello-no-schema ${DISPATCH_ROOT}/examples/nodejs6/hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-no-schema --json | jq -r .status" "READY" 8 5
}

@test "Create a function with duplicate name" {
    run dispatch create function nodejs6 node-hello-dup ${DISPATCH_ROOT}/examples/nodejs6/hello.js
    echo_to_log
    assert_success

    run dispatch create function nodejs6 node-hello-dup ${DISPATCH_ROOT}/examples/nodejs6/hello.js
    assert_failure
}

@test "Execute node function no schema" {
    run_with_retry "dispatch exec node-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 10 5
}

@test "Delete node function no schema" {
    run dispatch delete function node-hello-no-schema
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs node-hello-no-schema --json | jq '. | length'" 0 1 0
}

@test "Create python function no schema" {
    run dispatch create function python3 python-hello-no-schema ${DISPATCH_ROOT}/examples/python3/hello.py
    echo_to_log
    assert_success

    run_with_retry "dispatch get function python-hello-no-schema --json | jq -r .status" "READY" 10 5
}

@test "Execute python function no schema" {
    run_with_retry "dispatch exec python-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Create powershell function no schema" {
    run dispatch create function powershell powershell-hello-no-schema ${DISPATCH_ROOT}/examples/powershell/hello.ps1
    echo_to_log
    assert_success

    run_with_retry "dispatch get function powershell-hello-no-schema --json | jq -r .status" "READY" 8 5
}

@test "Execute powershell function no schema" {
    run_with_retry "dispatch exec powershell-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Create powershell function with runtime deps" {
    run dispatch create image powershell-with-slack powershell-base --runtime-deps ${DISPATCH_ROOT}/examples/powershell/requirements.psd1
    assert_success
    run_with_retry "dispatch get image powershell-with-slack --json | jq -r .status" "READY" 10 5

    run dispatch create function powershell-with-slack powershell-slack ${DISPATCH_ROOT}/examples/powershell/test-slack.ps1
    echo_to_log
    assert_success

    run_with_retry "dispatch get function powershell-slack --json | jq -r .status" "READY" 10 5
}

@test "Execute powershell with runtime deps" {
    run_with_retry "dispatch exec powershell-slack --wait --json | jq -r .output.result" "true" 5 5
}

@test "Create java function no schema" {
    run dispatch create function java8 java-hello-no-schema ${DISPATCH_ROOT}/examples/java8/Hello.java
    echo_to_log
    assert_success

    run_with_retry "dispatch get function java-hello-no-schema --json | jq -r .status" "READY" 20 5
}

@test "Execute java function no schema" {
    run_with_retry "dispatch exec java-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output" "Hello, Jon from Winterfell" 5 5
}

@test "Create java function with runtime deps" {
    run dispatch create image java8-with-deps java8-base --runtime-deps ${DISPATCH_ROOT}/examples/java8/pom.xml
    assert_success
    run_with_retry "dispatch get image java8-with-deps --json | jq -r .status" "READY" 20 5

    run dispatch create function java8-with-deps java-hello-with-deps ${DISPATCH_ROOT}/examples/java8/HelloWithDeps.java
    echo_to_log
    assert_success

    run_with_retry "dispatch get function java-hello-with-deps --json | jq -r .status" "READY" 10 5
}

@test "Execute java with runtime deps" {
    run_with_retry "dispatch exec java-hello-with-deps --wait --json | jq -r .output" "Hello, Someone from timezone UTC" 5 5
}

@test "Create java function with classes" {
    run dispatch create function java8 java-hello-with-classes ${DISPATCH_ROOT}/examples/java8/HelloWithClasses.java
    echo_to_log
    assert_success

    run_with_retry "dispatch get function java-hello-with-classes --json | jq -r .status" "READY" 20 5
}

@test "Execute java with classes" {
    run_with_retry "dispatch exec java-hello-with-classes --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Create java function with spring support" {
    run dispatch create image java8-spring java8-base --runtime-deps ${DISPATCH_ROOT}/examples/java8/spring-pom.xml
    assert_success

    run_with_retry "dispatch get image java8-spring --json | jq -r .status" "READY" 10 5

    run dispatch create function java8-spring spring-fn ${DISPATCH_ROOT}/examples/java8/HelloSpring.java
    echo_to_log
    assert_success

    run_with_retry "dispatch get function spring-fn --json | jq -r .status" "READY" 20 5
}

@test "Execute java with spring support" {
    run_with_retry "dispatch exec spring-fn --input='{\"name\":\"Jon\", \"place\":\"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Create node function with schema" {
    run dispatch create function nodejs6 node-hello-with-schema ${DISPATCH_ROOT}/examples/nodejs6/hello.js --schema-in ${DISPATCH_ROOT}/examples/nodejs6/hello.schema.in.json --schema-out ${DISPATCH_ROOT}/examples/nodejs6/hello.schema.out.json
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-with-schema --json | jq -r .status" "READY" 6 5
}

@test "Execute node function with schema" {
    run_with_retry "dispatch exec node-hello-with-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Validate schema errors" {
    skip "schema validation only returns null"
}

@test "Create python function with runtime deps" {
    run dispatch create function python3 http ${DISPATCH_ROOT}/examples/python3/http.py
    echo_to_log
    assert_success

    run_with_retry "dispatch get function http --json | jq -r .status" "READY" 6 5
}

@test "Execute python function with runtime deps" {
    run_with_retry "dispatch exec http --wait --json | jq -r .output.status" "200" 5 5
}

@test "Create python function with logging" {
    src_file=$(mktemp).py
cat << EOF > ${src_file}
import sys

def handle(ctx, payload):
    print("this goes to stdout")
    print("this goes to stderr", file=sys.stderr)
EOF

    run dispatch create function python3 logger ${src_file}
    echo_to_log
    assert_success

    run_with_retry "dispatch get function logger --json | jq -r .status" "READY" 6 5
}

@test "Execute python function with logging" {
    run_with_retry "dispatch exec logger --wait --json | jq -r \".logs.stderr | .[0]\"" "this goes to stderr" 5 5
    run_with_retry "dispatch exec logger --wait --json | jq -r \".logs.stdout | .[0]\"" "this goes to stdout" 5 5
}

@test "Execute function with a literal value payload" {
    run_with_retry "dispatch exec http --input='42' --wait --json | jq -r .output.status" "200" 5 5
}

@test "Update function python" {
    skip_if_faas riff

    run dispatch update --work-dir ${BATS_TEST_DIRNAME} -f function_update.yaml
    assert_success
    run_with_retry "dispatch get function python-hello-no-schema --json | jq -r .status" "READY" 6 5

    run_with_retry "dispatch exec python-hello-no-schema --wait --json | jq -r .output.myField" "Goodbye, Noone from Nowhere" 6 5
}

@test "Delete functions" {
    delete_entities function
}

@test "Cleanup" {
    cleanup
}
