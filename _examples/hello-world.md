---
layout: default
---

# Introduction

This tutorial will walk you through executing a _Hello World_ function in multiple languages that are supported out-of-box in Dispatch.
Ensure you have followed the [Quickstart](../_front/quickstart.md) guide and have a working Dispatch Virtual Appliance.

# Create Seed Images

Dispatch is bundled with a set of seed images for multiple languages to get you started easily. If you followed the guide, you should
have a set of images created in your Dispatch VM. Execute the following command to check:

```bash
$ dispatch get images
     NAME    |   URL                |    BASEIMAGE    | STATUS | CREATED DATE
-----------------------------------------------------------------------------
  java       | dispatch/f23f029e... | java-base       | READY  | ...
  nodejs     | dispatch/6f04f67d... | nodejs-base     | READY  | ...
  powershell | dispatch/edcbdda8... | powershell-base | READY  | ...
  python3    | dispatch/1937b329... | python3-base    | READY  | ...
``` 
 
If the list is empty, you can create the seed images using the following command:

```bash
$ dispatch create seed-images
Created BaseImage: nodejs-base
Created BaseImage: python3-base
Created BaseImage: powershell-base
Created BaseImage: java-base
Created Image: nodejs
Created Image: python3
Created Image: powershell
Created Image: java
```

**Note**: You need to wait a short while for the images to be in the `READY` state. You can always check the status using the `dispatch get images` command.

# Nodejs

Create a hello.js function:
```bash
$ cat << EOF > hello.js
module.exports = function (context, params) {
    let name = "Noone";
    if (params.name) {
        name = params.name;
    }
    let place = "Nowhere";
    if (params.place) {
        place = params.place;
    }
    return {myField: 'Hello, ' + name + ' from ' + place}
};
EOF
```
```bash
$ dispatch create function hello-world-js --image nodejs ./hello.js
Created function: hello-world-js
```
Wait for the function to become READY:
```bash
$ dispatch get function hello-world-js
       NAME      |  FUNCTIONIMAGE            | STATUS | CREATED DATE
-----------------------------------------------------------------
  hello-world-js | dispatch/func-ead7912d... | READY  | ...
```
Execute the function:
```bash
$ dispatch exec hello-world-js --input '{"name": "Jon", "place": "Winterfell"}' --wait
{
    "blocking": true,
    "executedTime": 1543347252,
    "faasId": "a77e2046-c717-42a6-a2ce-8bc21c10e79e",
    "finishedTime": 1543347252,
    "functionId": "6ccca838-fee8-4377-a473-c372e6c97f4e",
    "functionName": "hello-world-js",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "logs": {
        "stderr": null,
        "stdout": null
    },
    "name": "af60cd07-2365-44f5-aee8-8b766c618d5a",
    "output": {
        "myField": "Hello, Jon from Winterfell"
    },
    "reason": null,
    "secrets": [],
    "services": null,
    "status": "READY",
    "tags": []
}
```

# Python3

Create a hello.py function:
```bash
$ cat << EOF > hello.py
def handle(ctx, payload):
    name = "Noone"
    place = "Nowhere"
    if payload:
        name = payload.get("name", name)
        place = payload.get("place", place)
    return {"myField": "Hello, %s from %s" % (name, place)}
EOF
```
```bash
$ dispatch create function hello-world-py --image python3 ./hello.py
Created function: hello-world-py
```
Wait for the function to become READY:
```bash
$ dispatch get function hello-world-py
      NAME       |  FUNCTIONIMAGE            | STATUS | CREATED DATE
-----------------------------------------------------------------
  hello-world-py | dispatch/func-ead7912d... | READY  | ...
```
Execute the function:
```bash
$ dispatch exec hello-world-py --input '{"name": "Jon", "place": "Winterfell"}' --wait
{
    "blocking": true,
    "executedTime": 1543347345,
    "faasId": "f0f81825-a702-46fe-b4d7-d9825f80f6a5",
    "finishedTime": 1543347345,
    "functionId": "0e5e1c3c-60d9-4a1f-adf0-dfb91d8cc079",
    "functionName": "hello-world-py",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "logs": {
        "stderr": null,
        "stdout": [
            "Serving on http://0.0.0.0:9000"
        ]
    },
    "name": "56aa99a0-2431-4b34-9acd-d4fa3cabbf37",
    "output": {
        "myField": "Hello, Jon from Winterfell"
    },
    "reason": null,
    "secrets": [],
    "services": null,
    "status": "READY",
    "tags": []
}
```


