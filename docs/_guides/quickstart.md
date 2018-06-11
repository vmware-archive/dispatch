---
layout: default
---
# Quickstart

Dispatch depends on kubernetes.  To get stared quickly we recommend using minikube, however there are some gotchas
with regards to exactly how to configure minikube.  Please see [Installing Kubernetes via Minikube for Dispatch](minikube.html).  You should now have a working kubernetes cluster.

If you have any issues during installation, please see [Troubleshooting Dispatch](troubleshooting.html).

## Download Dispatch CLI
Get the dispatch command, make it executable, and put it in your path (if you are using the Linux VM from the minikube
instalation, you may skip this step):

#### Get the Latest Release Version
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

## Configure and install Dispatch:

Fetch the IP address of minikube as this will be used the host for dispatch services.
```bash
export DISPATCH_HOST=$(minikube ip)
```

Configure the installation (note: you must substitute the IP address where $DISPATCH_HOST is specified in the below config.yaml):
```bash
$ cat << EOF > config.yaml
apiGateway:
  host: $DISPATCH_HOST
dispatch:
  host: $DISPATCH_HOST
  debug: true
  skipAuth: true
EOF
```

```bash
$ dispatch install --file config.yaml
...
Config file written to: $HOME/.dispatch/config.json
```
Your dispatch config file $HOME/.dispatch/config.json will be generated
and have the following:-
```bash
cat $HOME/.dispatch/config.json
{
    "host": "<DISPATCH_HOST>",
    "port": <port>,
    "organization": "",
    "cookie": "",
    "insecure": true,
    "Json": false
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
$ dispatch create --file seed.yaml --work-dir examples/
$ dispatch get images
   NAME   |                    URL                |  BASEIMAGE   |   STATUS    |         CREATED DATE
------------------------------------------------------------------------------------------------------------------------
  nodejs  | dispatchframework/nodejs-base:0.0.8   | nodejs-base  | READY       | Wed Dec  6 14:28:30 PST 2017
  python3 | dispatchframework/python3-base:0.0.8  | python3-base | INITIALIZED | Wed Dec  6 14:28:30 PST 2017

$ dispatch get functions
    NAME   |  IMAGE  | STATUS |         CREATED DATE
------------------------------------------------------------
  hello-js | nodejs | READY  | Wed Dec  6 14:29:05 PST 2017
  hello-py | python3 | READY  | Wed Dec  6 14:28:52 PST 2017
```

## Execute a function:
```bash
$ dispatch exec hello-py --input '{"name": "Jon", "place": "Winterfell"}' --wait
{
    "blocking": true,
    "executedTime": 1515624222,
    "finishedTime": 1515624222,
    "functionId": "5138d918-e78f-41d6-aece-769addd3eed7",
    "functionName": "hello-py",
    "input": {
        "name": "Jon",
        "place": "Winterfell"
    },
    "logs": null,
    "name": "b5b3c1f5-fa8a-4b38-b7d1-475c44b76114",
    "output": {
        "myField": "Hello, Jon from Winterfell"
    },
    "reason": null,
    "secrets": [],
    "status": "READY"
}
```

## Add an API endpoint:
```bash
$ dispatch create api --https-only --method POST --path /hello post-hello hello-py
```

Find the port for the api-gateway service (we are using the NodePort service
type):

```bash
$ kubectl -n kong get service api-gateway-kongproxy
NAME                    CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
api-gateway-kongproxy   10.107.109.1   <nodes>       80:31788/TCP,443:32611/TCP   19m
```

We are looking at the port associated with https/443 => 32611

```bash
$ curl -k "https://$DISPATCH_HOST:32611/hello" -H "Content-Type: application/json" -d '{"name": "Jon", "place": "winterfell"}'
{"myField":"Hello, Jon from winterfell"}
```

Now go build something!