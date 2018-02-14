---
layout: default
---

# Custom event drivers in Dispatch

Dispatch is designed to accept events from different sources. To be able to do that, Dispatch introduces a concept of event driver.
Event Driver an image that knows how to ingest events from particular source. See more in [Working with events](working-with-events.md).
For a list of built-in drivers, see [Built-in event drivers](built-in-event-drivers.md).

If you would like to add your own event driver, read on.

## Writing custom event driver

You can write your own event driver, or using existing driver image. In both cases, what you need is a Docker image, which
you will register with Dispatch.

### Example: timer event driver

## Using custom event driver

### Driver registration

Once you have the driver image ready and available in image registry, you can register it with dispatch. 
This example will show how to register the timer event driver described earlier in [Writing custom event driver](#Writing custom event driver).

First, we need to register the new type with dispatch. To do that, run following command:
```
dispatch create event-driver-type timer-driver kars7e/timer-driver
```
The above command registers new driver type `timer-driver`, using `kars7e/timer-driver` image from Docker Hub.

Now, you can create the actual event-driver:

```
dispatch create event-driver timer-driver --name timer-driver1
```

And create a subscription:
