#!/usr/bin/env python2

# script for parsing go test coverage data into a json format

import os
import sys
import datetime as dt
import json

REPO_PREFIX = "github.com/vmware/dispatch/"
REPO_ROOT = os.path.join(os.path.dirname(__file__), "..")
workdir= os.path.join(REPO_ROOT + "/.cover")
reportdir= os.path.join(workdir, "report.out")

output = { "id": dt.datetime.now().isoformat(), "modules": [] }

try:
    with open(reportdir) as input_file:
        for line in input_file.readlines():
            segs = line.rstrip("\n").split("\t")
            if segs and segs[0].startswith("ok"):
                module_name = segs[1][len(REPO_PREFIX):]
                coverage =  float(segs[3].split(" ")[1].rstrip("%"))
                output["modules"].append({"module": module_name, "coverage":coverage})
    print json.dumps(output)
except:
    pass


