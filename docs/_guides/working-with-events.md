---
layout: default
---

# Working with Events in Dispatch

Dispatch is designed to accept events from different sources. To be able to do that, Dispatch introduces an event driver
concept. An Event Driver is an image that knows how to ingest events from a particular source. For example, if you would
like to run your functions in reaction to events coming from VMware vCenter, you would register a driver that knows how
to communicate to vCenter and translate the vCenter events.  Then you create an event driver instance of that type.

## Available Event Drivers

The following table lists out all of the currently supported pre-build event drivers.  Adding additional event drivers
is easy, see [Custom Event Drivers](custom-event-drivers.md) for information.  The GitHub link should provide additional
detail per driver type.

| Driver Type | Image | GitHub | Usage |
| ----------- | ----- | ------ | ----- |
| vcenter | dispatchframework/dispatch-events-vcenter | [dispatchframework/dispatch-events-vcenter](https://github.com/dispatchframework/dispatch-events-vcenter) | [Usage]({{ /documentation/guides/working-with-events | relative_url }}) |
| azure-eventgrid | dispatchframework/dispatch-events-eventgrid | [dispatchframework/dispatch-events-eventgrid](https://github.com/dispatchframework/dispatch-events-eventgrid) | [Usage](https://github.com/dispatchframework/dispatch-events-eventgrid#installation) |
| aws | dispatchframework/dispatch-events-aws | [dispatchframework/dispatch-events-aws](https://github.com/dispatchframework/dispatch-events-aws) | [Usage](https://github.com/dispatchframework/dispatch-events-aws#dispatch-event-driver-for-aws)
| cron | dispatchframework/cron-driver | [dispatchframework/dispatch-events-cron](https://github.com/dispatchframework/dispatch-events-cron) | [Usage](https://github.com/dispatchframework/dispatch-events-cron#create-the-event-driver)
| cloudevents | dispatchframework/dispatch-events-cloudevent | [dispatchframework/dispatch-events-cloudevent](https://github.com/dispatchframework/dispatch-events-cloudevent) | [Usage](https://github.com/dispatchframework/dispatch-events-cloudevent#installation) |


## Implementation

Event drivers are implemented as Kubernetes Pods. When a new event driver is created, a new Kubernetes pod is deployed,
with configuration matching the one provided in the request. The pod is wrapped in the deployment, so that it will be
rescheduled/restarted if needed. See [Registering an Event Driver Type](#registering-an-event-driver-type) for more details.

## Event Driver Usage

The examples below show both the YAML representation of the Dispatch resource which can be used with the CLI (`dispatch
create -f <path to yaml>`) and the pure CLI command.

### Registering an Event Driver Type

Before creating an event driver instance, event driver types must be registered with Dispatch.  Registering an event
driver type is simple.  Dispatch needs to know the name of the driver type, where the driver container image is located
and whether or not to expose an endpoint for the driver.

```yaml
kind: DriverType
image: dispatchframework/dispatch-events-vcenter
expose: false
name: vcenter
tags:
  - key: role
    value: example
```
```
dispatch create eventdrivertype vcenter dispatchframework/dispatch-events-vcenter
```

### Adding an Event Driver Instance

Once the event driver type has been registered, event driver instances can be created.  Unlike the event driver types, instances often take arguements.  To discover the arguments
appropriate to the driver type, see the documentation for the type.  For instance the vcenter
event driver takes vcenterurl as an argument which it uses to connect to the remote vcenter instance.

Event driver arguments can be passed one of two ways:

1. Directly via config or `--set` (do not use this method for passing sensitive data i.e. secrets)

   ```
   kind: Driver
   name: vcenter1.example.com
   type: vcenter
   config:
     - key: vcenterurl
       value: user:passwd@host:port
   tags:
     - key: role
       value: example
   ```
   ```
   dispatch create eventdriver vcenter1.example.com --name vcenter1.example.com --set    vcenterurl=user:passwd@host:port
   ```

2. Indirectly using secrets

   Create a secret with the arguments specfied as keys and values

   ```
   kind: Secret
   name: vcenter
   secrets:
     vcenterurl: user:passwd@host:port
   tags:
     - key: role
       value: example
   ```
   ```
   cat << EOF > vcenter.json
   {
     "vcenterurl": "user:passwd@host:port"
   }
   EOF
   dispatch create secret vcenter ./vcenter.json
   ```

   Then reference that secret when creating the event driver instance.

   ```
   kind: Driver
   name: vcenter1.example.com
   type: vcenter
   secrets:
     - vcenter
   tags:
     - key: role
       value: example
   ```
   ```
   dispatch create eventdriver vcenter --name vcenter1.example.com --secret vcenter
   ```

The above will create a new event driver of type `vcenter` and --name `vcenter1.example.com`. Creating an event-driver
may take some time - monitor its status by running:

```
dispatch get eventdriver vcenter1.example.com
```

When the status changes to `READY`, the event driver is ready to use.

Event driver by itself doesn't do much. it produces events which are ingested by Dispatch, but Dispatch does not know
what to do with them yet. To tell Dispatch to react to events, you need to create a subscription.

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

To find out the list of event types produced by event drivers, see the documentation for the event driver type.


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

> NOTE: The above example assume no authentication.  In a production environment, service account credentials must be
> passed as a JWT bearer token to an Authorization HTTP header.
