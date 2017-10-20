#! /bin/bash

vs get base-image
# Using config file: /Users/bjung/.vs.yaml
#   NAME | URL | STATUS | CREATED DATE
# ---------------------------------

vs create base-image photon-nodejs6 serverless-docker-local.artifactory.eng.vmware.com/photon-func-deps-node:7.7.4 --language nodejs6 --public
# Created base image: photon-nodejs6
test $(vs get base-image photon-nodejs6 --json | jq -r .status) = INITIALIZING
# Wait for image to be pulled (or attempted)
sleep 5
test $(vs get base-image photon-nodejs6 --json | jq -r .status) = READY


vs create base-image photon-nodejs6-to-delete serverless-docker-local.artifactory.eng.vmware.com/photon-func-deps-node:7.7.4 --language nodejs6 --public
# Created base image: photon-nodejs6-to-delete
test $(vs get base-image --json | jq '. | length') = 2

vs create base-image missing-image missing/image:latest --language nodejs6 --public
# Created base image: missing-image
test $(vs get base-image --json | jq '. | length') = 3
test $(vs get base-image missing-image --json | jq -r .status) = INITIALIZING
# Wait for image to be pulled (or attempted)
sleep 5
test $(vs get base-image missing-image --json | jq -r .status) = ERROR

vs get base-images
#             NAME           |                                      URL                                       | STATUS |         CREATED DATE
# ------------------------------------------------------------------------------------------------------------------------------------------------
#   missing-image            | missing/image:latest                                                           | ERROR  | Thu Oct 19 14:44:54 PDT 2017
#   photon-nodejs6           | serverless-docker-local.artifactory.eng.vmware.com/photon-func-deps-node:7.7.4 | READY  | Tue Oct 17 16:51:27 PDT 2017
#   photon-nodejs6-to-delete | serverless-docker-local.artifactory.eng.vmware.com/photon-func-deps-node:7.7.4 | READY  | Tue Oct 17 16:55:30 PDT 2017

vs delete base-image photon-nodejs6-to-delete
# Deleted base image: photon-nodejs6-to-delete
sleep 5
test $(vs get base-image --json | jq '. | length') = 2

vs get base-image
#             NAME           |                                      URL                                       | STATUS |         CREATED DATE
# ------------------------------------------------------------------------------------------------------------------------------------------------
#   photon-nodejs6           | serverless-docker-local.artifactory.eng.vmware.com/photon-func-deps-node:7.7.4 | READY  | Tue Oct 17 16:51:27 PDT 2017

vs create image base-node photon-nodejs6
# created image: base-node
test $(vs get image base-node --json | jq -r .status) = READY

vs get image
#     NAME    |                                      URL                                       |   BASEIMAGE    | STATUS |         CREATED DATE
# -------------------------------------------------------------------------------------------------------------------------------------------------
#   base-node | serverless-docker-local.artifactory.eng.vmware.com/photon-func-deps-node:7.7.4 | photon-nodejs6 | READY  | Thu Oct 19 15:26:01 PDT 2017

test -e examples/nodejs6/hello.js

vs create function base-node hello-no-schema examples/nodejs6/hello.js
# Created function: hello-no-schema
test $(vs get function hello-no-schema --json | jq -r .name) = hello-no-schema

test -e examples/nodejs6/hello.schema.in.json
test -e examples/nodejs6/hello.schema.out.json

vs create function base-node hello-in-out-schema examples/nodejs6/hello.js --schema-in examples/nodejs6/hello.schema.in.json --schema-out examples/nodejs6/hello.schema.out.json
# Created function: hello-in-out-schema
test $(vs get function hello-in-out-schema --json | jq -r .name) = hello-in-out-schema
test $(vs get function --json | jq '. | length') = 2

vs get function
#          NAME         |   IMAGE   | STATUS |         CREATED DATE
# ----------------------------------------------------------------------
#   hello-in-out-schema | base-node |        | Thu Oct 19 16:56:01 PDT 2017
#   hello-no-schema     | base-node |        | Thu Oct 19 16:55:51 PDT 2017

test "$(vs exec hello-no-schema --input='{"name": "Jon", "place": "Winterfell"}' --wait --json | jq -r .myField)" = "Hello, Jon from Winterfell"

test "$(vs exec hello-in-out-schema --input='{"name": "Jon", "place": "Winterfell"}' --wait --json | jq -r .myField)" = "Hello, Jon from Winterfell"

# TODO (bjung): Errors aren't returned as JSON, so hard to test.

vs create function base-node i-have-a-secret examples/nodejs6/i-have-a-secret.js
# Created function: i-have-a-secret
vs create secret open-sesame examples/nodejs6/secret.json
# secret file map[password:OpenSesame]
# created secret: open-sesame

test "$(vs exec i-have-a-secret --secrets '["open-sesame"]' --wait --json | jq -r .message)" = "The password is OpenSesame"

# Cleanup
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