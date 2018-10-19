#!/usr/bin/env python
import os

def handle(ctx, payload):
    if "password" not in os.environ:
        return {"message": "I know nothing"}
    else:
        return {"message": "The password is " + os.environ["password"]}