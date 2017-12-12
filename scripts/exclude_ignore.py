#!/usr/bin/env python2
import os
import subprocess
import sys

REPO_PREFIX = "github.com/vmware/dispatch/"
REPO_ROOT = os.path.join(os.path.dirname(__file__), "..")

m = {}

for line in sys.stdin:
    line = line.strip("\n")
    splitted = line.split(":")

    if len(splitted) == 2 and splitted[0].startswith(REPO_PREFIX):
        f = splitted[0][len(REPO_PREFIX):]
        try:
            tracked = m[f]
        except KeyError:
            with open(os.devnull, 'w') as devnull:
                p = subprocess.Popen(
                        "git ls-files %s --error-unmatch" % os.path.join(REPO_ROOT, f),
                        shell=True,
                        stdout=devnull,
                        stderr=devnull)
                ret = p.wait()

                if ret == 0:
                    m[f] = True
                else:
                    m[f] = False
                tracked = m[f]

        if tracked:
            print line

    else:
        print line
