#!/usr/bin/env python
#######################################################################
## Copyright (c) 2017 VMware, Inc. All Rights Reserved.
## SPDX-License-Identifier: Apache-2.0
#######################################################################

"""
Sample function to list templates in vSphere inventory.

** REQUIREMENTS **

* secret

cat << EOF > vsphere.json
{
    "password": "VSPHERE_PASSWORD",
    "username": "VSPHERE_USERNAME"
}
EOF
dispatch create secret vsphere vsphere.json

* image
dispatch create base-image python-vmomi kars7e/dispatch-openfaas-python3-vmomi:0.0.2-dev1 --language python3 --public
dispatch create image python-vmomi python-vmomi

Create a function:
dispatch create function python-vmomi gettemplates examples/python3/gettemplates.py --secret vsphere

Execute it:
dispatch exec gettemplates --wait --input='{"host": "VSPHERE_URL"}' --secret vsphere

"""

from __future__ import print_function

from pyVim.connect import SmartConnect, Disconnect
from pyVmomi import vim

import ssl

def handle(ctx, payload):
    host = payload.get("host")
    port = payload.get("port", 443)
    if host is None:
        raise Exception("Host required")
    secrets = ctx["secrets"]
    if secrets is None:
        raise Exception("Requires vsphere secrets")
    username = secrets["username"]
    password = secrets.get("password", "")
    context = None
    if hasattr(ssl, '_create_unverified_context'):
        context = ssl._create_unverified_context()
    si = SmartConnect(host=host,
                        user=username,
                        pwd=password,
                        port=port,
                        sslContext=context)
    if not si:
        raise Exception(
            "Could not connect to the specified host using specified "
            "username and password")
    try:
        content = si.RetrieveContent()
        container = content.viewManager.CreateContainerView(
            content.rootFolder, [vim.VirtualMachine], True)
        templates = []
        for c in container.view:
            if c.config.template:
                templates.append(
                    {
                        "name": c.name,
                        "guest": c.config.guestFullName,
                        "annotation": c.config.annotation
                    }
                )
        return templates
    except Exception as e:
        return {"exception": "%s" % e}
    finally:
        Disconnect(si)
