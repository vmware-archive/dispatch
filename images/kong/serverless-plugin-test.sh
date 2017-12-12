
#!/bin/bash

set -e
set -x

KONG_HOST=192.168.99.100
KONG_ADMIN_PORT=8002
KONG_PROXY_PORT=8000

KONG_ADMIN_URL=${KONG_HOST}:${KONG_ADMIN_PORT}
KONG_PROXY_URL=${KONG_HOST}:${KONG_PROXY_PORT}

curl -s -X GET http://${KONG_ADMIN_URL}/plugins | jq -r ' .data | .[] | .id'

# clean all plugins
for id in `curl -s -s -X GET http://${KONG_ADMIN_URL}/plugins | jq -r ' .data | .[] | .id'`; do
    curl -s -X DELETE http://${KONG_ADMIN_URL}/plugins/${id}
done

# clean all apis
for id in `curl -s -X GET http://${KONG_ADMIN_URL}/apis | jq -r ' .data | .[] | .id'`; do
    curl -s -X DELETE http://${KONG_ADMIN_URL}/apis/${id}
done

# install api
curl -s -X POST http://${KONG_ADMIN_URL}/apis -d 'name=postman-api' -d 'uris=/postman' -d 'upstream_url=https://postman-echo.com/post'

# install plugin
curl -s -X POST http://${KONG_ADMIN_URL}/plugins  \
    -d 'name=serverless-transformer' \
    -d 'config.enable.output=false' \
    -d 'config.add.header=cookie:cookie'

# first disable output transform, to test if cookie is added
echo 'test cookie header'
test $(curl -s -X GET http://${KONG_PROXY_URL}/postman  | jq -r '.headers.cookie') == "cookie"

# update plugin, now fully functional
for id in `curl -s -X GET http://${KONG_ADMIN_URL}/plugins?name=serverless-transformer | jq -r '.data | .[0] | .id'`; do
    curl -s -X PATCH http://${KONG_ADMIN_URL}/plugins/${id} \
        -d 'name=serverless-transformer' \
        -d 'config.substitute.input=input' \
        -d 'config.substitute.output=json' \
        -d 'config.enable.input=true' \
        -d 'config.enable.output=true' \
        -d 'config.http_method=POST'    \
        -d 'config.add.header=cookie:cookie' \
        -d 'config.header_prefix_for_insertion=x-serverless-' \
        -d 'config.insert_to_body.header=blocking:true'
done

echo 'test GET w/ parameter => "input": {} '
test $(curl -s -X GET http://${KONG_PROXY_URL}/postman | jq '. == { "input": {}, "blocking": true}') == true

echo 'test GET with parameter'
test $(curl -s -X GET http://${KONG_PROXY_URL}/postman?name=VMware\&place=PaloAlto \
    | jq '. == { "input": {"name": "VMware", "place": "PaloAlto"}, "blocking": true}') == true

echo 'test POST w/o json header w/o payload'
test $(curl -s -X POST http://${KONG_PROXY_URL}/postman \
    | jq '. == { "input": {}, "blocking": true}') == true

echo 'test POST w/  json header w/o payload'
test $(curl -s -X POST http://${KONG_PROXY_URL}/postman \
	-H 'content-type: application/json' \
    | jq '. == { "input": {}, "blocking": true}') == true

echo 'test POST w/o json header  w/  payload'
test $(curl -s -X POST http://${KONG_PROXY_URL}/postman \
    -d "hello world" | jq '. == { "input": "hello world", "blocking": true}') == true

echo "test POST w/  json header  w/  payload"
test $(curl -s -X POST http://${KONG_PROXY_URL}/postman \
	-H 'content-type: application/json' \
    -d '{ "name": "VMware", "place": "Palo Alto" }' \
    | jq '. == { "input": {"name": "VMware", "place": "Palo Alto"}, "blocking": true}') == true

echo 'test blocking:false header'
test $(curl -s -X POST http://${KONG_PROXY_URL}/postman \
	-H 'content-type: application/json' \
    -H 'x-serverless-blocking: false' \
    -d '{
        "name": "VMware", "place": "Palo Alto"
    }' | jq '. == { "input": {"name": "VMware", "place": "Palo Alto"}, "blocking": false}') == true

echo "all test successfully passed"