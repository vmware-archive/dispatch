---
layout: default
---

# Creating Event Drivers for Dispatch

Dispatch is designed to accept events from different sources. To be able to do that, Dispatch introduces a concept of
event driver. An Event Driver is an image that knows how to ingest events from a particular source. For a list of
built-in drivers as well as usage, see [Working with Events](working-with-events.md).

If you would like to add your own event driver, read on.

## Writing custom event driver

Creating event drivers is easy.  As stated previously an event driver is simply a container which can ingest events
from a remote location, translate that event into a cloudevent format and write that event over http (or stdout).  The
event driver is deployed with a side-car container which contains a webserver listening on port 8080, so events are
simply sent to `http://localhost:8080` at which point the side-car forwards the event onto the event bus.

### Credentials and arguments

Often times an event driver will need credentials or other arguments.  These are simply represented as string flags
which are passed through from Dispatch.  For insance if your driver requries a password, the driver should take an
arguement `--password` and the value is passed to the driver through a Dispatch secret which includes a key/value pair
for `"password"`.  Currently Dispatch only supports long form flags prefixed with double dashes (`--arg` not `-a` or
`-arg`).

### Ingesting Events

Dispatch supports both push and pull for ingesting events.  A pull based event driver is pretty straight forward as it
generally polls some remote source for new events.  Unlike functions, event drivers are long running containers which
man maintain some state (such as last timestamps).  However, they are not persistent, so if an event driver crashes,
state will be lost.

Push based event drivers expose an HTTP endpoint for posting event data to from a remote source.  The event driver
should implment an HTTP server listening on port 80 (configurable ports are currently not supported).  The event driver
then translates the HTTP payload into an event or events and sends it on just like all other event drivers. See the
[Driver registration](#driver-registration) section for information about how to expose the driver types.

### Example: ticker event driver

You can find an example custom event driver written in go in `examples/event-drivers/go/ticker-driver`. We will
progressively add more examples of event drivers written in other languages.

## Using custom event driver

### Driver registration

Once you have the driver image ready and available in an image registry, you can register it with dispatch.
This example will show how to register the timer event driver described earlier in [Writing custom event driver](#Writing custom event driver).

First, we need to register the new type with dispatch. To do that, run the following command:
```
dispatch create eventdrivertype ticker vmware/dispatch-ticker-driver:v0.1.0
```
The above command registers a new driver type `ticker` using `vmware/dispatch-ticker-driver:v0.1.0` image from Docker Hub.

Now, you can create the actual event-driver:

```
dispatch create eventdriver ticker --name my-ticker --set seconds=5
```

And create a subscription (assuming `myFunction` already exists):

```
dispatch create subscription --event-type ticker.tick myFunction
```
