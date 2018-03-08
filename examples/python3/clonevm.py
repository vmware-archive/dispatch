#!/usr/bin/env python
#######################################################################
## Copyright (c) 2017 VMware, Inc. All Rights Reserved.
## SPDX-License-Identifier: Apache-2.0
#######################################################################

"""
Example function to clone a VM from a template in vSphere.

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

cat << EOF > requirements.txt
certifi==2017.11.5
chardet==3.0.4
idna==2.6
pyvmomi==6.5.0.2017.5.post1
requests==2.18.4
six==1.11.0
urllib3==1.22
EOF

dispatch create base-image python3-base vmware/dispatch-python3-base:0.0.1-dev1 --language python3
dispatch create image python-vmomi python3-base --runtime-deps requirements.txt

Create a function:
dispatch create function python-vmomi clonevm examples/python3/clonevm.py --secret vsphere

Execute it:
dispatch exec clonevm --wait --input='{"host": "VSPHERE_URL","name": "TARGET_VM_NAME","template": "SOURCE_TEMPLATE_NAME"}' --secret vsphere

"""
from pyVmomi import vim
from pyVim.connect import SmartConnect, Disconnect

import atexit
import argparse
import getpass
import ssl


def get_obj(content, vimtype, name):
    """
    Return an object by name, if name is None the
    first found object is returned
    """
    obj = None
    container = content.viewManager.CreateContainerView(
        content.rootFolder, vimtype, True)
    for c in container.view:
        if name:
            if c.name == name:
                obj = c
                break
        else:
            obj = c
            break
    return obj


def clone_vm(
        content, template, vm_name, si,
        datacenter_name, vm_folder, host_name,
        resource_pool, power_on):
    """
    Clone a VM from a template/VM, datacenter_name, vm_folder, datastore_name
    cluster_name, resource_pool, and power_on are all optional.
    """

    # if none git the first one
    datacenter = get_obj(content, [vim.Datacenter], datacenter_name)

    if vm_folder:
        destfolder = get_obj(content, [vim.Folder], vm_folder)
    else:
        destfolder = datacenter.vmFolder

    host = get_obj(content, [vim.HostSystem], host_name)
    resource_pool = get_obj(content, [vim.ResourcePool], resource_pool)

    vmconf = vim.vm.ConfigSpec()

    relospec = vim.vm.RelocateSpec()
    relospec.datastore = template.datastore[0]
    relospec.pool = resource_pool
    relospec.host = host

    clonespec = vim.vm.CloneSpec()
    clonespec.location = relospec
    clonespec.powerOn = power_on

    print("cloning VM...")
    print("clone spec: %s" % clonespec)
    task = template.Clone(folder=destfolder, name=vm_name, spec=clonespec)
    return clonespec, task.info.state

def handle(ctx, payload):
    """
    Let this thing fly
    """
    host = payload.get("host")
    port = payload.get("port", 443)
    if host is None:
        raise Exception("Host required")
    secrets = ctx["secrets"]
    if secrets is None:
        raise Exception("Requires vsphere secrets")
    username = secrets["username"]
    password = secrets.get("password", "")

    template_name = payload.get("template")
    name = payload.get("name")
    dc_name = payload.get("datacenterName")
    host_name = payload.get("hostName")
    vm_folder = payload.get("vmFolder")
    resource_pool = payload.get("resourcePool")
    power_on = payload.get("powerOn", False)

    # connect this thing
    context = None
    if hasattr(ssl, '_create_unverified_context'):
        context = ssl._create_unverified_context()
    si = SmartConnect(
        host=host,
        user=username,
        pwd=password,
        port=port,
        sslContext=context)
    try:
        content = si.RetrieveContent()
        template = get_obj(content, [vim.VirtualMachine], template_name)
        state = "unknown"
        clonespec = None
        if template:
            clonespec, state = clone_vm(
                content, template, name, si,
                dc_name, vm_folder,
                host_name, resource_pool,
                power_on)
        else:
            print("template not found")
        return {
            "state": state
        }

    finally:
        Disconnect(si)

