#! /bin/bash
set -xe

# 3 * 20 = 60 seconds timeout
MAX_RETRY=20
RETRY_INTERVAL=3

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


vs get base-image
# Using config file: /Users/bjung/.vs.yaml
#   NAME | URL | STATUS | CREATED DATE
# ---------------------------------

vs create base-image photon-nodejs6 vmware/vs-openfaas-nodejs-base:0.0.2-dev1 --language nodejs6 --public
# Created base image: photon-nodejs6
test $(vs get base-image photon-nodejs6 --json | jq -r .status) = INITIALIZED
vs create base-image photon-python3 vmware/vs-openfaas-python3-base:0.0.4-dev1 --language python3 --public
# Created base image: photon-python3
test $(vs get base-image photon-python3 --json | jq -r .status) = INITIALIZED
# Wait for image to be pulled (or attempted)
retry_test "vs get base-image photon-nodejs6 --json | jq -r .status" READY

retry_test "vs get base-image photon-python3 --json | jq -r .status" READY


vs create base-image photon-nodejs6-to-delete vmware/vs-openfaas-nodejs-base:0.0.2-dev1 --language nodejs6 --public
# Created base image: photon-nodejs6-to-delete
test $(vs get base-image --json | jq '. | length') = 3

vs create base-image missing-image missing/image:latest --language nodejs6 --public
# Created base image: missing-image
test $(vs get base-image --json | jq '. | length') = 4
test $(vs get base-image missing-image --json | jq -r .status) = INITIALIZED
# Wait for image to be pulled (or attempted)
retry_test "vs get base-image missing-image --json | jq -r .status" ERROR

vs get base-images
#             NAME           |                                      URL                                           | STATUS |         CREATED DATE
# ------------------------------------------------------------------------------------------------------------------------------------------------
#   missing-image            | missing/image:latest                                                               | ERROR  | Thu Oct 19 14:44:54 PDT 2017
#   photon-nodejs6           | vmware/vs-openfaas-nodejs-base:0.0.2-dev1 | READY  | Tue Oct 17 16:51:27 PDT 2017
#   photon-nodejs6-to-delete | vmware/vs-openfaas-nodejs-base:0.0.2-dev1 | READY  | Tue Oct 17 16:55:30 PDT 2017

vs delete base-image photon-nodejs6-to-delete
# Deleted base image: photon-nodejs6-to-delete
retry_test "vs get base-image --json | jq '. | length'" 3

vs get base-image
#             NAME           |                                      URL                                           | STATUS |         CREATED DATE
# ------------------------------------------------------------------------------------------------------------------------------------------------
#   photon-nodejs6           | vmware/vs-openfaas-nodejs-base:0.0.2-dev1 | READY  | Tue Oct 17 16:51:27 PDT 2017

vs create image base-node photon-nodejs6
# created image: base-node
test $(vs get image base-node --json | jq -r .status) = READY
vs create image base-python3 photon-python3
# created image: base-python3
test $(vs get image base-python3 --json | jq -r .status) = READY

vs get image
#     NAME    |                                      URL                                           |   BASEIMAGE    | STATUS |         CREATED DATE
# -------------------------------------------------------------------------------------------------------------------------------------------------
#   base-node | vmware/vs-openfaas-nodejs-base:0.0.2-dev1 | photon-nodejs6 | READY  | Thu Oct 19 15:26:01 PDT 2017

test -e ../examples/nodejs6/hello.js

vs create function base-node hello-no-schema ../examples/nodejs6/hello.js
# Created function: hello-no-schema
test $(vs get function hello-no-schema --json | jq -r .name) = hello-no-schema

test -e ../examples/python3/hello.py

vs create function base-python3 py3-hello-no-schema ../examples/python3/hello.py
# Created function: py3-hello-no-schema
test $(vs get function py3-hello-no-schema --json | jq -r .name) = py3-hello-no-schema

test -e ../examples/nodejs6/hello.schema.in.json
test -e ../examples/nodejs6/hello.schema.out.json

vs create function base-node hello-in-out-schema ../examples/nodejs6/hello.js --schema-in ../examples/nodejs6/hello.schema.in.json --schema-out ../examples/nodejs6/hello.schema.out.json
# Created function: hello-in-out-schema
test $(vs get function hello-in-out-schema --json | jq -r .name) = hello-in-out-schema
test $(vs get function --json | jq '. | length') = 3

vs get function
#          NAME         |   IMAGE   | STATUS |         CREATED DATE
# ----------------------------------------------------------------------
#   hello-in-out-schema | base-node |        | Thu Oct 19 16:56:01 PDT 2017
#   hello-no-schema     | base-node |        | Thu Oct 19 16:55:51 PDT 2017
retry_test "vs exec hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .myField" "Hello, Jon from Winterfell"
retry_test "vs exec py3-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .myField" "Hello, Jon from Winterfell"
retry_test "vs exec hello-in-out-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .myField" "Hello, Jon from Winterfell"

# TODO (bjung): Errors aren't returned as JSON, so hard to test.

vs create function base-node i-have-a-secret ../examples/nodejs6/i-have-a-secret.js
# Created function: i-have-a-secret
vs create secret open-sesame ../examples/nodejs6/secret.json
# secret file map[password:OpenSesame]
# created secret: open-sesame
retry_test "vs exec i-have-a-secret --secret open-sesame --wait --json | jq -r .message" "The password is OpenSesame"


RUNSCOUNT=$(vs get runs hello-in-out-schema --json | jq '. | length')
vs create subscription test.event hello-in-out-schema
retry_test "vs get subscription test_event_hello-in-out-schema  --json | jq -r .status" "READY"
vs emit test.event --payload='{"name": "Jon", "place": "Winterfell"}'
retry_test "vs get runs hello-in-out-schema --json | jq '. | length'" "$(($RUNSCOUNT+1))"


# Cleanup

for i in $(vs get subscription --json | jq -r .[].name); do
    vs delete subscription $i
done

for i in $(vs get function --json | jq -r .[].name); do
    vs delete function $i
done

for i in $(vs get image --json | jq -r .[].name); do
    vs delete image $i
done

for i in $(vs get base-image --json | jq -r .[].name); do
    vs delete base-image $i
done

vs delete secret open-sesame
