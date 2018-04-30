# Tutorial: Using Dispatch with vCenter

## Description
This tutorial will guide you to write a serverless function that can be executed in response to events from your vCenter server.

## Overview
Serverless functions in Dispatch can be automatically executed based on events from a specific source or manually executed by the user. For Dispatch to monitor events from a source, an Event Driver must exist for that source. By default, Dispatch provides certain Event Drivers e.g for a VMware vCenter Server. In this tutorial, you will learn how to create an Event Driver instance for your vCenter Server and write a serverless function that will post messages to a Slack channel whenever a VM is created in vCenter.

Following is a summary of the serverless workflow :-

1. User creates a VM in vCenter.

1. Dispatch's vCenter Event Driver detects the VmBeingCreatedEvent event.

1. Dispatch's Event Manager triggers the subscribed serverless function.

1. The function posts a slack message on a specific channel with the details of the created VM.

## Prerequisites
* Dispatch framework must be installed

* VMware vCenter Server

* Slack Account with privileges to create an incoming webhook

## Steps

### Step 1: Create an instance of the vCenter Event Driver in Dispatch
```
$ dispatch create event-driver <name> vcenter --set vcenterurl='<user>:<password>@<vcenter_host>:443'
```
where
<dl>
<dt>name</dt>
<dd>The name of the event driver instance e.g my-vcenter</dd>
<dt>user</dt>
<dd>The vCenter user e.g administrator@vsphere.local</dd>
<dt>password</dt>
<dd>The vCenter password</dd>
<dt>vcenter_host</dt>
<dd>The vCenter Server Host IP address or hostname</dd>
</dl>

Validate the status of the event driver:
```
$ dispatch get event-driver
  NAME |  TYPE   | STATUS | SECRETS |                              CONFIG
--------------------------------------------------------------------------------------------------------
  demo | vcenter | READY  |         | vcenterurl=administrator@vsphere.local:Admin!23@10.160.220.2:443
--------------------------------------------------------------------------------------------------------
```

### Step 2: Configure an incoming webhook in your Slack Account
Set up an incoming webhook integration in your slack workspace by following the instructions in this <a href="https://api.slack.com/incoming-webhooks" target="_blank">page</a>. Once you have setup the integration, make a note of the webhook URL.

Create a json file `secret.json` with the incoming webhook URL e.g.
```
{
    "slack_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
}
```

Create a dispatch secret to store your Slack webhook URL.
```
$ dispatch create secret slack secret.json
$ dispatch get secret
Note: secret values are hidden, please use --all flag to get them

                   ID                  | NAME  | CONTENT
-------------------------------------------------------
  41383c4e-0ae2-11e8-85b1-fa6bb5c6aa04 | slack | <hidden>
```

### Step 4: Create the default images

```
cat << EOF > images.yaml
kind: base-image
name: nodejs6-base
dockerUrl: dispatchframework/nodejs-base:0.0.3
language: nodejs6
public: true
tags:
  - key: role
    value: test
---
kind: base-image
name: python3-base
dockerUrl: dispatchframework/python3-base:0.0.3
language: python3
public: true
tags:
  - key: role
    value: test
---
kind: image
name: nodejs6
baseImageName: nodejs6-base
tags:
  - key: role
    value: test
---
kind: image
name: python3
baseImageName: python3-base
tags:
  - key: role
    value: test
EOF
```

Now batch load them into dispatch:

```
$ dispatch create -f images.yaml
```

Wait till all images are READY:

```
$ dispatch get images
   NAME   |                              URL                              |  BASEIMAGE   | STATUS |         CREATED DATE
-----------------------------------------------------------------------------------------------------------------------------
  python3 | 10.99.39.122:5000/5b2d437b-1d87-4d56-83c5-e9e2b915aef1:latest | python3-base | READY  | Fri Dec 31 18:43:13 PST -0001
  nodejs6 | 10.99.39.122:5000/f8f43fb0-46aa-481b-b2f8-386d9042b2ca:latest | nodejs6-base | READY  | Fri Dec 31 18:43:13 PST -0001
```

### Step 3: Create a sample serverless function to post messages to Slack


Download and create the sample serverless function.
```
curl -LO https://raw.githubusercontent.com/vmware/dispatch/master/examples/vcenter/slack_vm_being_deployed.js

$ dispatch create function nodejs6 slack-post-message slack_vm_being_deployed.js
```

Verify the function is READY:

```
$ dispatch get function
         NAME        |  IMAGE  | STATUS |         CREATED DATE
--------------------------------------------------------------------
  slack-post-message | nodejs6 | READY  | Fri Dec 31 18:49:29 PST -0001
```

### Step 4: Subscribe to the vCenter Event
Subscribe to the event **vm.being.created** that is published by the vCenter Event Driver and specify the name of the function **slack-post-message** that must be executed when the event occurs.
```
$ dispatch create subscription vm.being.created slack-post-message --secret slack
```

Verify the subscription is READY:

```
$ dispatch get subscription
                  NAME                 |       TOPIC       | SUBSCRIBER TYPE |  SUBSCRIBER NAME   | STATUS |         CREATED DATE
-------------------------------------------------------------------------------------------------------------------------------------
  vm_being_created_slack-post-message  | vm.being.created  | function        | slack-post-message | READY  | Fri Dec 31 18:51:03 PST -0001
```

### Step 5: Create a VM in your vCenter Server
Create a VM in your vCenter Server to watch the serverless function get executed and post a message to your Slack channel.
