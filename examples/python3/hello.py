#!/usr/bin/env python
#######################################################################
## Copyright (c) 2017 VMware, Inc. All Rights Reserved.
## SPDX-License-Identifier: Apache-2.0
#######################################################################
"""
Example function "Hello World"

** REQUIREMENTS **

* image
dispatch create base-image python3 vmware/dispatch-openfaas-python-base:0.0.5-dev1 --language python3 --public
dispatch create image python3 python3

Create a function:
dispatch create function python3 hello-python examples/python3/hello.py

Execute it:
dispatch exec hello-python --wait --input='{"name": "Jon", "place": "Winterfell"}'

"""

def handle(ctx, payload):
    name = payload.get("name", "Noone")  
    place = payload.get("place", "Nowhere")
    return {"myField": "Hello, %s from %s" % (name, place)}
