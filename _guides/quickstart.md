---
layout: default
---
# Installing Dispatch using Dispatch Solo OVA

Dispatch can be installed as an all-in-one OVA. OVA can be deployed in vSphere or Fusion/Workstation and contains a single virtual machine that runs dispatch-server. You can use this OVA to get to know Dispatch.

## Download Dispatch OVA
You can download dispatch OVA from [here: Download Dispatch SoloOVA](https://s3-us-west-2.amazonaws.com/vmware-dispatch/dispatch-v0.0.1-solo.ova).

### Deploy the OVA
Deploy the OVA to the environment of your choice. Whether it's vSphere or VMware Fusion/Workstation, you will be prompted to provide some configuration options. Configure those that you need, but if you have DHCP enabled, only password field should be required:

![Importing Dispatch OVA. Only password is required.]({{ '/assets/images/console.png' | relative_url }})

### Start the OVA
Once the OVA started, login with `root` user (only for alpha version) and password you configured during OVA deployment. You can use dispatch right away - the CLI inside the VM is preconfigured and can be used to deploy functions!

![Dispatch VM comes with CLI preinstalled]({{ '/assets/images/fusion.png' | relative_url }})

## Download Dispatch CLI
You may run dispatch CLI locally, and point it to your Dispatch VM.

### Get the Latest Release Version
```bash
export LATEST=$(curl -s https://api.github.com/repos/vmware/dispatch/releases/latest | jq -r .name)
```

>Note: If you don't have [jq](https://stedolan.github.io/jq/) and don't want to install it, you can just manually parse
>the JSON response and grab the version.

#### For MacOS
```bash
$ curl -OL https://github.com/vmware/dispatch/releases/download/$LATEST/dispatch-darwin
$ chmod +x dispatch-darwin
$ mv dispatch-darwin /usr/local/bin/dispatch
```

#### For Linux
```bash
$ curl -OL https://github.com/vmware/dispatch/releases/download/$LATEST/dispatch-linux
$ chmod +x dispatch-linux
$ mv dispatch-linux /usr/local/bin/dispatch
```

### Configure and install Dispatch CLI:

Fetch the IP address of your Dispatch VM that is reachable from your machine, as this will be used the host for dispatch services.
```bash
export DISPATCH_HOST=1.2.3.4 # replace with your IP
```

Create dispatch config file $HOME/.dispatch/config.json like this:
```bash
cat $HOME/.dispatch/config.json
{
    "host": "${DISPATCH_HOST}",
    "port": 8080,
    "organization": "dispatch",
    "insecure": true,
}
```

## Get the Examples

To get the examples, you will need to clone the dispatch repository `https://github.com/vmware/dispatch.git` (if you're
using the Linux VM, the examples are in `~/code/dispatch/examples`):

```bash
$ git clone https://github.com/vmware/dispatch.git
$ cd dispatch
```

At this point, the environment is up and working.  Let's seed the service
with some images and functions.  In order to get the examples, you will need
to clone the repository (if you haven't already):
```bash
$ dispatch create seed-images
$ dispatch get images
     NAME    |                             URL                              |    BASEIMAGE    | STATUS |         CREATED DATE
---------------------------------------------------------------------------------------------------------------------------------
  python3    | 10.98.6.125:5000/9eb7dbeb-0048-433f-8d3c-78737e11c1a0:latest | python3-base    | READY  | Mon Aug 27 15:07:17 PDT 2018
  nodejs     | 10.98.6.125:5000/9a299495-8789-4e7b-ba83-b5e139aaa157:latest | nodejs-base     | READY  | Mon Aug 27 15:07:17 PDT 2018
  powershell | 10.98.6.125:5000/7365e72a-1a52-4819-a18f-6b0551c6d0e4:latest | powershell-base | READY  | Mon Aug 27 15:07:17 PDT 2018
  java       | 10.98.6.125:5000/a1ddfc9b-cad2-42fe-b259-3c9931fe6361:latest | java-base       | READY  | Mon Aug 27 15:07:17 PDT 2018
$ dispatch create --file seed.yaml --work-dir examples/
$ dispatch get functions
    NAME    |                           FUNCTIONIMAGE                           | STATUS |         CREATED DATE
--------------------------------------------------------------------------------------------------------------------
  hello-js  | 10.98.6.125:5000/func-317b2876-51f0-4c56-b225-1ca7dbdfb78a:latest | READY  | Mon Aug 27 15:09:53 PDT 2018
  hello-py  | 10.98.6.125:5000/func-929ef9bc-a8e9-4a53-87aa-ec69e530c0a9:latest | READY  | Mon Aug 27 15:09:53 PDT 2018
  http-py   | 10.98.6.125:5000/func-e9656752-a499-4c10-b41b-73abee56c19a:latest | READY  | Mon Aug 27 15:09:53 PDT 2018
  hello-ps1 | 10.98.6.125:5000/func-4e2d6be0-a329-4d6c-8da3-9978ff97f384:latest | READY  | Mon Aug 27 15:09:53 PDT 2018
```

## Execute a function:
```bash
$ dispatch exec hello-py --input '{"name": "Jon", "place": "Winterfell"}' --wait
{
    "blocking": true,
    "executedTime": 1535760194,
    "faasId": "929ef9bc-a8e9-4a53-87aa-ec69e530c0a9",
    "finishedTime": 1535760194,
    "functionId": "15122478-5cf6-497d-9ed7-6a284a64b647",
    "functionName": "hello-py",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "logs": {
        "stderr": null,
        "stdout": null
    },
    "name": "3ca12f09-5df4-4439-893d-f0e5ca380d57",
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

Now go build something!
