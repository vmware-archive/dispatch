---
layout: default
---

# Working with Events in Dispatch

Dispatch is designed to accept events from different sources. To be able to do that, Dispatch introduces a concept of event driver.
Event Driver an image that knows how to ingest events from particular source. For example, if you would like to run your functions
in reaction to events coming from VMware vCenter, you would create an event driver of type `vcenter`. There are few built-in types,
you can also register your own type by providing a specially crafted image.

## Implementation

Event drivers are implemented as Kubernetes Pods. When new event driver is created, a new Kubernetes pod is deployed,
with configuration matching the one provided in the request. Pod is wrapped in the deployment, so that the pod will be
rescheduled/restarted if needed. See [Registering custom event driver](custom-event-drivers.md) for more details.

## How to use event driver in your environment

### Adding event driver

You can use Dispatch CLI or API to add an event driver (UI support coming soon). In following example, we're adding 
new vCenter event driver. 

```
dispatch create event-driver vcenter --name vcenter1.corp.local
```

The above will create new event driver of type `vcenter` and --name `vcenter1.corp.local`. Creating an event-driver
takes a while - monitor its status by running:

```
dispatch get event-driver vcenter1.corp.local
```

When status changes to `RUNNING`, event driver is ready to use.

Event driver by itself doesn't do much. it produces events which are ingested by Dispatch, but Dispatch does not know what
to do with them yet. To tell Dispatch to react to events, you need to create a subscription.

### Event driver event types

To find out the list of event types produced by built-in event drivers, see [Built-in Event Drivers](built-in-event-drivers.md).

## Working with subscriptions
### Adding subscription

Subscription tells Dispatch to match certain events to certain function, so that when events matching defined conditions arrive
to the system, these functions will be executed. Events that do not match any subscriptions are automatically dropped.

To add a new subscription for `vm.being.created` event in vcenter event driver, run:

```
dispatch create subscription --event-type vm.being.created myFunction
```

The above command will print output similar to the following:

```
           NAME          | SOURCE ID | SOURCE TYPE |    EVENT TYPE    | FUNCTION NAME | STATUS |         CREATED DATE
-----------------------------------------------------------------------------------------------------------------------
  swift-pangolin-4235124 | *         | *           | vm.being.created | myFunction    | ACTIVE | Sun Feb 25 12:14:01 PST 2018
```

The above subscription will cause dispatch to execute function `myFunction` for every event of type `vm.being.created`. 
Note the asterisks in `SOURCE ID` and `SOURCE TYPE` columns. These are default values and say "match all source IDs and source Types".
What are source IDs and source Types? If you create event driver of `vcenter` type and name `vcenter1.corp.local`,
`vcenter` is your source type and `vcenter1.corp.local` is your source ID. Currently, the source ID is unique across event types,
so specifying both source ID and source Type does not make much sense. Dispatch will return an error if both of those fields are provided.


Note that you can omit `event-type` as well, and just run:
```
dispatch create subscription myFunction
```

Beware, though: the above will execute `myFunction` for EVERY event that comes to the system. There are also special events in the system,
with source type and source id equal to `dispatch`. Those will not be included in the above catch-all subscription. To subscribe to
dispatch system events, you must explicitly subscribe to `dispatch` source ID.

You can also specify a name for your subscription using `--name` parameter. if you don't, a random, human-readable name will be created.  

### Working with events

### Emitting events via CLI or API

Event driver is a powerful concept, but there are cases when you want to emit an event by your self or through some external application.
Dispatch exposes dedicated API to emit events.

To emit an event using CLI, you can use `dispatch emit` subcommand.

In a following example, We emit an event of type `my.event`:

```
dispatch emit my.event --data '{ "example": "payload"}'
```  

Events in Dispatch follow [Cloud Events specification](https://github.com/cloudevents/spec/blob/460a90d7a69f6257246487e37746797aa2ae919f/spec.md).
There are certain attributes in Cloud Events that are mandatory. when emitting them through CLI, CLI will pick reasonable defaults.
You can customize those defaults using CLI attributes:

* `--event-namespace` - Event namespace. Defaults to `dispatchframework.io` 
* `--source-type` - Event source type. Defaults to `external`
* `--source-id` - Event source ID. Defaults to `external`
* `--event-id` - Event ID, should be unique within the scope of producer. Defaults to generated UUIDv4. 
  NOTE: If source ID is `external`, event-id will ALWAYS be overwritten by Dispatch and set to random UUIDv4. 

You can also provide following optional details about the emitted event:
* `--event-type-version` - Version of the event, specific to the source. Defaults to empty string.
* `--event-time` - Time when event was generated. Defaults to current time.
* `--schema-url` - URL to schema describing event data. Defaults to empty string.
* `--content-type` - Content Type of optional data payload. Defaults to empty string.

You can provide data payload for Event using `--data` parameter. `--data-from-file` will read the content from specified
file path. Use `--binary` switch to automatically encode the data using base64. 

TODO: Add API example.

## System events

TODO