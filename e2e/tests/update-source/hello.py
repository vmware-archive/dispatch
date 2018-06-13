#!/usr/bin/env python
#######################################################################
## Copyright (c) 2017 VMware, Inc. All Rights Reserved.
## SPDX-License-Identifier: Apache-2.0
#######################################################################
"""
Example function "Hello World"

** REQUIREMENTS **

* image
dispatch create base-image python3 dispatchframework/python3-base:0.0.8 --language python3
dispatch create image python3 python3

Create a function:
dispatch create function --image=python3 hello-python examples/python3 --handler=hello.handle

Execute it:
dispatch exec hello-python --wait --input='{"name": "Jon", "place": "Winterfell"}'

"""

def handle(ctx, payload):
    name = "Noone"
    place = "Nowhere"
    if payload:
        name = payload.get("name", name)
        place = payload.get("place", place)
    return {"myField": "Goodbye, %s from %s" % (name, place)}