# Powershell
Create a hello.ps1 function:
```bash
$ cat << EOF > hello.ps1
function handle(\$context, \$payload) {

    \$name = \$payload.name
    if (!\$name) {
        \$name = "Noone"
    }
    \$place = \$payload.place
    if (!\$place) {
        \$place = "Nowhere"
    }

    return @{myField="Hello, \$name from \$place"}
}
EOF
```
```bash
$ dispatch create function hello-world-ps --image powershell ./hello.ps1
Created function: hello-world-ps
```
Wait for the function to become READY:
```bash
$ dispatch get function hello-world-ps
       NAME      |  FUNCTIONIMAGE            | STATUS | CREATED DATE
-----------------------------------------------------------------
  hello-world-ps | dispatch/func-ead7912d... | READY  | ...
```
Execute the function:
```bash
$ dispatch exec hello-world-ps --input '{"name": "Jon", "place": "Winterfell"}' --wait
{
    "blocking": true,
    "executedTime": 1543347955,
    "faasId": "b820b463-09c0-42a4-8e0c-54563527fae9",
    "finishedTime": 1543347956,
    "functionId": "f490e771-4701-40f3-8acd-1e807c3a777f",
    "functionName": "hello-world-ps",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "logs": {
        "stderr": null,
        "stdout": null
    },
    "name": "b4924cad-6a4d-4eba-bc61-2d84a76967fb",
    "output": {
        "myField": "Hello, Jon from Winterfell"
    },
    "reason": null,
    "secrets": [],
    "services": null,
    "status": "READY",
    "tags": []
}
```

# Java

Create a Hello.java function:
```bash
$ cat << EOF > Hello.java
package io.dispatchframework.examples;

import java.util.Map;
import java.util.function.BiFunction;

public class Hello implements BiFunction<Map<String, Object>, Map<String, Object>, String> {
    public String apply(Map<String, Object> context, Map<String, Object> payload) {
        final Object name = payload.getOrDefault("name", "Someone");
        final Object place = payload.getOrDefault("place", "Somewhere");

        return String.format("Hello, %s from %s", name, place);
    }
}
EOF
```
```bash
$ dispatch create function hello-world-java --image java ./Hello.java
Created function: hello-world-java
```
Wait for the function to become READY:
```bash
$ dispatch get function hello-world-java
        NAME       |  FUNCTIONIMAGE            | STATUS | CREATED DATE
-----------------------------------------------------------------
  hello-world-java | dispatch/func-ead7912d... | READY  | ...
```
Execute the function:
```bash
$ dispatch exec hello-world-java --input '{"name": "Jon", "place": "Winterfell"}' --wait
{
    "blocking": true,
    "executedTime": 1543348425,
    "faasId": "2c70f588-af34-449f-ab41-b383d086f98f",
    "finishedTime": 1543348425,
    "functionId": "e57b234d-65a4-4c4f-bfa1-7f476b802d7c",
    "functionName": "hello-world-java",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "logs": {
        "stderr": null,
        "stdout": null
    },
    "name": "153aa1cc-5219-41d0-8bec-2e8eb2ead742",
    "output": "Hello, Jon from Winterfell",
    "reason": null,
    "secrets": [],
    "services": null,
    "status": "READY",
    "tags": []
}
```
