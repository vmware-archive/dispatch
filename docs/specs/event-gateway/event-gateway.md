# Event gateway

A service which handles events in Dispatch.

## Problem statement

Functions are commonly associated with events and are executed in response to them. To effectively manage their
relation, we need a way to handle events lifecycle (CRUD) as well as subscriptions and triggering.

### User stories (examples)

1. As a *developer*, I want to create an *event* defintion identifed by *name/topic*, and optional predefined *payload
   schema*.
2. As a *developer*, I want to trigger an *event* by providing its *name/topic* and JSON-formatted *payload*, which will
   in turn cause execution of one or more functions *subscribed* to that event.
3. As a *developer*, I want to create a *subscription* between *event* and *function*. When *event* is triggered,
   *function(s)* is/are executed with event payload.

## Proposed Solution

An Event Manager component, responsible for event and subscription lifecycle.

Event manager:
* Exposes API for CRUD operations on events.
* Exposes API for CRUD operations on subscriptions.
* Exposes API to trigger events.
* Executes functions subscribed to particular events.

## Requirements

* API which is used as central storage
* Function executor, which Event Manager will connect to to run functions subscribed to event.

## Design

### Subscription object schema
* `subscriptionId` - `string` - subscription ID
* `event` - `string` - event name which we subscribe to.
* `functionId` - ID of a function to be executed when event is triggered.
* `metadata` - custom metadata.

### Event object schema
* `event` - `string` - event name which we subscribe to.
* `functionId` - ID of a function to be executed when event is triggered.
* `receivedAt` - `datetime` - Set by Event Manager, date and time when event was triggered most recently.
* `data` - type depends on `dataType` - payload to be sent to function.
* `dataType` - `string` - mime type of payload.
* `schema` - optional schema for validation (TODO: Do we want to include it here?)
* `metadata` - custom metadata.

## Milestones

TBD

## Open Issues

* Single event can trigger multiple functions. Should we allow to customize execution strategy?
