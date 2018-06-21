---
layout: default
---

# Working with Events in Dispatch

Dispatch is designed to accept events from different sources. To be able to do that, Dispatch introduces a concept of event driver.
An Event Driver is an image that knows how to ingest events from a particular source. For example, if you would like to run your functions
in reaction to events coming from VMware vCenter, you would create an event driver of type `vcenter`. There are built-in types,
you can also register your own type by providing a specially crafted image.

## Implementation

Event drivers are implemented as Kubernetes Pods. When a new event driver is created, a new Kubernetes pod is deployed,
with configuration matching the one provided in the request. The pod is wrapped in the deployment, so that it will be
rescheduled/restarted if needed. See [Registering custom event driver](custom-event-drivers.md) for more details.

## How to use event driver in your environment

### Adding event driver

You can use Dispatch CLI or API to add an event driver (UI support coming soon). In the following example, we're adding
a new vCenter event driver. 

```
dispatch create event-driver vcenter --name vcenter1.corp.local
```

The above will create a new event driver of type `vcenter` and --name `vcenter1.corp.local`. Creating an event-driver
takes a while - monitor its status by running:

```
dispatch get event-driver vcenter1.corp.local
```

When the status changes to `RUNNING`, the event driver is ready to use.

Event driver by itself doesn't do much. it produces events which are ingested by Dispatch, but Dispatch does not know what
to do with them yet. To tell Dispatch to react to events, you need to create a subscription.

## Working with subscriptions
### Adding subscription

Subscription tells Dispatch to match certain events to certain functions, so that when events matching defined conditions arrive
to the system, these functions will be executed. Events that do not match any subscriptions are automatically dropped.

To add a new subscription for the `vm.being.created` event in a vcenter event driver, run:

```
dispatch create subscription --event-type vm.being.created myFunction
```

The above command will print output similar to the following:

```
           NAME          |    EVENT TYPE    | FUNCTION NAME | STATUS |         CREATED DATE
---------------------------------------------------------------------------------------------------------
  complete-cicada-410962 | vm.being.created | myFunction    | READY  | Fri Dec 31 17:18:54 PST -0001
```

The above subscription will cause dispatch to execute function `myFunction` for every event of type `vm.being.created`. 

You can also specify a name for your subscription using `--name` parameter. if you don't, a random, human-readable name will be created.  

### Event driver event types

To find out the list of event types produced by built-in event drivers, see [Built-in Event Drivers](built-in-event-drivers.md).


## Emitting events via CLI or API

Event driver is a powerful concept, but there are cases when you want to emit an event by yourself or through some external application.

### Emitting events via CLI

To emit an event using CLI, you can use the `dispatch emit` subcommand.

In the following example, We emit an event of type `my.event`:

```
dispatch emit my.event --data '{ "example": "payload"}'
```  

Events in Dispatch follow [Cloud Events specification](https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/spec.md).
There are certain attributes in Cloud Events that are mandatory. when emitting them through the CLI, the CLI will pick reasonable defaults.
You can customize those defaults using CLI attributes:

* `--source` - Event source. Defaults to `dispatch`
* `--event-id` - Event ID, should be unique within the scope of producer. Defaults to generated UUIDv4. 

You can also provide following optional details about the emitted event:
* `--event-type-version` - Version of the event, specific to the source. Defaults to empty string.
* `--schema-url` - URL to schema describing event data. Defaults to empty string.
* `--content-type` - Content Type of optional data payload. Defaults to `application/json` when using `--data`, or automatically
detected type when using `--data-from-file`. If specified, always takes precedence over detected type.

You can provide data payload for Event using `--data` parameter. `--data-from-file` will read the content from the specified
file path. Use the `--binary` switch to automatically encode the data using base64. If dispatch CLI detects binary type,
the content will automatically be base64-encoded.

### Emitting events via API

You can also emit an event using Dispatch API directly. Here is an example using cURL:

```bash
curl -X "POST" "https://${DISPATCH_URL}/v1/event/" \
     -H 'Cookie: cookie' \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "event": {
    "source": "dispatch",
    "contentType": "application/json",
    "eventTime": "2018-03-02T23:31:35.818-08:00",
    "eventType": "test.event33",
    "eventID": "b4620ea5-8e9d-42d5-a566-6ad2f7873d63",
    "cloudEventsVersion": "0.1",
  }
}'
```
