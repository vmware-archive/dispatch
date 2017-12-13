#! /bin/bash
set -xe

# 3 * 20 = 60 seconds timeout
MAX_RETRY=20
RETRY_INTERVAL=3

function cleanup {
    for i in $(dispatch get subscription --json | jq -r .[].name); do
        dispatch delete subscription $i
    done

    for i in $(dispatch get function --json | jq -r .[].name); do
        dispatch delete function $i
    done

    for i in $(dispatch get image --json | jq -r .[].name); do
        dispatch delete image $i
    done

    for i in $(dispatch get base-image --json | jq -r .[].name); do
        dispatch delete base-image $i
    done

    dispatch delete secret open-sesame || true
}

function retry_test {
  set +x
  local count=${MAX_RETRY}
  while [ $count -gt 0 ]; do
    echo "Running $1"
    local result="$(eval $1)"
    local expect="$2"
    echo "Running $1 = $result"
    echo "comparing $result == $expect"
    test "$result" = "$expect" && break
    count=$(($count - 1))
    sleep ${RETRY_INTERVAL}
  done

  set -xe
  if [ $count -eq 0 ]; then
    echo "Retry failed [$MAX_RETRY]: $@"
    return 1
  fi
  return 0
}

cleanup

dispatch get base-image
# Using config file: /Users/bjung/.dispatch.yaml
#   NAME | URL | STATUS | CREATED DATE
# ---------------------------------

dispatch create base-image photon-nodejs6 vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1 --language nodejs6 --public
# Created base image: photon-nodejs6
test $(dispatch get base-image photon-nodejs6 --json | jq -r .status) = INITIALIZED

dispatch create base-image photon-python3 vmware/dispatch-openfaas-python-base:0.0.5-dev1 --language python3 --public
# Created base image: photon-python3
test $(dispatch get base-image photon-python3 --json | jq -r .status) = INITIALIZED
# Wait for image to be pulled (or attempted)
retry_test "dispatch get base-image photon-nodejs6 --json | jq -r .status" READY

retry_test "dispatch get base-image photon-python3 --json | jq -r .status" READY


dispatch create base-image photon-nodejs6-to-delete vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1 --language nodejs6 --public
# Created base image: photon-nodejs6-to-delete
test $(dispatch get base-image --json | jq '. | length') = 3

dispatch create base-image missing-image missing/image:latest --language nodejs6 --public
# Created base image: missing-image
test $(dispatch get base-image --json | jq '. | length') = 4
test $(dispatch get base-image missing-image --json | jq -r .status) = INITIALIZED
# Wait for image to be pulled (or attempted)
retry_test "dispatch get base-image missing-image --json | jq -r .status" ERROR

dispatch get base-images
#             NAME           |                                      URL                                           | STATUS |         CREATED DATE
# ------------------------------------------------------------------------------------------------------------------------------------------------
#   missing-image            | missing/image:latest                                                               | ERROR  | Thu Oct 19 14:44:54 PDT 2017
#   photon-nodejs6           | vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1 | READY  | Tue Oct 17 16:51:27 PDT 2017
#   photon-nodejs6-to-delete | vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1 | READY  | Tue Oct 17 16:55:30 PDT 2017

dispatch delete base-image photon-nodejs6-to-delete
# Deleted base image: photon-nodejs6-to-delete
retry_test "dispatch get base-image --json | jq '. | length'" 3

dispatch get base-image
#             NAME           |                                      URL                                           | STATUS |         CREATED DATE
# ------------------------------------------------------------------------------------------------------------------------------------------------
#   photon-nodejs6           | vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1 | READY  | Tue Oct 17 16:51:27 PDT 2017

dispatch create image base-node photon-nodejs6
# created image: base-node
test $(dispatch get image base-node --json | jq -r .status) = READY
dispatch create image base-python3 photon-python3
# created image: base-python3
test $(dispatch get image base-python3 --json | jq -r .status) = READY

dispatch get image
#     NAME    |                                      URL                                           |   BASEIMAGE    | STATUS |         CREATED DATE
# -------------------------------------------------------------------------------------------------------------------------------------------------
#   base-node | vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1 | photon-nodejs6 | READY  | Thu Oct 19 15:26:01 PDT 2017

test -e ../examples/nodejs6/hello.js

dispatch create function base-node hello-no-schema ../examples/nodejs6/hello.js
# Created function: hello-no-schema
test $(dispatch get function hello-no-schema --json | jq -r .name) = hello-no-schema
retry_test "dispatch get function hello-no-schema --json | jq -r .status" "READY"

test -e ../examples/python3/hello.py

dispatch create function base-python3 py-hello-no-schema ../examples/python3/hello.py
# Created function: py-hello-no-schema
test $(dispatch get function py-hello-no-schema --json | jq -r .name) = py-hello-no-schema
retry_test "dispatch get function py-hello-no-schema --json | jq -r .status" "READY"

test -e ../examples/nodejs6/hello.schema.in.json
test -e ../examples/nodejs6/hello.schema.out.json

dispatch create function base-node hello-in-out-schema ../examples/nodejs6/hello.js --schema-in ../examples/nodejs6/hello.schema.in.json --schema-out ../examples/nodejs6/hello.schema.out.json
# Created function: hello-in-out-schema
test $(dispatch get function hello-in-out-schema --json | jq -r .name) = hello-in-out-schema
test $(dispatch get function --json | jq '. | length') = 3

dispatch get function
#          NAME         |   IMAGE   | STATUS |         CREATED DATE
# ----------------------------------------------------------------------
#   hello-in-out-schema | base-node |        | Thu Oct 19 16:56:01 PDT 2017
#   hello-no-schema     | base-node |        | Thu Oct 19 16:55:51 PDT 2017
retry_test "dispatch exec hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .myField" "Hello, Jon from Winterfell"
retry_test "dispatch exec py-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .myField" "Hello, Jon from Winterfell"
retry_test "dispatch exec hello-in-out-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .myField" "Hello, Jon from Winterfell"

# test on function logs
dispatch create function base-node hello-logs ../examples/nodejs6/hello-logs.js
# Created function: hello-logs
retry_test "dispatch get function hello-logs --json | jq -r .status" "READY"

retry_test "dispatch exec hello-logs --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json --all | jq -r '.logs | .[0]'" "returning message name=Jon and place=Winterfell"

# TODO (bjung): Errors aren't returned as JSON, so hard to test.

dispatch create function base-node i-have-a-secret ../examples/nodejs6/i-have-a-secret.js
# Created function: i-have-a-secret
retry_test "dispatch get function i-have-a-secret --json | jq -r .status" "READY"

dispatch create secret open-sesame ../examples/nodejs6/secret.json
# secret file map[password:OpenSesame]
# created secret: open-sesame
retry_test "dispatch exec i-have-a-secret --secret open-sesame --wait --json | jq -r .message" "The password is OpenSesame"


RUNSCOUNT=$(dispatch get runs hello-in-out-schema --json | jq '. | length')
dispatch create subscription test.event hello-in-out-schema
retry_test "dispatch get subscription test_event_hello-in-out-schema  --json | jq -r .status" "READY"
dispatch emit test.event --payload='{"name": "Jon", "place": "Winterfell"}'
retry_test "dispatch get runs hello-in-out-schema --json | jq '. | length'" "$(($RUNSCOUNT+1))"

echo "all test successfully passed"

# Cleanup
cleanup
