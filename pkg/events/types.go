///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package events

import (
	"context"
	"io"
)

// NO TESTS

// Transport is an abstraction over possible implementation of messaging
type Transport interface {
	// TODO: improve the interface when Kafka support is added

	// Publish publishes the event using the underlying transport
	Publish(ctx context.Context, event *CloudEvent, topic string, tenant string) error

	// Subscribe takes a handler to run on every event received on topic. Returns a cancelable Subscription
	Subscribe(ctx context.Context, topic string, handler Handler) (Subscription, error)
	Close()
}

// Handler is a callback function used to handle received event
type Handler func(context.Context, *CloudEvent)

// Subscription represents an active subscription within Transport. Subscription can be stopped
// by calling Unsubscribe()
type Subscription interface {
	Unsubscribe() error
}

// Validator takes a CloudEvent and validates it. Although CloudEvent struct includes tags following
// go-playground/validator convention, validation schema is up to the implementation.
type Validator interface {
	Validate(event *CloudEvent) error
}

// StreamParser takes io.Reader and returns slice of CloudEvents. It is up to the implementation
// whether incorrect CloudEvent should be omitted from slice, or error should be returned.
// If error is returned, events slice is expected to be nil.
type StreamParser interface {
	Parse(io.Reader) ([]CloudEvent, error)
}
