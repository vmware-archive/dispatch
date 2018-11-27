---
layout: default
---
# Installing Dispatch using Dispatch Solo OVA

Dispatch can be installed as an all-in-one OVA. OVA can be deployed in vSphere or Fusion/Workstation and contains a single virtual machine that runs dispatch-server. You can use this OVA to get to know Dispatch.

## Download Dispatch OVA
You can download dispatch OVA from here: [Download Dispatch Solo OVA](http://vmware-dispatch.s3-website-us-west-2.amazonaws.com/ova/dispatch-latest-solo.ova).

### Deploy the OVA
Deploy the OVA to the environment of your choice. Whether it's vSphere or VMware Fusion/Workstation, you will be prompted to provide some configuration options. Configure those that you need, but if you have DHCP enabled, only password field should be required:

![Importing Dispatch OVA. Only password is required.]({{ '/assets/images/fusion.png' | relative_url }}){:width="800px"}

### Start the OVA
Once the OVA started, login with `root` user (only for alpha version) and password you configured during OVA deployment. You can use dispatch right away - the CLI inside the VM is preconfigured and can be used to deploy functions!

![Dispatch VM comes with CLI preinstalled]({{ '/assets/images/console.png' | relative_url }}){:width="800px"}

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

Fetch the IP address of your Dispatch VM that is reachable from your machine, as this will be used the host for dispatch services.  The simplest way of gettin your IP address is simply running `ifconfig eth0` from within the
VM:

![Get IP]({{ '/assets/images/get-IP.png' | relative_url }}){:width="800px"}


```bash
export DISPATCH_HOST=10.64.237.69 # replace with your IP
```

Create dispatch config file $HOME/.dispatch/config.json like this:
```bash
cat << EOF > $HOME/.dispatch/config.json
{
    "current": "solo",
    "contexts": {
        "solo": {
            "host": "${DISPATCH_HOST}",
            "port": 8080,
            "scheme": "http",
            "organization": "dispatch",
            "cookie": "cookie",
            "insecure": true,
            "api-https-port": 443
        }
    }
}
EOF
```

## Hello World

At this point, the environment is up and working.  Let's seed the service
with some images and build our first function:

```bash
$ dispatch create seed-images
Created BaseImage: nodejs-base
Created BaseImage: python3-base
Created BaseImage: powershell-base
Created BaseImage: java-base
Created Image: nodejs
Created Image: python3
Created Image: powershell
Created Image: java
```
```bash
$ dispatch get images
     NAME    |   URL                |    BASEIMAGE    | STATUS | CREATED DATE
-----------------------------------------------------------------------------
  java       | dispatch/f23f029e... | java-base       | READY  | ...
  nodejs     | dispatch/6f04f67d... | nodejs-base     | READY  | ...
  powershell | dispatch/edcbdda8... | powershell-base | READY  | ...
  python3    | dispatch/1937b329... | python3-base    | READY  | ...
```
Create a hello.py function:
```bash
$ cat << EOF > hello.py
def handle(ctx, payload):
    name = "Noone"
    place = "Nowhere"
    if payload:
        name = payload.get("name", name)
        place = payload.get("place", place)
    return {"myField": "Hello, %s from %s" % (name, place)}
EOF
```
```bash
$ dispatch create function hello-world --image python3 ./hello.py
Created function: hello-world
```
Wait for the function to become READY:
```bash
$ dispatch get function
     NAME     |  FUNCTIONIMAGE            | STATUS | CREATED DATE
-----------------------------------------------------------------
  hello-world | dispatch/func-ead7912d... | READY  | ...
```
Execute the function:
```bash
$ dispatch exec hello-world --input '{"name": "Jon", "place": "Winterfell"}' --wait
{
    "blocking": true,
    "executedTime": 1542240146,
    "faasId": "ead7912d-1f18-4577-91ee-de4415ee10d0",
    "finishedTime": 1542240146,
    "functionId": "7f37559d-a182-446f-bb6f-b41fbcff1368",
    "functionName": "hello-world",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "logs": {
        "stderr": null,
        "stdout": null,
    },
    "name": "09bc2e3b-bc7a-49b3-ac23-e11a3ade89a5",
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

Check out the examples section for a full list of examples.

Now go build something!
