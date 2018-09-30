#!/usr/bin/env python3

import subprocess
import argparse
import pathlib
import json
import os
import sys

def install():
    parser = argparse.ArgumentParser(description='Install dispatch dependencies')
    parser.add_argument("cluster", help="gke cluster name")
    parser.add_argument("--gcloud-key", dest="gcloud_key", required=True, help="gcloud service account key (json format)")
    parser.add_argument("--revision", dest="revision", default="origin/master", help="knative serving revision")
    opts = parser.parse_args()

    with open(opts.gcloud_key) as fh:
        key_contents = json.load(fh)

    home = pathlib.Path.home()

    cmd = [
        '/usr/local/bin/docker',
        'run',
        '-v',
        '%s/.kube:/root/.kube' % home,
        '-v',
        '%s:/root/key.json' % os.path.abspath(opts.gcloud_key),
        '-v',
        '/var/run/docker.sock:/var/run/docker.sock',
        'dispatchframework/knative-installer-gke:0.1',
        key_contents["project_id"],
        opts.cluster,
        '/root/key.json',
        opts.revision]

    ret = subprocess.run(cmd)
    sys.exit(ret.returncode)

if __name__ == "__main__":
    install()
