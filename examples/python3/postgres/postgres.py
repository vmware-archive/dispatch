#!/usr/bin/env python
#######################################################################
## Copyright (c) 2017 VMware, Inc. All Rights Reserved.
## SPDX-License-Identifier: Apache-2.0
#######################################################################
"""
Example function leveraging postgres service

** REQUIREMENTS **

* image
dispatch create base-image python3 dispatchframework/python3-base:0.0.6 --language python3
dispatch create image python3-pg python3 --runtime-deps requirements.txt

* azure postgres service
dispatch create serviceinstance azure-pg azure-postgresql basic50 --params '
    {
        "location": "westus",
        "resourceGroup": "demo",
        "firewallRules": [
            {
                "startIPAddress": "0.0.0.0",
                "endIPAddress": "255.255.255.255",
                "name": "AllowAll"
            }
        ]
    }'

Below is an example of the context passed to the function:
    {
        "secrets": {},
        "serviceBindings": {
            "azure-pg": {
                "database": "tsfuw5flxw",
                "host": "cc41126d-4bc5-46b6-80c9-74d528779c91.postgres.database.azure.com",
                "password": "*******",
                "port": "5432",
                "sslRequired": "true",
                "tags": "[\"postgresql\"]",
                "uri": "postgresql://luryo9enzo%40cc41126d-4bc5-46b6-80c9-74d528779c91:*******@cc41126d-4bc5-46b6-80c9-74d528779c91.postgres.database.azure.com:5432/tsfuw5flxw?\u0026sslmode=require",
                "username": "luryo9enzo@cc41126d-4bc5-46b6-80c9-74d528779c91"
            }
        }
    }

Create a function:
dispatch create function --image=python3-pg pg-example . --handler=postgres.handle --service azure-pg

Execute it:
dispatch exec pg-example --wait --input '{"num": 1, "data": "hello guest"}'

"""

import sys
import uuid

import psycopg2

CREATE_TEST = """CREATE TABLE IF NOT EXISTS demo (id serial PRIMARY KEY, num integer, data varchar)"""
INSERT_TEST = """INSERT INTO demo (num, data) VALUES (%s, %s)"""
SELECT_TEST = """SELECT * FROM demo"""

conn = None

def cursor(db):
    global conn
    if conn is None:
        try:
            conn = psycopg2.connect(
                host=db["host"],
                port=db["port"],
                database=db["database"],
                user=db["username"],
                password=db["password"],
                sslmode="require")
            cur = conn.cursor()
            cur.execute(CREATE_TEST)
            return cur
        except:
            print("I am unable to initialize database", file=sys.stderr)
            conn = None
            raise
    return conn.cursor()


def handle(ctx, payload):
    try:
        # Hard-coding the name of the database/binding isn't great, need to
        # find a cleaner solution
        db = ctx["serviceBindings"]["azure-pg"]

        cur = cursor(db)
        cur.execute(INSERT_TEST, (payload["num"], payload["data"]))
        cur.execute(SELECT_TEST)
        rows = cur.fetchall()
        result = []
        for row in rows:
            result.append({
                "num": row[1],
                "data": row[2]
            })
            print("num: %s, data: %s" % (row[1], row[2]), file=sys.stderr)
        return result
    except Exception as e:
        print(e, file=sys.stderr)
