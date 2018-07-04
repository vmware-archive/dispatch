---
title: Functions
---

# Functions

Functions are code artifacts which are layered on top of [images](images.md).  They should be consise, containing only
business logic.  As much as possible, function dependencies should be captured within images.

## Adding a Function

Adding a function is simple in dispatch.  Generally speaking all you need to do is point to a directory containing
source code and specify a path to the handler which represents the function entry point.  The path to the handler is
language specific.  For language specifics, see the base image repositories:

* [python3-base-image](https://github.com/dispatchframework/python3-base-image#creating-functions)
* [nodejs-base-image](https://github.com/dispatchframework/nodejs-base-image#creating-functions)
* [java-base-image](https://github.com/dispatchframework/java-base-image#creating-functions)
* [powershell-base-image](https://github.com/dispatchframework/powershell-base-image#creating-functions)

The examples below will assume python3 as the function language.

### Create from a Source Directory

The following is the canonical way to create functions, however see the section below for usage optimizations
for the simple (and common) cases.

1. Create a source directory:

    ```
    $ mkdir hello
    $ cd hello
    ```

2. Add files:

    ```
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

2. Create Dispatch function:

    The format is `dispatch create function [function name] [source dir] --image [image name] --handler [entry point]`

    ```
    $ dispatch create function hello . --image python3 --handler hello.handle
    Created function: hello
    ```

    > Note: The handler/entry point is language specifice.  For instance, for a node function it is a unix path to the
    file (e.g. `hello.js`) relative to the source directory

### Create from a Single File

Most functions are single files therefore packaging a directory seems unnecessary.  Additionally, if we use standard
names for the handler function (e.g. `handle`), we can make some inferrences which make building functions simpler:

```
$ dispatch create function hello hello.py --image python3
Created function: hello
```

Here we simply point directly to the function file and the default handler (`handle`) is used.

### Creating and Updating via YAML

When iterating on a function, it's easier to use the generic `dispatch create -f [YAML]` and `dispatch update -f [YAML]`
commands.  Just define your function[s] in a YAML document:

```
$ cat function.yaml
kind: Function
name: function-1
sourcePath: 'function-1.py'
image: python3
schema: {}
secrets:
  - example-secret
services:
  - example-service
tags:
  - key: role
    value: example
---
kind: Function
name: function-2
sourcePath: 'function-2.py'
image: python3
schema: {}
secrets:
  - example-secret
services:
  - example-service
tags:
  - key: role
    value: example
```

Creating and updating is now as simple as:

```
$ dispatch create -f function.yaml
Created function: function-1
Created function: function-2
$ dispatch update -f function.yaml
Updated function: function-1
Updated function: function-2
```

Even deletion works:

```
$ dispatch delete -f function.yaml
Deleted function: function-1
Deleted function: function-2
```

## Executing a Function

To simply execute a function from the CLI use the `exec` command:

```
$ dispatch exec --wait hello --input '{"name": "Jon", "place": "Winterfell"}'
{
    "blocking": true,
    "executedTime": 1530657442,
    "faasId": "694d66b9-a9f0-452c-832f-40300c627991",
    "finishedTime": 1530657444,
    "functionId": "bb64d31e-d586-4869-8317-860098b7d751",
    "functionName": "hello",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "logs": {
        "stderr": null,
        "stdout": null
    },
    "name": "6e5acc98-ac22-4b44-8741-eb4d5ea11fb2",
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

The `--input` is presented to the function as the `payload`.  The `--wait` flag was added to simulate a synchronous
execution.  Function execution is actually asynchronous.  To demonstrate leave off the `--wait` flag:

```
$ dispatch exec hello --input '{"name": "Jon", "place": "Winterfell"}'
{
    "executedTime": 1530657613,
    "faasId": "694d66b9-a9f0-452c-832f-40300c627991",
    "finishedTime": -62135596800,
    "functionId": "bb64d31e-d586-4869-8317-860098b7d751",
    "functionName": "hello",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "name": "a4fc2ee8-e007-4131-b334-ea6723b1338b",
    "reason": null,
    "secrets": [],
    "services": null,
    "status": "INITIALIZED",
    "tags": []
}
```

In both cases we get a `run` object back from Dispatch.  However note that the status is `INITIALIZED` rather than
`READY`.  We can poll on the `name` to get the result when it is ready:

```
$ dispatch get run hello a4fc2ee8-e007-4131-b334-ea6723b1338b --json
{
    "executedTime": 1530657613,
    "faasId": "694d66b9-a9f0-452c-832f-40300c627991",
    "finishedTime": 1530657615,
    "functionId": "bb64d31e-d586-4869-8317-860098b7d751",
    "functionName": "hello",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "logs": {
        "stderr": null,
        "stdout": null
    },
    "name": "a4fc2ee8-e007-4131-b334-ea6723b1338b",
    "output": {
        "myField": "Hello, Jon from Winterfell"
    },
    "reason": null,
    "secrets": null,
    "services": null,
    "status": "READY",
    "tags": []
}
```